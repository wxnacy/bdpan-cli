package handler

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/wxnacy/bdpan-cli/internal/api"
	"github.com/wxnacy/bdpan-cli/internal/config"
	"github.com/wxnacy/bdpan-cli/internal/dto"
	"github.com/wxnacy/bdpan-cli/internal/model"
	"github.com/wxnacy/bdpan-cli/internal/tasker"
	"github.com/wxnacy/dler"
	"github.com/wxnacy/go-bdpan"
	"github.com/wxnacy/go-tools"
)

var fileHandler *FileHandler

func GetFileHandler() *FileHandler {
	if fileHandler == nil {
		fileHandler = &FileHandler{
			acceccToken: config.Get().Access.AccessToken,
			limit:       1000,
		}
	}
	return fileHandler
}

type FileHandler struct {
	acceccToken string
	limit       int32
}

func (h *FileHandler) GetFiles(dir string, page int) ([]*model.File, error) {
	req := bdpan.NewGetFileListReq().SetDir(dir).SetLimit(h.limit).SetPage(page)
	res, err := bdpan.GetFileList(h.acceccToken, req)
	if err != nil {
		return nil, err
	}
	return model.NewFiles(res.List), nil
}

func (h *FileHandler) GetFilesAndSave(dir string, page int) ([]*model.File, error) {
	files, err := h.GetFiles(dir, page)
	if err != nil {
		return nil, err
	}
	for _, v := range files {
		model.Save(v)
	}
	return files, nil
}

func (h *FileHandler) GetFilesFromDBOrReal(dir string, page int) ([]*model.File, error) {
	var err error
	files := model.FindFilesByDir(dir, page)
	if len(files) == 0 {
		files, err = h.GetFiles(dir, page)
		if err != nil {
			return nil, err
		}
		for _, v := range files {
			model.Save(v)
		}
	}
	return files, nil
}

func (h *FileHandler) GetDirAllFiles(dir string) ([]*bdpan.FileInfo, error) {
	req := bdpan.NewGetFileListReq()
	req.Dir = dir
	totalList := []*bdpan.FileInfo{}
	fileList := []*bdpan.FileInfo{}
	page := 1
	for {
		req.SetPage(page)
		res, err := bdpan.GetFileList(h.acceccToken, req)
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

func (h *FileHandler) DeleteFile(path string) (*bdpan.ManageFileRes, error) {
	return bdpan.DeleteFile(h.acceccToken, path)
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
func (h *FileHandler) GetFileByPath(path string) (*bdpan.FileInfo, error) {

	getFileByPage := func(dir, name string, page int) (*bdpan.FileInfo, bool, error) {
		req := bdpan.NewGetFileListReq()
		req.Dir = dir
		req.SetPage(page)
		res, err := bdpan.GetFileList(h.acceccToken, req)
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
			req := bdpan.NewGetFileInfoReq(f.FSID)

			// 获取带有下载地址的文件详情
			infoRes, err := bdpan.GetFileInfo(h.acceccToken, req)
			if err != nil {
				return nil, err
			}
			info := &infoRes.FileInfo
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

func (h *FileHandler) CmdDelete(req *dto.DeleteReq) error {
	var info *bdpan.FileInfo
	var err error
	if req.FSID > 0 {
		fmt.Println("通过 FSID 查询文件")
		info, err = api.GetFileInfo(h.acceccToken, req.FSID)
	} else {
		fmt.Println("通过 Path 查询文件")
		info, err = h.GetFileByPath(req.Path)
	}
	if err != nil {
		fmt.Println("找不到文件")
		return nil
	}
	if info.IsDir() {
		if !req.Yes {
			var confirm bool
			err = huh.NewConfirm().
				Title("目标是个目录，是否确认删除").
				Affirmative("Yes!").
				Negative("No.").
				Value(&confirm).WithTheme(huh.ThemeCatppuccin()).Run()
			if err != nil {
				return nil
			}
			if !confirm {
				fmt.Println("取消删除")
				return nil
			}
		}
	}
	var path = info.Path
	res, err := bdpan.DeleteFile(h.acceccToken, path)
	if err != nil {
		return err
	}
	fmt.Printf("删除文件: %s 成功\n", path)
	if res.Taskid > 0 {
		fmt.Printf("异步删除，任务 ID: %d\n", res.Taskid)
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
		model.NewFile(f).Resave()
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

func (h *FileHandler) Limit(l int32) *FileHandler {
	h.limit = l
	return h
}

func FormatPath(path string) string {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}
