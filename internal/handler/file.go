package handler

import (
	"github.com/wxnacy/bdpan"
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

func (h *FileHandler) GetDirAllFiles(dir string) ([]*bdpan.FileInfoDto, error) {
	req := file.NewGetFileListReq()
	req.Dir = dir
	totalList := []*bdpan.FileInfoDto{}
	fileList := []*bdpan.FileInfoDto{}
	page := 1
	for {
		req.SetPage(page)
		res, err := file.GetFileList(h.acceccToken, req)
		if err != nil {
			return nil, err
		}
		fileList = res.List
		totalList = append(totalList, fileList...)

		if len(fileList) <= 0 || len(fileList) < int(req.Limit) {
			break
		}
		page++
	}
	return totalList, nil

}
