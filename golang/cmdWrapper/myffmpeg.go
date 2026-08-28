package cmdWrapper

import (
	"fmt"
	"path/filepath"
	"strings"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

func checkInvalidCharacter(fn string) bool {
	invalidChar := []string{"(", ")", "&"}
	for _, v := range invalidChar {
		if strings.Contains(fn, v) {
			return true
		}
	}
	return false
}

// GenPng 获取视频10S后的第一帧
func GenPng(fn string) {
	output := fmt.Sprintf("png/%s.png", fn)
	executeFfmpegCmd("-ss", "10", "-i", fn, "-vframes", "1", output)
}

// MergeMp4ByFileList 合并视频列表 ffmpeg -f concat -safe 0 -i filelist.txt -c copy output.mp4
func MergeMp4ByFileList(fn string, targetFn string) {
	executeFfmpegCmd("-f", "concat", "-safe", "0", "-i", fn, "-c", "copy", targetFn)
}

// Mp3 分离mp4视频流中的音频流,MP3格式
// -i 1.mp4 -q:a 0 -map a 1.mp3
func Mp3(fn string) {
	output := fn[0:len(fn)-1] + "3"
	executeFfmpegCmd("-i", fn, "-q:a", "0", "-map", "a", output)
}

// MergeMp4Mp3 ffmpeg -i input.mp4 -i new.mp3 -c:v copy -map 0:v:0 -map 1:a:0 output.mp4
func MergeMp4Mp3(mp4Fn string, mp3Fn string) {
	tmpMp4Fn := mp4Fn + ".mp4"
	core.RenameFile(mp4Fn, tmpMp4Fn)
	executeFfmpegCmd("-i", tmpMp4Fn, "-i", mp3Fn, "-c:v", "copy", "-map", "0:v:0", "-map", "1:a:0", mp4Fn)
}

// SplitVideo ffmpeg -ss 00:04:00 -noaccurate_seek -i "1.mp4" -c copy -reset_timestamps 1 -avoid_negative_ts 1 output2.mp4
// SplitVideo 重指定时间开始切割视频
func SplitVideo(fn string, startTime string) {
	input := fmt.Sprintf("input/%s", fn)
	out := fmt.Sprintf("out/%s", fn)
	start := fmt.Sprintf("00:%s", startTime)
	executeFfmpegCmd("-ss", start, "-noaccurate_seek", "-i", input, "-c", "copy", "-reset_timestamps", "1", "-avoid_negative_ts", "1", out)
}

func DeleteVideoOneMinute(fp string) {
	core.MkDirALl0755(core.GetTempDir() + "/videoTemp")
	tmpFilepath := core.GetTempDir() + "/videoTemp/" + filepath.Base(fp)
	core.RenameFile(fp, tmpFilepath)
	executeFfmpegCmd("-i", tmpFilepath, "-ss", "00:01:00", "-c", "copy", fp)
}

// TransposeVideo 旋转视频
func TransposeVideo(fn string, transpose int) {
	// 构建FFmpeg命令
	trans := fmt.Sprintf("transpose=%d", transpose)
	out := fmt.Sprintf("out/%s", fn)
	//"-c:v", "h264_qsv" 这个是使用GPU加速，如果不需要GPU加速可以删除
	//经过测试，这个方法不适合使用intelGPU独显加速，速度反而更慢
	executeFfmpegCmd("-i", fmt.Sprintf("input/%s", fn), "-vf", trans, out)
}

func executeFfmpegCmd(arg ...string) string {
	output, err := RunWithOutput("ffmpeg", arg...)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return output
}

func executeFfprobeCmd(arg ...string) string {
	output, err := RunWithOutput("ffprobe", arg...)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return output
}

// GetVideoResolution 获取视频的分辨率 width,height
func GetVideoResolution(fn string) (int, int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("捕获到 panic:", r)
		}
	}()
	if checkInvalidCharacter(fn) {
		fmt.Printf("文件名包含非法字符: %s", fn)
		return 0, 0
	}
	output := executeFfprobeCmd("-v", "error", "-show_entries", "stream=width,height", "-of", "default=noprint_wrappers=1:nokey=1", fn)
	arr := strings.Split(output, "\n")
	if len(arr) > 1 {
		return core.StrToInt(strings.TrimSpace(arr[0])), core.StrToInt(strings.TrimSpace(arr[1]))
	} else {
		return 0, 0
	}
}
