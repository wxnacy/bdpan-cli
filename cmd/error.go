package cmd

import "errors"

var ErrQuit = errors.New("quit bdpan")
var ErrNotCopyFile = errors.New("请先使用 yy 复制文件")

var BottomErrs = []error{
	ErrNotCopyFile,
}

func IsInErrors(e error, errs []error) bool {
	for _, err := range errs {
		if e.Error() == err.Error() {
			return true
		}
	}
	return false
}
