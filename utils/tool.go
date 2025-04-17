package utils

import (
	"regexp"
	"time"
)

func RegexpUtils(str, s string) []string {
	//定义正则表达式
	regexpCompile := regexp.MustCompile(str)
	//使用正则表达式找与之相匹配的字符串，返回一个数组包含子表达式匹配的字符串
	return regexpCompile.FindStringSubmatch(s)
}
func TimeFormat(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
func TimeFormat20060102(t time.Time) string {
	return t.Format("20060102150405")
}
func TimeParse(t string) string {
	t0, _ := time.Parse("20060102150405", t)
	return TimeFormat(t0)
}

//获取当前月的时间函数封装
func GetMonthDay(now time.Time) (string, string) {
	//now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	//f := firstOfMonth.Unix()
	//l := lastOfMonth.Unix()
	//return time.Unix(f, 0).Format("2006-01-02") + " 00:00:00", time.Unix(l, 0).Format("2006-01-02") + " 23:59:59"

	//上面注释掉的特么多此一举，稍微优化一下这个函数
	return firstOfMonth.Format("2006-01-02") + " 00:00:00", lastOfMonth.Format("2006-01-02") + " 23:59:59"
}
