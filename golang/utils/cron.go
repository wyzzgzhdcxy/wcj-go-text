package utils

import (
	"fmt"
	"github.com/robfig/cron"
	"time"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

type CronRes struct {
	TimeStr string
	Weekday string
}

func NextTimeList(cronExpress string, startDate string, count int) []CronRes {
	var timeList []CronRes
	// 解析cron表达式
	spec, err := cron.Parse(cronExpress)
	if err != nil {
		fmt.Println(err.Error())
		return timeList
	}
	var curTime time.Time
	if len(startDate) != 0 {
		curTime = time.Now()
	} else {
		curTime = core.StringToTime(startDate)
	}
	// 获取下次执行时间
	for i := 0; i < count; i++ {
		nextTime := spec.Next(curTime)
		nextTimeStr := core.TimeFormat(nextTime)
		week := core.WeekdayToChinese(nextTime.Weekday())
		timeList = append(timeList, CronRes{TimeStr: nextTimeStr, Weekday: week})
		curTime = nextTime
	}
	return timeList
}
