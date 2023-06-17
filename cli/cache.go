package cli

import (
	"github.com/wxnacy/bdpan-cli/terminal"
)

var (
	// key: fileDirPath
	CacheSelectMap = make(map[string]*terminal.Select, 0)
)
