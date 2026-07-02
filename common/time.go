package common

import (
	"fmt"
	"time"

	"github.com/juhaozero/util/model"
)

type WeekType int

const (
	LastWeek    WeekType = -1
	CurrentWeek WeekType = 0
	NextWeek    WeekType = 1
)

// GetToday 获取一个时间的今天
func GetToday(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}
func GetTimeFormat(format string) string {
	return time.Now().Format(format)
}

// GetDayTimeFormat 获取n天前的日期格式化
// day 偏移的天数
// format 格式化字符串
func GetDayTimeFormat[T model.Number](day T, format string) string {
	return time.Now().AddDate(0, 0, int(day)).Format(format)
}

// GetTimeIsSame 判断时间是否是n天前/后
// times 时间戳
// day 偏移的天数
func GetTimeIsSame[T model.Number](times, day T) bool {
	now := time.Now().AddDate(0, 0, int(day)).Format(time.DateTime)
	sign := time.Unix(int64(times), 0).Format(time.DateTime)
	return now == sign
}

// GetExpDaySecond 获取n天前/后相差的秒数
// day 偏移的天数
func GetExpDaySecond[T model.Number](day T) T {
	timeStr := time.Now().Format("2006-01-02")
	t2, _ := time.ParseInLocation("2006-01-02", timeStr, time.Local)
	if day > 0 {
		return T(t2.AddDate(0, 0, int(day)).Unix() - time.Now().Unix())
	} else {
		return T(time.Now().Unix() - t2.AddDate(0, 0, int(day)).Unix())
	}
}

// GetExpTime 获取偏移后的时间类型
// bufferTime 偏移的时间
func GetExpTime[T model.Number](bufferTime T) time.Time {
	return time.Now().Add(time.Duration(bufferTime))
}

// GetExpDay 获取偏移后的时间戳
// day 偏移的天数
func GetExpDayTime[T model.Number](day T) time.Time {
	return time.Now().AddDate(0, 0, int(day))
}

// GetTimeToUinx 获取时间戳
func GetTimeToUinx[T model.Number](t time.Time) T {
	return T(t.Unix())
}

// GetMonthDays 获取一个时间当月共有多少天
func GetMonthDays(t time.Time) int {
	t = GetToday(t)
	year, month, _ := t.Date()
	if month != 2 {
		if month == 4 || month == 6 || month == 9 || month == 11 {
			return 30
		}
		return 31
	}

	if ((year%4 == 0) && (year%100 != 0)) || year%400 == 0 {
		return 29
	}

	return 28
}

// WeekDay 获取一个时间是星期几
// 1 ~ 7
func WeekDay(t time.Time) int {
	t = GetToday(t)
	week := int(t.Weekday())
	if week == 0 {
		week = 7
	}

	return week
}

// GetSecond 获取共有多少秒
func GetSecond[T model.Number](t time.Time) T {
	return T(t.Unix() - time.Now().Unix())
}

// IsSameDay 两个时间是否是同一天
func IsSameDay(t1, t2 time.Time) bool {
	t1, t2 = GetToday(t1), GetToday(t2)
	return t1.Unix() == t2.Unix()
}

// IsSameHour 两个时间是否是同一小时
func IsSameHour(t1, t2 time.Time) bool {
	return t1.Hour() == t2.Hour() && t1.Day() == t2.Day() && t1.Month() == t2.Month() && t1.Year() == t2.Year()
}

/*
*
获取指定周一的时间
// week 周类型 -1 上周 0 本周 1 下周
// off 偏移的周数
// format 格式化字符串
*/
func WeekIntervalTime(week WeekType, off int, format string) (startTime string) {
	now := time.Now()
	offset := int(time.Monday - now.Weekday())
	//周日做特殊判断 因为time.Monday = 0
	if offset > 0 {
		offset = -6
	}

	year, month, day := now.Date()
	thisWeek := time.Date(year, month, day, 0, 0, 0, 0, time.Local)

	startTime = thisWeek.AddDate(0, 0, offset+7*(int(week)+off)).Format(format)

	return startTime
}

// GetTimeIntervalFormat 获取时间间隔格式化
func GetTimeIntervalFormat(start, end int64) string {
	diff := end - start
	hours := diff / 3600
	minutes := (diff % 3600) / 60
	seconds := diff % 60
	return fmt.Sprintf("%d小时%d分%d秒", hours, minutes, seconds)
}
