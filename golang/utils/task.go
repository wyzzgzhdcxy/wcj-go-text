package utils

import (
	"log"

	"github.com/robfig/cron"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

func StartTask() {
	cron2 := cron.New() //创建一个cron实例
	//执行定时任务（每5秒执行一次）
	//if err := cron2.AddFunc("*/1 * * * * *", myComputer.Summary); err != nil {
	//	log.Info(err)
	//}
	if err := cron2.AddFunc("0 0 * * * *", startFileSyncTask); err != nil {
		log.Printf("%v", err)
	}
	//启动/关闭
	cron2.Start()
}

func startFileSyncTask() {
	log.Printf("startFileSyncTask start")
	var fileSyncRequestTask []FileSyncRequestTask
	fileSyncConfig := []byte("[{\"src\":\"D:\\\\个人信息\",\"dst\":\"F:\\\\file_backup\\\\个人信息\"},{\"src\":\"E:\\\\归档文件\",\"dst\":\"F:\\\\file_backup\\\\归档文件\"},{\"src\":\"E:\\\\我的视频\",\"dst\":\"F:\\\\file_backup\\\\我的视频\"},{\"src\":\"E:\\\\我的软件\",\"dst\":\"F:\\\\file_backup\\\\我的软件\"},{\"src\":\"E:\\\\网络文档\",\"dst\":\"F:\\\\file_backup\\\\网络文档\"}]")
	core.JsonToObject(&fileSyncConfig, &fileSyncRequestTask)
	CompareDirByStr(fileSyncRequestTask)
}
