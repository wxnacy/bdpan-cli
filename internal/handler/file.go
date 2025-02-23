package handler

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/bdpan-cli/internal/config"
	"github.com/wxnacy/bdpan-cli/internal/dto"
	"github.com/wxnacy/bdpan-cli/internal/tasker"
	"github.com/wxnacy/bdpan/file"
	"github.com/wxnacy/dler"
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

func (h *FileHandler) CmdDownload(req *dto.DownloadReq) error {
	fmt.Printf("查找文件地址: %s\n", req.Path)
	f, err := h.GetFileByPath(req.Path)
	if err != nil {
		return err
	}
	fmt.Printf("文件ID: %d\n", f.FSID)
	if f.IsDir() {
		fmt.Println("文件类型是文件夹")
		_, name := filepath.Split(f.Path)
		tasker.DownloadFile(f, filepath.Join(req.OutputDir, name))
	} else {
		fmt.Printf("文件下载地址: %s\n", f.Dlink)
		t := dler.NewFileDownloadTasker(f.Dlink).
			SetDownloadPath(req.OutputPath).SetDownloadDir(req.OutputDir).
			SetCacheDir(req.OutputDir)
		if req.IsVerbose {
			t.Request.EnableVerbose()
		}
		// t.Out = d.Out
		// t.IsNotCover = d.IsNotCover
		// t.OutputFunc = LogInfoString

		// t.Config.UseProgressBar = true
		err = t.Exec()
	}

	// downloadSmallFile(f.Dlink, "/Users/wxnacy/Downloads/test1212.mp4")

	return err
}

// 根据地址查找文件
// 在文件目录中循环查找是否有该名称文件
func (h *FileHandler) GetFileByPath(path string) (*bdpan.FileInfoDto, error) {

	getFileByPage := func(dir, name string, page int) (*bdpan.FileInfoDto, bool, error) {
		req := file.NewGetFileListReq()
		req.Dir = dir
		req.SetPage(page)
		res, err := file.GetFileList(h.acceccToken, req)
		// fmt.Println(req)
		// fmt.Println(page, res, err)
		if err != nil {
			return nil, false, err
		}
		// 返回是否有下一页
		if len(res.List) == 0 {
			return nil, false, nil
		}
		// 过滤文件
		for _, f := range res.List {
			if f.GetFilename() == name {
				return f, false, nil
			}
		}
		return nil, true, nil
	}
	dir, name := filepath.Split(path)
	// 10万页循环
	for i := 0; i < 100000; i++ {
		f, hasMove, err := getFileByPage(dir, name, i+1)
		if err != nil {
			return nil, err
		}
		if f != nil {
			req := file.NewGetFileInfoReq(f.FSID)

			// 获取带有下载地址的文件详情
			infoRes, err := file.GetFileInfo(h.acceccToken, req)
			if err != nil {
				return nil, err
			}
			info := &infoRes.FileInfoDto
			info.Dlink = fmt.Sprintf("%s&access_token=%s", info.Dlink, h.acceccToken)
			return info, nil
		} else {
			// 判断是否有下一页，没有直接返回
			if !hasMove {
				break
			}
		}
	}
	return nil, errors.New("file not found")
}
