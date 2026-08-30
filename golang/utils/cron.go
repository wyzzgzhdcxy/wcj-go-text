package utils

import (
	"fmt"
	"github.com/robfig/cron"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
	"time"
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
		curTime = core.StringToTime(startDate)
	} else {
		curTime = time.Now()
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
