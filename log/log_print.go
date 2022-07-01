package log

import (
	"fmt"
	"time"
)

const (
	LOG_DEBUG = iota
	LOG_TRACE
	LOG_INFO
	LOG_WARNING
	LOG_ERROR
	LOG_PANIC
)

type MQTT_LogPring struct {
	level int
}

var (
	GlobalLog MQTT_LogPring
)

func LogInit(level int) {
	GlobalLog.level = level
}

func LogPrint(level int, format string, params ...interface{}) {
	if GlobalLog.level > level {
		return
	}
	var levelStr string
	currentTime := time.Now()
	year, month, day := currentTime.Date()
	hour := currentTime.Hour()
	minute := currentTime.Minute()
	second := currentTime.Second()
	millsecond := currentTime.Nanosecond()/1000000
	switch level {
	case LOG_DEBUG:
		levelStr = "[DEBUG]";
	case LOG_TRACE:
		levelStr = "[TRACE]";
	case LOG_INFO:
		levelStr = "[INFOR]";
	case LOG_WARNING:
		levelStr = "[WARNG]";
	case LOG_ERROR:
		levelStr = "[ERROR]";
	case LOG_PANIC:
		levelStr = "[PANIC]";
	default:
		return;
	}
	logString := fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d %03d->%s:", year,month,day,hour,minute,second,millsecond,levelStr)
	if params == nil {
		logString += format
	} else {
		logString += fmt.Sprintf(format, params...)
	}
	fmt.Println(logString)
}