package app

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"wcj-go-text/golang/utils"
)

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeLines(path string, lines []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, l := range lines {
		w.WriteString(l)
		w.WriteString("\n")
	}
	return w.Flush()
}

func (a *App) ListFile(dir string) []string {
	return utils.ListFiles(dir)
}

func (a *App) GetVideoFilesInDir(filePath string) []string {
	files := utils.ListFiles(filePath)
	var videos []string
	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f))
		if ext == ".mp4" || ext == ".avi" || ext == ".mkv" || ext == ".mov" || ext == ".flv" {
			videos = append(videos, f)
		}
	}
	return videos
}

func (a *App) RenameFile(dir string) []string {
	utils.RenameFileByDir(dir)
	return utils.ListFiles(dir)
}

func (a *App) Recovery(dir string) []string {
	return utils.RecoveryFilename(dir)
}

func (a *App) CleanFilename(dir string) []string {
	return utils.CleanFilename(dir)
}

func (a *App) CategorizeFiles(dir string, categories string) map[string][]string {
	var cats []string
	json.Unmarshal([]byte(categories), &cats)
	utils.CategorizeFiles(dir, cats)
	result := make(map[string][]string)
	for _, cat := range cats {
		result[cat] = []string{}
	}
	return result
}

func (a *App) GetFilePrefixesInDir(dir string, count int) []string {
	return utils.GetFilePrefixesInDir(dir, count)
}

func (a *App) SplitFile(req FileSplitReq) FileSplitRes {
	if strings.TrimSpace(req.FilePath) == "" {
		return FileSplitRes{Success: false, Message: "源文件路径不能为空"}
	}
	srcInfo, err := os.Stat(req.FilePath)
	if err != nil {
		return FileSplitRes{Success: false, Message: err.Error()}
	}
	if srcInfo.IsDir() {
		return FileSplitRes{Success: false, Message: "源路径是目录,无法分割"}
	}

	outputDir := strings.TrimSpace(req.OutputDir)
	if outputDir == "" {
		outputDir = filepath.Dir(req.FilePath)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return FileSplitRes{Success: false, Message: "创建输出目录失败: " + err.Error()}
	}

	splitType := req.SplitType
	if splitType == "" {
		splitType = "splitBySize"
	}

	var chunkSize int64
	switch splitType {
	case "splitBySize":
		chunkSize, err = parseSize(req.SplitSizeStr)
		if err != nil {
			return FileSplitRes{Success: false, Message: "大小解析失败: " + err.Error()}
		}
		if chunkSize <= 0 {
			return FileSplitRes{Success: false, Message: "每个文件大小必须大于 0"}
		}
	case "splitByCount":
		if req.SplitValue <= 0 {
			return FileSplitRes{Success: false, Message: "分割个数必须大于 0"}
		}
		count := int64(req.SplitValue)
		chunkSize = int64(math.Ceil(float64(srcInfo.Size()) / float64(count)))
		if chunkSize <= 0 {
			chunkSize = 1
		}
	default:
		if req.LineCount > 0 {
			return splitFileByLines(req, outputDir, srcInfo.Size())
		}
		return FileSplitRes{Success: false, Message: "未知的分割方式: " + splitType}
	}

	files, err := splitFileBySize(req.FilePath, outputDir, chunkSize)
	if err != nil {
		return FileSplitRes{Success: false, Message: err.Error()}
	}

	return FileSplitRes{
		Success:      true,
		Message:      fmt.Sprintf("分割为 %d 个文件", len(files)),
		OutputDir:    outputDir,
		TotalSize:    srcInfo.Size(),
		TotalSizeStr: formatSize(srcInfo.Size()),
		SplitFiles:   files,
	}
}

func splitFileByLines(req FileSplitReq, outputDir string, totalSize int64) FileSplitRes {
	lines, err := readLines(req.FilePath)
	if err != nil {
		return FileSplitRes{Success: false, Message: err.Error()}
	}
	total := len(lines)
	if req.LineCount <= 0 {
		req.LineCount = 1000
	}
	base := strings.TrimSuffix(filepath.Base(req.FilePath), filepath.Ext(req.FilePath))
	ext := filepath.Ext(req.FilePath)
	var files []string
	for i, start := 0, 0; start < total; i++ {
		end := start + req.LineCount
		if end > total {
			end = total
		}
		partPath := filepath.Join(outputDir, fmt.Sprintf("%s_part%d%s", base, i+1, ext))
		if err := writeLines(partPath, lines[start:end]); err != nil {
			return FileSplitRes{Success: false, Message: err.Error()}
		}
		files = append(files, partPath)
		start = end
	}
	return FileSplitRes{
		Success:      true,
		Message:      fmt.Sprintf("分割为 %d 个文件", len(files)),
		OutputDir:    outputDir,
		TotalSize:    totalSize,
		TotalSizeStr: formatSize(totalSize),
		SplitFiles:   files,
	}
}

func splitFileBySize(filePath, outputDir string, chunkSize int64) ([]string, error) {
	src, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	base := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	ext := filepath.Ext(filePath)

	buf := make([]byte, chunkSize)
	var files []string
	idx := 0
	for {
		idx++
		n, rerr := io.ReadFull(src, buf)
		if n == 0 {
			break
		}
		if rerr != nil && rerr != io.EOF && rerr != io.ErrUnexpectedEOF {
			return nil, rerr
		}
		partPath := filepath.Join(outputDir, fmt.Sprintf("%s_part%d%s", base, idx, ext))
		if err := os.WriteFile(partPath, buf[:n], 0o644); err != nil {
			return nil, err
		}
		files = append(files, partPath)
		if rerr != nil {
			break
		}
	}
	if len(files) == 0 {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		partPath := filepath.Join(outputDir, fmt.Sprintf("%s_part1%s", base, ext))
		if err := os.WriteFile(partPath, data, 0o644); err != nil {
			return nil, err
		}
		files = append(files, partPath)
	}
	return files, nil
}

func parseSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("大小不能为空")
	}
	upper := strings.ToUpper(s)
	multiplier := int64(1)
	suffix := ""
	switch {
	case strings.HasSuffix(upper, "GB"):
		suffix, multiplier = "GB", 1024*1024*1024
	case strings.HasSuffix(upper, "MB"):
		suffix, multiplier = "MB", 1024*1024
	case strings.HasSuffix(upper, "KB"):
		suffix, multiplier = "KB", 1024
	case strings.HasSuffix(upper, "B"):
		suffix, multiplier = "B", 1
	}
	num := strings.TrimSpace(strings.TrimSuffix(upper, suffix))
	if num == "" {
		return 0, fmt.Errorf("大小格式无效: %s", s)
	}
	var v float64
	if _, err := fmt.Sscanf(num, "%f", &v); err != nil {
		return 0, fmt.Errorf("大小格式无效: %s", s)
	}
	if v <= 0 {
		return 0, fmt.Errorf("大小必须大于 0")
	}
	return int64(math.Ceil(v * float64(multiplier))), nil
}

func formatSize(bytes int64) string {
	if bytes <= 0 {
		return "0 B"
	}
	const k = 1024
	sizes := []string{"B", "KB", "MB", "GB", "TB"}
	i := int(math.Floor(math.Log(float64(bytes)) / math.Log(k)))
	if i >= len(sizes) {
		i = len(sizes) - 1
	}
	if i < 0 {
		i = 0
	}
	val := float64(bytes) / math.Pow(k, float64(i))
	return fmt.Sprintf("%.2f %s", val, sizes[i])
}

func (a *App) MergeFiles(req FileMergeReq) FileMergeRes {
	if len(req.FilePaths) == 0 {
		return FileMergeRes{Success: false, Message: "请选择要合并的文件"}
	}
	if strings.TrimSpace(req.Output) == "" {
		return FileMergeRes{Success: false, Message: "输出文件路径不能为空"}
	}
	if err := os.MkdirAll(filepath.Dir(req.Output), 0o755); err != nil {
		return FileMergeRes{Success: false, Message: "创建输出目录失败: " + err.Error()}
	}

	out, err := os.Create(req.Output)
	if err != nil {
		return FileMergeRes{Success: false, Message: err.Error()}
	}
	defer out.Close()
	w := bufio.NewWriter(out)
	for _, fp := range req.FilePaths {
		data, err := os.ReadFile(fp)
		if err != nil {
			return FileMergeRes{Success: false, Message: err.Error()}
		}
		if _, err := w.Write(data); err != nil {
			return FileMergeRes{Success: false, Message: err.Error()}
		}
	}
	if err := w.Flush(); err != nil {
		return FileMergeRes{Success: false, Message: err.Error()}
	}

	if req.DeleteSrc {
		var failed []string
		for _, fp := range req.FilePaths {
			if fp == req.Output {
				continue
			}
			if err := os.Remove(fp); err != nil {
				failed = append(failed, fmt.Sprintf("%s: %v", fp, err))
			}
		}
		if len(failed) > 0 {
			return FileMergeRes{
				Success: true,
				Message: "合并完成,但删除源文件失败: " + strings.Join(failed, "; "),
				Output:  req.Output,
			}
		}
	}

	return FileMergeRes{Success: true, Message: "合并完成", Output: req.Output}
}

type FileSyncRes struct {
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors"`
}

func (a *App) Compare(tasks []utils.FileSyncRequestTask) FileSyncRes {
	res := utils.CompareDirByStr(tasks)
	r := FileSyncRes{}
	for range res.Result {
		r.Success++
	}
	return r
}

type FolderAnalysisResult struct {
	TotalFiles int      `json:"totalFiles"`
	TotalSize  int64    `json:"totalSize"`
	FileTypes  []string `json:"fileTypes"`
}

func (a *App) AnalyzeFolders() (FolderAnalysisResult, error) {
	return FolderAnalysisResult{}, nil
}

func (a *App) ModifyOriDir(oriDir []string) {}

func (a *App) BackUp(path string) {}

func (a *App) RestoreFiles(tarFn string, path string) {}

func (a *App) MoveFiles(req MoveFilesReq) MoveFilesRes {
	success := 0
	var errors []string
	for _, name := range req.FileNames {
		src := filepath.Join(req.SrcDir, name)
		dst := filepath.Join(req.DstDir, name)
		if err := os.Rename(src, dst); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", name, err))
		} else {
			success++
		}
	}
	return MoveFilesRes{Success: success, Failed: len(req.FileNames) - success, Errors: errors}
}

func (a *App) WriteLinesToFile(filePath string, lines []string) error {
	return writeLines(filePath, lines)
}

func (a *App) ReadFileLines(filePath string) []string {
	lines, _ := readLines(filePath)
	return lines
}

func (a *App) GetFileInfoLocal(filePath string) string {
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return fmt.Sprintf("Name: %s\nSize: %d\nModified: %s",
		info.Name(), info.Size(), info.ModTime().Format("2006-01-02 15:04:05"))
}

func (a *App) GenMySqlDoc(connUrl string) bool {
	return false
}

func (a *App) DdlSql(connUrl string) (string, error) {
	return "", nil
}

func (a *App) GetTableNames(connUrl string) ([]string, error) {
	return []string{}, nil
}

func (a *App) GetTableData(connUrl string, tableName string) ([][]string, error) {
	return [][]string{}, nil
}

func (a *App) GenCodeByConnUrl(connUrl string) string {
	return ""
}

func (a *App) GetLastConnUrl() string {
	return getSetting("last_conn_url")
}

func (a *App) SaveLastConnUrl(connUrl string) {
	saveSetting("last_conn_url", connUrl)
}

func (a *App) OpenTmpDir(path string) error {
	if path == "" {
		path = os.TempDir()
	}
	a.OpenExplorer(path)
	return nil
}

func (a *App) OpenQrCodeExplorer() {
	a.OpenExplorer(os.TempDir())
}

func (a *App) EvaluableExpression(expression string) string {
	return utils.EvaluableExpression(expression)
}

func (a *App) ComputerSummary() string {
	info := utils.GetComputerInfo()
	data, _ := json.Marshal(info)
	return string(data)
}

func (a *App) ShowAlert(message string) {
	fmt.Println("Alert:", message)
}

func (a *App) TurnOffDisplayString() error {
	return utils.TurnOffDisplay()
}

func (a *App) ListAssets() ([]string, error) {
	if Assets == (embed.FS{}) {
		return nil, nil
	}
	entries, err := Assets.ReadDir("config")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names, nil
}
