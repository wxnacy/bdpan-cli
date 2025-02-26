package handler

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/bdpan-cli/internal/config"
	"github.com/wxnacy/bdpan-cli/internal/dto"
	"github.com/wxnacy/bdpan-cli/internal/model"
	"github.com/wxnacy/bdpan-cli/internal/tasker"
	"github.com/wxnacy/bdpan/file"
	"github.com/wxnacy/dler"
	"github.com/wxnacy/go-tools"
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

func (h *FileHandler) CmdList(req *dto.ListReq) error {
	// model.GetDB().AutoMigrate(&model.File{})
	r := file.NewGetFileListReq()
	r.Dir = req.Path
	r.Limit = req.Limit
	r.SetPage(req.Page)
	res, err := file.GetFileList(h.acceccToken, r)
	if err != nil {
		return err
	}
	for _, f := range res.List {
		file := model.FindFirstByID(f.FSID)

		// var size = int64(f.Size)
		// if f.IsDir() {
		// var path = FormatPath(f.Path)
		// sfiles := model.FindFilesPrefixPath(path, false)
		// for _, sf := range sfiles {
		// size += int64(sf.Size)
		// }
		// }
		fmt.Printf("%18d\t%s\t%s\t%s\n", f.FSID, f.GetFileType(), file.GetSize(), f.Path)
	}
	return nil
}

func (h *FileHandler) CmdRefresh(req *dto.RefreshReq) error {
	// model.GetDB().AutoMigrate(&model.File{})
	path := FormatPath(req.Path)

	// 刷新目标目录
	if req.IsSync {
		h.refreshFiles(path)
		for {
			// 获取需要刷新的目录逐级刷新
			infos := model.FindNeedRefreshFiles(path)
			if len(infos) == 0 {
				break
			}
			fmt.Printf("Refresh %s dir count: %d\n", path, len(infos))
			total := len(infos)
			for i, f := range infos {
				begin := time.Now()
				h.refreshFiles(f.Path)
				f.IsRefresh = 1
				f.Save()
				timeUsed := time.Now().Sub(begin)
				fmt.Printf("[%d/%d]Saved path: %s time used: %v\n", i, total, f.Path, timeUsed)
			}
		}
	}

	fmt.Println("开始刷新目录数据大小")
	curDir := model.FindFirstByPath(req.Path)
	h.refreshDirSize(curDir)
	refreshDirs := model.FindFilesPrefixPath(path, true)
	for _, dir := range refreshDirs {
		h.refreshDirSize(dir)
	}

	return nil
}

func (h *FileHandler) refreshFiles(path string) error {
	if path == "/" {
		model.NewRootFile().Resave()
	}
	files, err := h.GetDirAllFiles(path)
	if err != nil {
		return err
	}
	for _, f := range files {
		model.NewFileFromDto(f).Resave()
	}
	return nil
}

func (h *FileHandler) refreshDirSize(dir *model.File) error {
	dPath := FormatPath(dir.Path)
	subFiles := model.FindFilesPrefixPath(dPath, false)
	size := 0
	for _, sf := range subFiles {
		size += sf.Size
	}
	dir.Size = size
	res := dir.Save()
	fmt.Printf(
		"Path: %s Size: %s UpdateCount: %d Err: %v\n",
		dPath,
		tools.FormatSize(int64(size)),
		res.RowsAffected,
		res.Error,
	)
	return nil
}

func FormatPath(path string) string {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}
