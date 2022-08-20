package common

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	TimeNow string = "now"
)

func GetTimeFromDuration(d time.Duration) (time.Time, time.Time) {
	now := time.Now()
	then := now.Add(-d)
	return then, now

}

func toValidDuration(durationStr string) (string, error) {
	if strings.HasSuffix(durationStr, "d") {
		dStr := strings.Trim(durationStr, "d")
		days, err := strconv.Atoi(dStr)
		if err != nil {
			return durationStr, nil
		}
		hours := days * 24
		return fmt.Sprintf("%dh", hours), nil
	}
	return durationStr, nil
}

//durationStr:  "10h", "1h10m10s"
func GetTimeFromNow(durationStr string) (time.Time, time.Time, error) {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed parse time %s - %s", durationStr, err.Error())
	}

	then, now := GetTimeFromDuration(duration)
	return then, now, err
}

func GetTimeWindow(windowSize string, windowEndOffset string) (time.Time, time.Time, error) {
	if windowEndOffset == TimeNow {
		return GetTimeFromNow(windowSize)
	}

	windowSize, err := toValidDuration(windowSize)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed converting to valid duration window %s - %s", windowEndOffset, err.Error())
	}

	windowEndOffset, err = toValidDuration(windowEndOffset)

	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed converting to valid duration window offset %s - %s", windowEndOffset, err.Error())
	}

	windowDuration, err := time.ParseDuration(windowSize)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed parsing time %s - %s ", windowSize, err.Error())
	}

	windowEnd, _, err := GetTimeFromNow(windowEndOffset)

	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed parsing window offset time %s - %s", windowEndOffset, err.Error())
	}
	startWindow := windowEnd.Add(-windowDuration)

	return startWindow, windowEnd, nil
}

func Get_ISO_UTC_Timeoffset() string {
	_, secondsOffset := time.Now().Zone()
	hourOffset := secondsOffset / 3600
	offset := ""
	if hourOffset >= 0 {
		offset = "+"
	}
	return fmt.Sprintf("UTC%s%d", offset, hourOffset)
}
