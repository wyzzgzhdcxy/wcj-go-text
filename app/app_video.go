package app

import (
	"fmt"
	"os"
	"sort"

	"wcj-go-text/golang/myVideo"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

// FileItem 目录文件项
type FileItem struct {
	Name     string `json:"name"`
	IsDir    bool   `json:"isDir"`
	FileSize int64  `json:"fileSize"`
}

// BatchExtractAudio 批量从视频中提取音频
func (a *App) BatchExtractAudio(req myVideo.BatchExtractAudioReq) myVideo.BatchExtractAudioRes {
	return myVideo.BatchExtractAudio(req, a.sendVideoMsg)
}

// BatchExtractAudioByFiles 按指定文件列表提取音频
func (a *App) BatchExtractAudioByFiles(req myVideo.BatchExtractAudioByFilesReq) myVideo.BatchExtractAudioRes {
	return myVideo.BatchExtractAudioByFiles(req, a.sendVideoMsg)
}

// RotateVideo 旋转视频
func (a *App) RotateVideo(req myVideo.RotateVideoReq) myVideo.RotateVideoRes {
	return myVideo.RotateVideo(req, a.sendVideoMsg)
}

// ExtractFramesByFiles 批量抽帧
func (a *App) ExtractFramesByFiles(req myVideo.ExtractFramesByFilesReq) myVideo.ExtractFramesByFilesRes {
	return myVideo.ExtractFramesByFiles(req, a.sendVideoMsg)
}

// ExtractAudio 从视频中分离音频
func (a *App) ExtractAudio(req myVideo.ExtractAudioReq) myVideo.ExtractAudioRes {
	return myVideo.ExtractAudio(req, a.sendVideoMsg)
}

// ScanVideos 扫描指定目录中的所有视频文件
func (a *App) ScanVideos(req myVideo.ScanVideosReq) myVideo.ScanVideosRes {
	return myVideo.ScanVideos(req, a.sendVideoMsg)
}

// ExtractFrames 从视频中随机抽取 N 帧
func (a *App) ExtractFrames(req myVideo.ExtractFramesReq) myVideo.ExtractFramesRes {
	return myVideo.ExtractFrames(req, a.sendVideoMsg)
}

// ExtractVideoThumbnail 从视频中提取一帧作为缩略图
func (a *App) ExtractVideoThumbnail(req myVideo.ExtractVideoThumbnailReq) myVideo.ExtractVideoThumbnailRes {
	return myVideo.ExtractVideoThumbnail(req)
}

// ScanVideoDir 扫描视频目录，统计视频个数和各类型个数
func (a *App) ScanVideoDir(req myVideo.ScanVideoDirReq) myVideo.ScanVideoDirRes {
	return myVideo.ScanVideoDir(req, a.sendVideoMsg)
}

// TrimVideoStart 去除视频片头
func (a *App) TrimVideoStart(req myVideo.TrimVideoStartReq) myVideo.TrimVideoStartRes {
	return myVideo.TrimVideoStart(req, a.sendVideoMsg)
}

// TrimVideoStartByFiles 批量去除视频片头
func (a *App) TrimVideoStartByFiles(req myVideo.TrimVideoStartByFilesReq) myVideo.TrimVideoStartByFilesRes {
	return myVideo.TrimVideoStartByFiles(req, a.sendVideoMsg)
}

// TrimVideoEnd 去除视频片尾
func (a *App) TrimVideoEnd(req myVideo.TrimVideoEndReq) myVideo.TrimVideoEndRes {
	return myVideo.TrimVideoEnd(req, a.sendVideoMsg)
}

// TrimVideoEndByFiles 批量去除视频片尾
func (a *App) TrimVideoEndByFiles(req myVideo.TrimVideoEndByFilesReq) myVideo.TrimVideoEndByFilesRes {
	return myVideo.TrimVideoEndByFiles(req, a.sendVideoMsg)
}

// GetKeyframes 获取视频关键帧位置信息
func (a *App) GetKeyframes(req myVideo.GetKeyframesReq) myVideo.GetKeyframesRes {
	return myVideo.GetKeyframes(req, a.sendVideoMsg)
}

// MergeVideos 将目录下的子目录中的视频分别合并
func (a *App) MergeVideos(req myVideo.MergeVideosReq) myVideo.MergeVideosRes {
	return myVideo.MergeVideos(req, a.sendVideoMsg)
}

// MergeVideosByFiles 合并指定的视频文件列表
func (a *App) MergeVideosByFiles(req myVideo.MergeVideosByFilesReq) myVideo.MergeVideosByFilesRes {
	return myVideo.MergeVideosByFiles(req, a.sendVideoMsg)
}

// ClassifyVideosByResolution 将目录下的视频按分辨率分类到不同文件夹
func (a *App) ClassifyVideosByResolution(req myVideo.ClassifyVideosByResolutionReq) myVideo.ClassifyVideosByResolutionRes {
	return myVideo.ClassifyVideosByResolution(req, a.sendVideoMsg)
}

// sendVideoMsg 视频处理回调（发送到前端 back_msg 事件）
func (a *App) sendVideoMsg(message []byte) {
	fmt.Println("send to frontend:", string(message))
	runtime.EventsEmit(a.ctx, "back_msg", string(message))
}

// GetDirContents 获取目录内容（文件列表）
func (a *App) GetDirContents(dirPath string) ([]FileItem, error) {
	if dirPath == "" {
		return nil, fmt.Errorf("目录路径为空")
	}

	if !core.FileExist(dirPath) {
		return nil, fmt.Errorf("目录不存在：%s", dirPath)
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败：%v", err)
	}

	var items []FileItem
	for _, entry := range entries {
		var fileSize int64
		if !entry.IsDir() {
			info, err := entry.Info()
			if err == nil {
				fileSize = info.Size()
			}
		}
		items = append(items, FileItem{
			Name:     entry.Name(),
			IsDir:    entry.IsDir(),
			FileSize: fileSize,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	return items, nil
}
