package common

import (
	"time"

	"github.com/juhaozero/util/model"
)

type WeekType int

const (
	LastWeek    WeekType = -1
	CurrentWeek WeekType = 0
	NextWeek    WeekType = 1
)

func GetTimeMicro() int64 {
	return time.Now().UnixMicro()
}
func GetTimeNano() int64 {
	return time.Now().UnixNano()
}
func GetTimeMs() int64 {
	return time.Now().UnixMilli()
}
func GetTime() int64 {
	return time.Now().Unix()
}
func GetTimeFormat(format string) string {
	return time.Now().Format(format)
}
func GetNumIsEven[T model.Number](data T) bool {
	return int64(data)&1 == 0
}
func GetDayTimeFormat[T model.Number](day T, format string) string {
	return time.Now().AddDate(0, 0, int(day)).Format(format)
}

// GetTimeIsSame 判断时间是否是n天前/后
// day 偏移的天数
func GetTimeIsSame[T model.Number](times, day T, format string) bool {
	now := time.Now().AddDate(0, 0, int(day)).Format(format)
	sign := time.Unix(int64(times), 0).Format(format)
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
func GetExpTime[T model.Number](bufferTime T) time.Time {
	return time.Now().Add(time.Duration(bufferTime))
}

// GetExpUnix 获取偏移后的时间戳
func GetExpUnix(day int) int64 {
	return time.Now().AddDate(0, 0, day).Unix()
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
//   - 1 ~ 7
func WeekDay(t time.Time) int {
	t = GetToday(t)
	week := int(t.Weekday())
	if week == 0 {
		week = 7
	}

	return week
}

// GetToday 获取一个时间的今天
func GetToday(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

// GetSecond 获取共有多少秒
func GetSecond(d time.Duration) int {
	return int(d / time.Second)
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
	//endTime = thisWeek.AddDate(0, 0, offset+6+7*week).Format("2006-01-02")

	return startTime
}
