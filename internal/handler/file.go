package handler

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/wxnacy/bdpan-cli/internal/api"
	"github.com/wxnacy/bdpan-cli/internal/config"
	"github.com/wxnacy/bdpan-cli/internal/dto"
	"github.com/wxnacy/bdpan-cli/internal/logger"
	"github.com/wxnacy/bdpan-cli/internal/model"
	"github.com/wxnacy/bdpan-cli/internal/tasker"
	"github.com/wxnacy/bdpan-cli/pkg/bdtools"
	"github.com/wxnacy/dler"
	"github.com/wxnacy/go-bdpan"
	gotasker "github.com/wxnacy/go-tasker"
	"github.com/wxnacy/go-tools"
)

var fileHandler *FileHandler

func GetFileHandler() *FileHandler {
	if fileHandler == nil {
		fileHandler = &FileHandler{
			accessToken: config.Get().Access.AccessToken,
			limit:       1000,
		}
	}
	return fileHandler
}

type FileHandler struct {
	accessToken string
	limit       int32
}

func (h *FileHandler) GetFiles(dir string, page int) ([]*model.File, error) {
	req := bdpan.NewGetFileListReq().SetDir(dir).SetLimit(h.limit).SetPage(page)
	res, err := bdpan.GetFileList(h.accessToken, req)
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
		res, err := bdpan.GetFileList(h.accessToken, req)
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

func (h *FileHandler) DeleteFiles(paths ...string) (*bdpan.ManageFileRes, error) {
	return bdpan.DeleteFiles(h.accessToken, paths...)
}

func (h *FileHandler) MoveFiles(dir string, paths ...string) (*bdpan.ManageFileRes, error) {
	return bdpan.MoveFiles(h.accessToken, dir, paths...)
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
	return bdtools.GetFileByPath(h.accessToken, path)
}

func (h *FileHandler) CmdDelete(req *dto.DeleteReq) error {
	var info *bdpan.FileInfo
	var err error
	if req.FSID > 0 {
		fmt.Println("通过 FSID 查询文件")
		info, err = api.GetFileInfo(h.accessToken, req.FSID)
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
	path := info.Path
	res, err := bdpan.DeleteFiles(h.accessToken, path)
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
				timeUsed := time.Since(begin)
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

func (h *FileHandler) CmdUpload(req *dto.UploadReq) error {
	logger.Printf("上传文件 %s => %s", req.Local, req.Path)
	file, err := bdtools.UploadFile(
		h.accessToken,
		req.Local,
		req.Path,
		gotasker.NewBubblesProgressBar(),
		bdtools.Printf(logger.Infof),
		bdtools.IsRewrite(req.IsRewrite),
	)
	if err != nil {
		return err
	}
	logger.Printf("上传文件成功")
	bdtools.PrintFileInfo(file)
	return nil
}

func FormatPath(path string) string {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}
