package handler

import (
	"github.com/wxnacy/bdpan-cli/internal/config"
	"github.com/wxnacy/bdpan/file"
)

var fileHandler *FileHandler

func GetFileHandler() *FileHandler {
	if fileHandler == nil {
		fileHandler = &FileHandler{
			acceccToken: config.Get().Access.AccessToken,
		}
	}
	return fileHandler
}

type FileHandler struct {
	acceccToken string
}

func (h *FileHandler) GetFileList() {
	file.GetFileList(h.acceccToken, file.NewGetFileListReq())
}
