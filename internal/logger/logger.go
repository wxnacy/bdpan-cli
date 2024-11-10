package logger

import "github.com/wxnacy/bdpan"

func Debugf(format string, args ...interface{}) {
	bdpan.Log.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	bdpan.Log.Infof(format, args...)
}

func Errorf(format string, args ...interface{}) {
	bdpan.Log.Errorf(format, args...)
}
