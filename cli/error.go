package cli

import (
	"errors"

	"github.com/wxnacy/bdpan"
)

var ErrQuit = errors.New("quit bdpan")
var ErrNotCopyFile = errors.New("请先使用 yy 复制文件")
var ErrFileExists = errors.New("文件已存在")
var ErrActionFail = errors.New("操作失败")
var ErrNoFileSelect = errors.New("请先选择文件")
var ErrNoTypeToPaste = errors.New("请先进行复制或剪切操作")

var BottomErrs = []error{
	ErrNotCopyFile,
	ErrActionFail,
	bdpan.ErrPathExists,
	ErrNoFileSelect,
	ErrNoTypeToPaste,
}

func IsInErrors(e error, errs []error) bool {
	for _, err := range errs {
		if e.Error() == err.Error() {
			return true
		}
	}
	return false
}

func CanCacheError(e error) bool {
	return IsInErrors(e, BottomErrs)
}
