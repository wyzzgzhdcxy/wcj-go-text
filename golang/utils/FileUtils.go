package utils

import (
	"fmt"

	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

type WFile struct {
	Path           string
	Name           string      // base name of the file
	Size           string      // length in bytes for regular files; system-dependent for others
	IntSize        int64       // length in bytes for regular files; system-dependent for others
	Mode           os.FileMode // file mode bits
	ModTime        string      // modification time
	IsDir          bool        // abbreviation for Mode().IsDir()
	Sys            any         // underlying data source (can return nil)
	TargetSize     string      // length in bytes for regular files; system-dependent for others
	TargetModeTime string      // length in bytes for regular files; system-dependent for others
}

// ListFiles 列出当前目录下的文件
func ListFiles(dir string) []string {
	var fileList []string
	// 读取目录
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println("Error reading directory:"+dir, err)
		return nil
	}
	// 列出文件
	for _, file := range files {
		// 过滤出文件（非目录）
		if !file.IsDir() {
			fileList = append(fileList, file.Name())
		}
	}
	return fileList
}

// DeleteSubDir 删除子目录
func DeleteSubDir(dir string) {
	// 读取目录
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println("Error reading directory:"+dir, err)
		return
	}
	for _, file := range files {
		if file.IsDir() {
			err := os.RemoveAll(dir + "/" + file.Name())
			log.Printf("删除目录失败:"+dir+"/"+file.Name()+",%v", err)
		}
	}
}

func ListDir(dir string) []WFile {
	var files []WFile
	if exist, _ := core.PathExists(dir); !exist {
		return nil
	}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		timeStr := info.ModTime().Format("2006-01-02 15:04:05")

		files = append(files, WFile{Path: path, Name: info.Name(), IntSize: info.Size(), Size: core.FormatSize(info.Size()),
			Mode: info.Mode().Perm(), ModTime: timeStr, IsDir: info.IsDir(), Sys: info.Sys()})
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}

func ListDirReturnAllDir(dir string) []string {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}

type CompareResult struct {
	Path      string
	WFile     WFile
	CyType    int
	SrcMd5    string
	TargetMd5 string
}
type FileSyncRes struct {
	SrcCount       int
	SrcSize        int64
	SrcSizeStr     string
	DstCount       int
	DstSize        int64
	DstSizeStr     string
	MigrateSizeStr string
	MigrateCount   int
	MigrateSize    int64
	CompareCost    string
	MigrateCost    string
	Result         []CompareResult
}

type FileSyncRequest struct {
	CyData bool
	Task   []FileSyncRequestTask
}

type FileSyncRequestTask struct {
	Src    string `json:"src"`
	Dst    string `json:"dst"`
	cyData bool
}

func CompareDir(src string, dst string) FileSyncRes {
	var result []CompareResult
	files := ListDir(src)
	dstFiles := ListDir(dst)
	dstFilesMap := make(map[string]WFile)
	for _, item := range dstFiles {
		dstFilesMap[item.Path] = item
	}
	for _, file := range files {
		if !file.IsDir {
			result = genCyResult(src, dst, file, dstFilesMap, result)
		}
	}

	var srcSize int64 = 0
	var dstSize int64 = 0
	for _, file := range files {
		srcSize = srcSize + file.IntSize
	}
	for _, file := range dstFiles {
		dstSize = dstSize + file.IntSize
	}
	return FileSyncRes{Result: result, SrcCount: len(files),
		DstCount: len(dstFiles), SrcSize: srcSize, DstSize: dstSize}
}

func genCyResult(src string, dst string, file WFile, dstFilesMap map[string]WFile, result []CompareResult) []CompareResult {
	tarFilePath := strings.Replace(file.Path, src, dst, -1)
	exist := core.MapExistKey(&dstFilesMap, tarFilePath)
	cyType := 0
	if exist {
		file.TargetModeTime = dstFilesMap[tarFilePath].ModTime
		file.TargetSize = dstFilesMap[tarFilePath].Size
		if file.Size != file.TargetSize || strings.Compare(file.ModTime, file.TargetModeTime) > 1 {
			//数据差异
			cyType = 1
		}
	} else {
		//源存在,目标不存在
		cyType = 2
	}
	if !(cyType == 0) {
		result = append(result, CompareResult{Path: file.Path, WFile: file, CyType: cyType})
	}
	return result
}

func CompareDirByStr(fileSyncRequestTask []FileSyncRequestTask) FileSyncRes {
	var result []CompareResult

	var multiRes = make([]FileSyncRes, len(fileSyncRequestTask))
	var wg sync.WaitGroup
	for index, task := range fileSyncRequestTask {
		wg.Add(1)
		go func(inx int, mSrc string, mDst string) {
			defer wg.Done()
			multiRes[inx] = CompareDir(mSrc, mDst)
		}(index, task.Src, task.Dst)
	}
	wg.Wait()
	srcCount := 0
	dstCount := 0
	var srcSize int64 = 0
	var dstSize int64 = 0
	for _, r := range multiRes {
		result = append(result, r.Result...)
		srcCount = srcCount + r.SrcCount
		dstCount = dstCount + r.DstCount
		srcSize = srcSize + r.SrcSize
		dstSize = dstSize + r.DstSize
	}

	res := FileSyncRes{Result: result, SrcCount: srcCount, DstCount: dstCount,
		SrcSize: srcSize, DstSize: dstSize}
	var migrateSize int64
	if result != nil {
		res.MigrateCount = len(result)
		for _, r := range result {
			migrateSize = migrateSize + r.WFile.IntSize
		}
		res.MigrateSize = migrateSize
		res.MigrateSizeStr = core.FormatSize(res.MigrateSize)
	}
	return res
}

func StartFileSync(fileSyncRequestTask []FileSyncRequestTask, callback func([]byte)) string {
	var wg sync.WaitGroup
	for _, task := range fileSyncRequestTask {
		wg.Add(1)
		taskTmp := task
		go func() {
			defer wg.Done()
			fileSyncTask(taskTmp, callback)
		}()
	}
	wg.Wait()
	return "true"
}

type FileSyncWebSocketRes struct {
	Path   string
	Status int
}

func fileSyncTask(task FileSyncRequestTask, callback func([]byte)) (string, bool) {
	for _, dir := range ListDirReturnAllDir(task.Src) {
		tarFilePath := strings.Replace(dir, task.Src, task.Dst, -1)
		exist, _ := core.PathExists(tarFilePath)
		if !exist {
			err := os.MkdirAll(tarFilePath, os.ModePerm)
			if err != nil {
				return "", true
			}
		}
	}
	result := CompareDir(task.Src, task.Dst)
	for _, compareResult := range result.Result {
		tarFilePath := strings.Replace(compareResult.Path, task.Src, task.Dst, -1)
		if !compareResult.WFile.IsDir {
			res := FileSyncWebSocketRes{
				Path:   compareResult.Path,
				Status: 3, //3-同步中 0-同步完成
			}
			callback([]byte(core.ToJsonString(res)))
			size, err := core.CopyFile(compareResult.Path, tarFilePath)
			if err == nil {
				res.Status = 0 //3-同步中 0-同步完成
				callback(core.ToByteArray(res))
				log.Printf("复制成功,路径：%s,目标路径:%s,大小：%d", compareResult.Path, tarFilePath, size)
			} else {
				res.Status = 4 //4-同步异常
				callback(core.ToByteArray(res))
				log.Printf("复制失败,路径：%s,err：%d", compareResult.Path, err)
			}
		}
	}
	return "", false
}

func GenFileList(dir string) {
	var nameList []string
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), ".mp4") {
			nameList = append(nameList, info.Name())
		}
		return nil
	})

	//按名称重低到高排序
	sort.Slice(nameList, func(i, j int) bool {
		name1 := strings.Split(nameList[i], ".")[0]
		name2 := strings.Split(nameList[j], ".")[0]
		return core.StrToInt(name1) < core.StrToInt(name2)
	})

	var builder strings.Builder
	for _, v := range nameList {
		builder.WriteString("file " + v + "\n")
	}

	core.WriteStrToFile(dir+"/input.txt", builder.String())

	//执行命令行,合并文件
	cmdStr := "ffmpeg -f concat -i input.txt -c copy merged_video.mp4"
	core.Execute(cmdStr, dir)
}

// MoveFileList  提前目录中的指定后缀文件移动到目录下面
func MoveFileList(dir string) {
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), ".wma") || strings.HasSuffix(info.Name(), ".mp3") ||
			strings.HasSuffix(info.Name(), ".wav") || strings.HasSuffix(info.Name(), ".mkv") ||
			strings.HasSuffix(info.Name(), ".mkv") || strings.HasSuffix(info.Name(), ".avi") ||
			strings.HasSuffix(info.Name(), ".WAV") || strings.HasSuffix(info.Name(), ".cue") ||
			strings.HasSuffix(info.Name(), ".APE") || strings.HasSuffix(info.Name(), ".ape") ||
			strings.HasSuffix(info.Name(), ".m4a") || strings.HasSuffix(info.Name(), ".ape") ||
			strings.HasSuffix(info.Name(), ".mp4") {
			// 移动文件
			err := os.Rename(path, dir+"/"+info.Name())
			if err != nil {
				log.Printf("移动文件错误:" + err.Error())
			}
		}
		return nil
	})
}

// FileListBySplit 列出文件夹下的所有文件名，逗号分割
func FileListBySplit(rootDir string) string {
	var builder strings.Builder

	// 打开目录
	dir, err := os.Open(rootDir)
	if err != nil {
		fmt.Println("Error opening directory:", err)
		return ""
	}
	defer core.Close(dir)

	// 读取目录内容
	files, err := dir.ReadDir(0) // 第二个参数是读取的条目数量，0表示读取所有
	if err != nil {
		log.Printf("Error reading directory:", err)
		return ""
	}
	// 遍历目录中的文件
	for _, file := range files {
		// 判断是否是文件
		if !file.IsDir() {
			// 处理文件，比如打印文件名
			builder.WriteString(file.Name() + ",")
		}
	}
	return builder.String()
}

// Mkdir

func Mkdir() {

}

func RenameFileByDir(dir string) {
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		err = core.RemoveBeforeFirstChinese(info.Name(), dir)
		if err != nil {
			fmt.Println("Error renaming file:", err)
		}
		return nil
	})
}

// FindDuplicatesFile 遍历文件系统：
// 使用os和path/filepath包来遍历目录和子目录。
// 递归地遍历指定的根目录。
// 计算文件哈希值：
// 使用crypto/md5或crypto/sha256包来计算文件的哈希值。
// 打开文件并读取其内容以计算哈希值。
// 管理哈希值和文件路径：
// 使用map[string][]string来存储哈希值和对应的文件路径列表。
// 当遇到一个文件的哈希值已经在map中时，将其路径添加到对应的列表中。
// 报告重复文件：
// 遍历哈希值map，查找包含多个文件路径的列表。
// 输出或处理这些重复文件。
func FindDuplicatesFile(root string) {
	duplicates := make(map[string][]string)
	var md5Text []string

	// 遍历文件系统
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			hash := core.CalculateCRC(path)
			fileBasicInfo := info.Name() + core.ToString(info.Size()) + core.TimeFormat(info.ModTime())
			md5Text = append(md5Text, core.CalculateMD5(path+fileBasicInfo)+"\t"+hash)
			duplicates[hash] = append(duplicates[hash], path)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error walking the path:", err)
		return
	}

	core.WriteStrToFile("D://md5Text.txt", strings.Join(md5Text, "\n"))

	// 报告重复文件
	for hash, paths := range duplicates {
		if len(paths) > 1 {
			fmt.Printf("Duplicate files with hash %s:\n", hash)
			for _, p := range paths {
				fmt.Println(p)
			}
			fmt.Println()
		}
	}
}
