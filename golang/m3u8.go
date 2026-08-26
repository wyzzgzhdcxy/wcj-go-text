package golang

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func DownloadM3u8(url string, fn string, dir string, threadCount int) {
	if core.FileExist(dir + "/" + fn + ".mp4") {
		fmt.Println("文件已经存在", dir+"/"+fn+".mp4")
		return
	}
	os.MkdirAll(dir, 0755)
	os.MkdirAll(core.GetTempDir()+"/tmp", 0755)
	cmd := exec.Command("N_m3u8DL-RE", url,
		"--thread-count", strconv.Itoa(threadCount),
		"--save-name", fn,
		"--save-dir", dir,
		"--tmp-dir", core.GetTempDir()+"/tmp",
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Run()
}

// DownloadM3u8WithProgress 异步下载m3u8并实时回调进度
// status: "downloading" | "completed" | "failed"
func DownloadM3u8WithProgress(url string, fn string, dir string, threadCount int, progressCallback func(status string, percent float64, speed string)) {
	tmpDir := core.GetTempDir() + "/tmp"
	os.MkdirAll(dir, 0755)
	os.MkdirAll(tmpDir, 0755)

	cmd := exec.Command("N_m3u8DL-RE", url,
		"--thread-count", strconv.Itoa(threadCount),
		"--save-name", fn,
		"--save-dir", dir,
		"--tmp-dir", tmpDir,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("创建stdout管道失败:", err)
		progressCallback("failed", 0, "创建stdout管道失败:"+err.Error())
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("创建stderr管道失败:", err)
		progressCallback("failed", 0, "创建stderr管道失败:"+err.Error())
		return
	}

	fmt.Printf("开始下载: url=%s fn=%s dir=%s\n", url, fn, dir)

	if err := cmd.Start(); err != nil {
		fmt.Println("启动命令失败:", err)
		progressCallback("failed", 0, "启动失败:"+err.Error())
		return
	}

	// 共享状态：总分片数（从初始日志解析），用互斥锁保护
	var (
		mu            sync.Mutex
		totalSegments int
	)
	segRegex := regexp.MustCompile(`(\d+)\s*Segments?`)

	// N_m3u8DL-RE 在 Windows 上输出 GBK 编码的中文，用 GBK 解码避免乱码
	gbkDecoder := simplifiedchinese.GBK.NewDecoder()
	decodedReader := transform.NewReader(io.MultiReader(stdout, stderr), gbkDecoder)

	done := make(chan struct{})
	go func() {
		defer close(done)
		scanner := bufio.NewScanner(decodedReader)
		buf := make([]byte, 1024*1024)
		scanner.Buffer(buf, 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println("[m3u8输出]", line)
			// 解析总分片数（INFO 日志里会有形如 "339 Segments" 的字段）
			if m := segRegex.FindStringSubmatch(line); len(m) > 1 {
				if n, err := strconv.Atoi(m[1]); err == nil && n > 0 {
					mu.Lock()
					if totalSegments == 0 {
						totalSegments = n
						fmt.Printf("解析到总分片数: %d\n", totalSegments)
					}
					mu.Unlock()
				}
			}
			parseAndCallback(line, progressCallback)
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("扫描输出错误:", err)
		}
	}()

	// 进度轮询：用文件系统中已下载的分片数估算进度
	// 不依赖 stdout，因为 N_m3u8DL-RE 的进度条走 Windows Console API，
	// 在管道（非 TTY）下不会输出。
	progressDone := make(chan struct{})
	go func() {
		defer close(progressDone)
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		// 等待初始日志解析（最多 10 秒）
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			mu.Lock()
			got := totalSegments > 0
			mu.Unlock()
			if got {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		mu.Lock()
		total := totalSegments
		mu.Unlock()
		if total <= 0 {
			fmt.Println("[m3u8输出] 未解析到总分片数，跳过轮询")
			return
		}

		// 临时目录结构：tmp/<save_name>/<group_id>/<segment>.ts
		// 例如 tmp/demo/0____/000.ts，下载中的分片可能是 .ts.tmp
		var (
			lastPercent float64 = -1
			lastBytes   int64
			lastTime    = time.Now()
			smoothedBps float64
		)
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				count, bytes := countDownloadedSegments(tmpDir, fn)
				percent := float64(count) / float64(total) * 100
				if percent > 100 {
					percent = 100
				}

				// 用两次采样的字节增量除以时间差，得到瞬时速率，再做指数平滑
				now := time.Now()
				dt := now.Sub(lastTime).Seconds()
				var instBps float64
				if dt > 0 && bytes >= lastBytes {
					instBps = float64(bytes-lastBytes) / dt
				}
				if smoothedBps == 0 {
					smoothedBps = instBps
				} else {
					smoothedBps = smoothedBps*0.6 + instBps*0.4
				}
				lastBytes = bytes
				lastTime = now

				// "downloading" 阶段最多上报到 99%；分片全部下载完成但进程尚未退出
				// （N_m3u8DL-RE 正在合并/封装最终视频）时，切换到 "merging" 状态，
				// 避免前端把任务误判为已完成。
				status := "downloading"
				if count >= total {
					status = "merging"
					percent = 99
				}
				if percent > 99 && status == "downloading" {
					percent = 99
				}

				if percent-lastPercent >= 0.5 || status == "merging" {
					speed := formatSpeed(smoothedBps)
					progressCallback(status, percent, speed)
					lastPercent = percent
					fmt.Printf("[m3u8轮询] %d/%d = %.2f%%, status=%s, %s\n", count, total, percent, status, speed)
				}
			}
		}
	}()

	<-done
	<-progressDone

	waitErr := cmd.Wait()
	if waitErr != nil {
		fmt.Println("命令执行错误:", waitErr)
		progressCallback("failed", 0, "命令执行失败: "+waitErr.Error())
		return
	}

	outputFile := dir + "/" + fn + ".mp4"
	if !core.FileExist(outputFile) {
		progressCallback("failed", 0, "未找到输出文件: "+outputFile)
		return
	}

	progressCallback("completed", 100, "")
}

// countDownloadedSegments 统计 tmp/<fn> 下已下载的分片文件数和总字节数
// 包含下载完成的 .ts 和下载中的 .ts.tmp
func countDownloadedSegments(tmpDir, fn string) (int, int64) {
	count := 0
	var totalBytes int64
	root := filepath.Join(tmpDir, fn)
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		name := info.Name()
		if strings.HasSuffix(name, ".ts") || strings.HasSuffix(name, ".ts.tmp") {
			count++
			totalBytes += info.Size()
		}
		return nil
	})
	return count, totalBytes
}

// formatSpeed 把字节/秒格式化成 "X.XX MB/s" 或 "XXX KB/s"
func formatSpeed(bps float64) string {
	if bps <= 0 {
		return ""
	}
	const kb = 1024.0
	const mb = 1024.0 * 1024.0
	if bps >= mb {
		return fmt.Sprintf("%.2f MB/s", bps/mb)
	}
	return fmt.Sprintf("%.0f KB/s", bps/kb)
}

func parseAndCallback(line string, callback func(string, float64, string)) {
	percentRegex := regexp.MustCompile(`(\d+(?:\.\d+)?)%`)
	percentMatches := percentRegex.FindStringSubmatch(line)
	percent := float64(0)
	if len(percentMatches) > 1 {
		percent, _ = strconv.ParseFloat(percentMatches[1], 64)
	}

	speedRegex := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(MB/s|KB/s|MBps|KBps|mb/s|kb/s|mbps|kbps)`)
	speedMatches := speedRegex.FindStringSubmatch(line)
	speed := ""
	if len(speedMatches) > 2 {
		unit := speedMatches[2]
		lower := strings.ToLower(unit)
		if lower == "mbps" || lower == "kbps" {
			unit = unit[:len(unit)-2] + "/s"
		} else if lower == "mb/s" || lower == "kb/s" {
			unit = strings.ToUpper(unit)
		}
		speed = speedMatches[1] + " " + unit
	}

	// 上限 99：避免 N_m3u8DL-RE 在合并阶段输出 100% 时被前端误判为下载完成
	if percent > 99 {
		percent = 99
	}

	if percent > 0 || speed != "" {
		callback("downloading", percent, speed)
	}
}
