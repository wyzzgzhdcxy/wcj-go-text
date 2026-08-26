package main

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
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
	lines, err := readLines(req.FilePath)
	if err != nil {
		return FileSplitRes{Success: false, Message: err.Error()}
	}
	total := len(lines)
	if req.LineCount <= 0 {
		req.LineCount = 1000
	}
	var files []string
	base := strings.TrimSuffix(req.FilePath, filepath.Ext(req.FilePath))
	ext := filepath.Ext(req.FilePath)
	for i, start := 0, 0; start < total; i++ {
		end := start + req.LineCount
		if end > total {
			end = total
		}
		partPath := fmt.Sprintf("%s_part%d%s", base, i+1, ext)
		if err := writeLines(partPath, lines[start:end]); err != nil {
			return FileSplitRes{Success: false, Message: err.Error()}
		}
		files = append(files, partPath)
		start = end
	}
	return FileSplitRes{Success: true, Message: fmt.Sprintf("分割为 %d 个文件", len(files)), Files: files}
}

func (a *App) MergeFiles(req FileMergeReq) FileMergeRes {
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
		w.Write(data)
		w.WriteString("\n")
	}
	w.Flush()
	return FileMergeRes{Success: true, Message: "合并完成"}
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

func (a *App) GetFileInfoLocal(filePath string) string {	info, err := os.Stat(filePath)
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
	if assetsFS == (embed.FS{}) {
		return nil, nil
	}
	entries, err := assetsFS.ReadDir("config")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names, nil
}
