package handler

import (
	"errors"
	"fmt"
	"os"
	"path"
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
			accessToken: config.GetAccessToken(),
			limit:       1000,
		}
	}
	return fileHandler
}

type FileHandler struct {
	accessToken string
	limit       int32
}

func (h *FileHandler) GetAccessToken() string {
	return h.accessToken
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

func (h *FileHandler) RenameFile(pathS, newName string) (*bdpan.ManageFileRes, error) {
	return bdpan.RenameFiles(h.accessToken, bdpan.NewFileManager(pathS, "", newName))
}

// 批量重命名文件列表
func (h *FileHandler) BatchRenameFiles(files []*model.File) (*bdpan.ManageFileRes, error) {
	// 获取要修改的名字
	names := make([]string, 0)
	for _, file := range files {
		names = append(names, file.GetFilename())
	}

	newName, err := tools.EditTextInEditer("nvim", strings.Join(names, "\n"))
	if err != nil {
		logger.Errorf("读取修改后的名字失败: %v", err)
		return nil, err
	}
	newName = strings.Trim(newName, "\n")
	newNames := strings.Split(newName, "\n")
	if len(names) != len(newNames) {
		return nil, fmt.Errorf("名称不批量，修改失败: %s", newName)
	}

	reqManagers := make([]*bdpan.FileManager, 0)
	for i, f := range files {
		reqManagers = append(reqManagers, bdpan.NewFileManager(f.Path, "", newNames[i]))
	}
	return bdpan.RenameFiles(h.accessToken, reqManagers...)
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

// 上传文件夹
func (h *FileHandler) UploadDir(req *dto.UploadReq, fromDir, toDir string) error {
	logger.Printf("上传文件夹 %s => %s", fromDir, toDir)
	req.IsRewrite = true

	begin := time.Now()
	existFiles, err := bdtools.GetDirAllFiles(h.accessToken, toDir)
	if err != nil && err.Error() != bdpan.ErrFilenameNotFound.Error() {
		return err
	}
	logger.Infof("获取已存在文件耗时: %v", time.Since(begin))
	fsids := make([]uint64, 0)
	for _, f := range existFiles {
		fsids = append(fsids, f.FSID)
	}

	begin = time.Now()
	allBatchFiles, err := bdtools.BatchGetFileInfos(h.accessToken, fsids)
	if err != nil {
		return err
	}
	logger.Infof("获取已存在文件详情耗时: %v", time.Since(begin))
	logger.Printf("当前目录已存在文件数量: %d", len(allBatchFiles))

	// 将批量获取到的文件信息更新到 existFileMap 中
	existFileMap := make(map[string]*bdpan.FileInfo, 0)
	for _, f := range allBatchFiles {
		existFileMap[f.Path] = f
	}

	fromPaths := make([]any, 0)
	err = filepath.Walk(fromDir,
		func(pathStr string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// 不处理文件夹
			if info.IsDir() {
				return nil
			}
			fromPaths = append(fromPaths, pathStr)
			return nil
		})
	if err != nil {
		return err
	}

	tools.ExecLoop(fromPaths, len(fromPaths), func(total, index int, item any) error {
		fromPath := item.(string)
		// logger.Printf("[%4d/%d] %s", index, total, fromPath)
		toPath := path.Join(toDir, strings.ReplaceAll(fromPath, fromDir, ""))
		toFile := existFileMap[toPath]
		return h.UploadFile(
			req,
			fromPath,
			toPath,
			toFile,
			false,
			tools.Printf(logger.Infof),
		)
	},
		tools.Printf(logger.Infof),
		gotasker.NewBubblesProgressBar(),
	)

	return nil
}

// 上传文件
// 上传之前查看上传记录
//
//	如果有记录直接对比两次文件的远程 md5 是否相同
//	如果没有记录，则需要对比本地和远程记录，较慢
//
// req.IsRewrite = true 时，直接执行覆盖上传
// req.IsRewrite = false 时
//
//	如果本地文件和远程文件md5相同，打印信息直接返回
//	如果本地文件和远程文件md5不相同，则询问是否覆盖
//
// 上传成功后需要保存上传记录
func (h *FileHandler) UploadFile(
	req *dto.UploadReq,
	fromPath, toPath string,
	toFile *bdpan.FileInfo,
	printFile bool,
	args ...any,
) error {
	uPrintf := logger.Printf
	for _, arg := range args {
		switch val := arg.(type) {
		case tools.Printf:
			uPrintf = val
		}
	}

	uPrintf("上传文件 %s => %s", fromPath, toPath)
	fileMD5, err := tools.Md5File(fromPath)
	if err != nil {
		return err
	}
	logger.Infof("File MD5: %s", fileMD5)
	if toFile != nil {
		// 查找上传记录，直接比对 md5
		logger.Infof("通过文件 md5 查找上传记录: %s", fileMD5)
		existHistory := model.FindUploadHistoryByLocalMD5(fileMD5)
		if existHistory == nil {
			// 如果没有上传记录直接获取远程 md5 和本地进行对比
			logger.Infof("通过文件地址获取 md5: %s", toFile.Dlink)
			remoteMD5, err := bdtools.GetFileContentMD5(toFile)
			if err != nil {
				return err
			}
			logger.Infof("Remote File MD5: %s", remoteMD5)
			if fileMD5 == remoteMD5 {
				uPrintf("文件已存在: %s", toPath)
				return nil
			}
		} else {
			// 有上传记录使用两个远程对比
			logger.Infof("获取到上次上传记录 FSID: %d md5: %s", existHistory.FSID, existHistory.MD5)
			if existHistory.MD5 == toFile.MD5 {
				uPrintf("文件已存在: %s", toPath)
				return nil
			}

		}

		if !req.IsRewrite {
			var confirm bool
			err = huh.NewConfirm().
				Title("文件已存在，是否覆盖？").
				Affirmative("Yes!").
				Negative("No.").
				Value(&confirm).WithTheme(huh.ThemeCatppuccin()).Run()
			if err != nil {
				return err
			}
			if !confirm {
				uPrintf("取消上传: %s", toPath)
				return nil
			}
			req.IsRewrite = true
		}
	}
	createFileRes, err := bdtools.UploadFile(
		h.accessToken,
		fromPath,
		toPath,
		gotasker.NewBubblesProgressBar(),
		bdtools.Printf(logger.Infof),
		bdtools.IsRewrite(req.IsRewrite),
	)
	if err != nil {
		return err
	}
	uPrintf("上传文件成功")
	// 保存上传记录
	saveHistory := &model.UploadHistory{
		FSID:           createFileRes.FSId,
		Path:           createFileRes.Path,
		Size:           createFileRes.Size,
		Category:       createFileRes.Category,
		ServerFilename: createFileRes.ServerFilename,
		MD5:            createFileRes.Md5,
		LocalMD5:       fileMD5,
		CTime:          createFileRes.Ctime,
		MTime:          createFileRes.Mtime,
	}
	saveHistory.Init()
	model.Save(saveHistory)

	if printFile {
		file, err := bdtools.GetFileInfo(h.accessToken, createFileRes.FSId)
		if err != nil {
			return err
		}
		bdtools.PrintFileInfo(file)
	}
	return nil
}

func (h *FileHandler) CmdUpload(req *dto.UploadReq) error {
	fromPath := req.Local
	toPath := req.Path
	if tools.FileExists(fromPath) {
		// 上传文件
		toFile, _ := bdtools.GetFileByPath(h.accessToken, toPath)
		return h.UploadFile(req, fromPath, toPath, toFile, true)
	} else if tools.DirExists(fromPath) {
		// 上传文件夹
		return h.UploadDir(req, fromPath, toPath)
	} else {
		return fmt.Errorf("文件不存在: %s", fromPath)
	}
}

func (h *FileHandler) CmdBackup(req *dto.BackupReq) error {
	fromDir := req.Local
	if !tools.FileExists(fromDir) && !tools.DirExists(fromDir) {
		return fmt.Errorf("本地文件夹不存在: %s", fromDir)
	}
	if tools.FileExists(fromDir) {
		return errors.New("文件上传请直接调用 upload 命令")
	}

	backupName := time.Now().Format("2006-01-02-150405")
	backupDir := path.Join(req.Path, "Backups", backupName)
	uploadReq := dto.NewUploadReq()
	uploadReq.IsRewrite = true
	err := h.UploadDir(uploadReq, fromDir, backupDir)
	if err != nil {
		return err
	}
	logger.Printf("%s 已经成功备份到 %s 中", fromDir, backupDir)
	return nil
}

func FormatPath(path string) string {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}

func PrintFileInfo(file *bdpan.FileInfo) {
	logger.Printf("文件详情")
	bdtools.PrintFileInfo(file)
}
