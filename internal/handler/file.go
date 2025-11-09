package handler

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/wxnacy/bdpan-cli/internal/api"
	"github.com/wxnacy/bdpan-cli/internal/config"
	"github.com/wxnacy/bdpan-cli/internal/downloader"
	"github.com/wxnacy/bdpan-cli/internal/dto"
	"github.com/wxnacy/bdpan-cli/internal/logger"
	"github.com/wxnacy/bdpan-cli/internal/model"
	"github.com/wxnacy/bdpan-cli/internal/taskstore"
	"github.com/wxnacy/bdpan-cli/pkg/bdtools"
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
	logger.Printf("开始重命名，请稍后...")
	return bdpan.RenameFiles(h.accessToken, reqManagers...)
}

// 执行下载命令
//
// 参数：
// - req: 具体字段描述见 cmd/download.ga init() 中每个 cobra.Command 初始化 usage 字段
//
// 返回：
// - error: 错误信息
//
// 实现逻辑：
//
// 1. 根据路径查找文件信息
// 2. 判断是文件夹还是文件，分别调用对应方法
// 3. 如果是文件夹，调用 h.DownloadDir
// 4. 如果是文件，调用 h.DownloadFile
// 5. 输出下载结果
// 6. 意外失败，使用 logger.Errorf 写入日志，返回友好错误信息，提示 bdpan log 查看原因
func (h *FileHandler) CmdDownload(req *dto.DownloadReq) error {
	fmt.Printf("正在查找文件: %s\n", req.Path)

	// 1. 查找文件
	f, err := h.GetFileByPath(req.Path)
	if err != nil {
		return fmt.Errorf("查找文件失败: %w", err)
	}

	fmt.Printf("文件名: %s\n", f.GetFilename())
	fmt.Printf("文件ID: %d\n", f.FSID)
	fmt.Printf("文件大小: %s\n", tools.FormatSize(int64(f.Size)))

	// 2. 根据文件类型处理
	if f.IsDir() {
		// 下载文件夹
		fmt.Println("\n开始下载文件夹...")
		outputDir, err := h.DownloadDir(f, req)
		if err != nil {
			// 处理用户取消
			if errors.Is(err, context.Canceled) {
				fmt.Println("\n✗ 下载已取消")
				return nil
			}
			// 记录详细错误到日志，返回友好提示
			logger.Errorf("下载文件夹失败: %v", err)
			return fmt.Errorf("网络原因下载失败，请重试，具体报错请通过 bdpan log 命令查看")
		}
		if outputDir != "" {
			fmt.Printf("\n✓ 文件夹下载完成: %s\n", outputDir)
		}
	} else {
		// 下载单个文件
		fmt.Println("\n开始下载文件...")
		outputPath, err := h.DownloadFile(f, req)
		if err != nil {
			// 处理用户取消
			if errors.Is(err, context.Canceled) {
				fmt.Println("\n✗ 下载已取消")
				return nil
			}
			// 记录详细错误到日志，返回友好提示
			logger.Errorf("下载文件失败: %v", err)
			return fmt.Errorf("网络原因下载失败，请重试，具体报错请通过 bdpan log 命令查看")
		}
		if outputPath != "" {
			fmt.Printf("\n✓ 文件下载完成: %s\n", outputPath)
		}
	}

	return nil
}

// 下载文件夹
//
// 参数：
// - file: 查询到的百度文件夹实例
// - req: 具体字段描述见 cmd/download.ga init() 中每个 cobra.Command 初始化 usage 字段
//
// 返回：
// - string: 下载的文件夹路径
// - error: 错误信息
//
// 实现逻辑：
//
// 1. 获取文件夹下的所有文件列表（递归）
// 2. 使用 `bdtools.BatchGetFileInfos` 批量获取文件详情（包含下载链接）
// 3. 并发下载文件，默认并发数 3
// 4. 显示下载进度，统计已下载/总数量
func (h *FileHandler) DownloadDir(file *bdpan.FileInfo, req *dto.DownloadReq) (string, error) {
	// 确定下载目标目录
	_, dirName := filepath.Split(file.Path)
	outputDir := filepath.Join(req.OutputDir, dirName)

	fmt.Printf("开始下载文件夹: %s -> %s\n", file.Path, outputDir)

	// 1. 获取文件夹下的所有文件
	files, err := h.GetDirAllFiles(file.Path)
	if err != nil {
		return "", fmt.Errorf("获取文件列表失败: %w", err)
	}

	// 过滤出文件（排除文件夹）
	fileList := make([]*bdpan.FileInfo, 0)
	for _, f := range files {
		if !f.IsDir() {
			fileList = append(fileList, f)
		}
	}

	fmt.Printf("找到 %d 个文件\n", len(fileList))
	if len(fileList) == 0 {
		return outputDir, nil
	}

	// 2. 批量获取文件详情（获取 Dlink）
	fsids := make([]uint64, 0, len(fileList))
	for _, f := range fileList {
		fsids = append(fsids, f.FSID)
	}

	fmt.Println("正在获取文件下载链接...")
	detailFiles, err := bdtools.BatchGetFileInfos(h.accessToken, fsids)
	if err != nil {
		return "", fmt.Errorf("获取文件详情失败: %w", err)
	}

	// 计算总字节数，用于聚合进度条
	var totalBytes int64
	for _, f := range detailFiles {
		totalBytes += int64(f.Size)
	}

	// ===== Task detection & claim =====
	identity := taskstore.BuildIdentitySHA1("download", "dir", file.Path, outputDir)
	tdata := taskstore.DownloadData{Path: file.Path, OutputDir: outputDir, IsDir: true}
	taskID, attached, err := taskstore.ClaimOrCreate(context.Background(), taskstore.TaskTypeDownload, identity, "", totalBytes, tdata)
	if err != nil {
		return "", err
	}
	if attached {
		fmt.Printf("已有下载任务正在运行: %s\n使用: bdpan task status %s 查看进度\n", taskID, taskID)
		return "", nil
	}

	// 3. 并发下载文件
	var (
		wg               sync.WaitGroup
		concurrency      = 3
		sem              = make(chan struct{}, concurrency)
		errChan          = make(chan error, len(detailFiles))
		successCount     int
		failedCount      int
		mu               sync.Mutex
		globalMu         sync.Mutex
		globalDownloaded int64
		progWriter       *downloader.ProgressWriter
		activeMu         sync.Mutex
		activeSet        = make(map[string]struct{})
	)

	buildActiveStatus := func() string {
		activeMu.Lock()
		defer activeMu.Unlock()
		if len(activeSet) == 0 {
			return ""
		}
		names := make([]string, 0, 3)
		for name := range activeSet {
			names = append(names, name)
			if len(names) >= 3 {
				break
			}
		}
		return "正在下载: " + strings.Join(names, " | ")
	}

	// 父级上下文与取消，统一控制所有任务
	parentCtx, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()

	// 捕获 Ctrl+C，统一取消
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	canceledOnce := make(chan struct{}, 1)
	go func() {
		<-sigCh
		select {
		case canceledOnce <- struct{}{}:
			parentCancel()
		default:
		}
	}()

	// 启动聚合 TUI 进度条（按 q 或 Ctrl+C 取消所有）
	if totalBytes > 0 {
		// 仅传入文件夹名，避免标题出现重复的“下载:”前缀
		model := downloader.NewProgressModel(filepath.Base(file.Path), totalBytes, parentCancel)
		p := tea.NewProgram(model)
		progWriter = downloader.NewProgressWriter(p, totalBytes)
		go func() {
			if _, err := p.Run(); err != nil {
				logger.Errorf("目录进度条运行错误: %v", err)
			}
		}()
	}

	fmt.Printf("开始并发下载（并发数: %d）\n", concurrency)

	// ===== 5s Heartbeat to DB =====
	hbTicker := time.NewTicker(5 * time.Second)
	hbQuit := make(chan struct{})
	go func() {
		defer hbTicker.Stop()
		var last int64
		for {
			select {
			case <-hbQuit:
				return
			case <-parentCtx.Done():
				return
			case <-hbTicker.C:
				globalMu.Lock()
				cur := globalDownloaded
				globalMu.Unlock()
				delta := cur - last
				if delta < 0 {
					delta = 0
				}
				last = cur
				speed := delta / 5
				prog := 0.0
				if totalBytes > 0 {
					prog = float64(cur) / float64(totalBytes)
				}
				cancelReq, _ := taskstore.Heartbeat(context.Background(), taskID, taskstore.HeartbeatData{
					DownloadedBytes: cur,
					TotalBytes:      totalBytes,
					Progress:        prog,
					SpeedBPS:        speed,
					ETASeconds:      0,
				})
				if cancelReq {
					parentCancel()
				}
			}
		}
	}()

	for _, f := range detailFiles {
		wg.Add(1)
		go func(fileInfo *bdpan.FileInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// 计算目标文件路径（保持相对路径结构）
			relPath := strings.TrimPrefix(fileInfo.Path, file.Path)
			relPath = strings.TrimPrefix(relPath, "/")
			targetPath := filepath.Join(outputDir, relPath)

			// 检查文件是否已存在
			if _, err := os.Stat(targetPath); err == nil {
				if progWriter != nil {
					progWriter.UpdateStatus("✓ 文件已存在，跳过: " + relPath)
				} else {
					fmt.Printf("✓ 文件已存在，跳过: %s\n", relPath)
				}
				mu.Lock()
				successCount++
				mu.Unlock()
				return
			}

			// 使用无 TUI 的单文件下载，继承父级上下文，统一取消
			if progWriter != nil {
				activeMu.Lock()
				activeSet[relPath] = struct{}{}
				activeMu.Unlock()
				progWriter.UpdateStatus(buildActiveStatus())
			} else {
				fmt.Printf("↓ 下载: %s\n", relPath)
			}
			if err := h.downloadSingleNoTUIWithAgg(parentCtx, fileInfo, targetPath, req.IsSync, progWriter, &globalDownloaded, &globalMu, totalBytes); err != nil {
				// 用户取消不重复打印
				if errors.Is(err, context.Canceled) {
					if progWriter != nil {
						activeMu.Lock()
						delete(activeSet, relPath)
						activeMu.Unlock()
						progWriter.UpdateStatus(buildActiveStatus())
					}
					errChan <- err
					return
				}
				errChan <- fmt.Errorf("下载 %s 失败: %w", fileInfo.Path, err)
				mu.Lock()
				failedCount++
				mu.Unlock()
				if progWriter != nil {
					activeMu.Lock()
					delete(activeSet, relPath)
					activeMu.Unlock()
					progWriter.UpdateStatus(buildActiveStatus())
				} else {
					fmt.Printf("✗ 下载失败: %s - %v\n", relPath, err)
				}
			} else {
				mu.Lock()
				successCount++
				mu.Unlock()
				if progWriter != nil {
					activeMu.Lock()
					delete(activeSet, relPath)
					activeMu.Unlock()
					progWriter.UpdateStatus(buildActiveStatus())
				} else {
					fmt.Printf("✓ 下载完成: %s\n", relPath)
				}
			}
		}(f)
	}

	wg.Wait()
	close(errChan)

	// 统计结果
	fmt.Println("\n================================")
	fmt.Printf("下载完成！\n")
	fmt.Printf("总数: %d, 成功: %d, 失败: %d\n", len(detailFiles), successCount, failedCount)
	fmt.Printf("保存目录: %s\n", outputDir)
	fmt.Println("================================")

	// 结束聚合进度条
	if progWriter != nil {
		if successCount == len(detailFiles) {
			progWriter.Complete()
		}
		// 给 UI 一点渲染时间
		time.Sleep(100 * time.Millisecond)
	}

	// 如果因取消导致的错误，统一返回 context.Canceled
	for e := range errChan {
		if errors.Is(e, context.Canceled) {
			if progWriter != nil {
				progWriter.Error(context.Canceled)
				time.Sleep(100 * time.Millisecond)
			}
			close(hbQuit)
			_ = taskstore.SetCanceled(context.Background(), taskID)
			return outputDir, context.Canceled
		}
		// 返回第一个非取消错误
		if progWriter != nil {
			progWriter.Error(e)
			time.Sleep(100 * time.Millisecond)
		}
		close(hbQuit)
		_ = taskstore.Fail(context.Background(), taskID, e.Error())
		return outputDir, e
	}

	close(hbQuit)
	_ = taskstore.Complete(context.Background(), taskID)
	return outputDir, nil
}

// downloadSingleNoTUI 目录下载场景下的单文件下载（无 TUI），可继承外部上下文
func (h *FileHandler) downloadSingleNoTUI(ctx context.Context, file *bdpan.FileInfo, targetPath string, isSync bool) error {
	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}

	// 创建缓存目录
	cacheDir := filepath.Join(config.GetCacheDir(), file.MD5)

	// 4. 创建分片下载器
	d := downloader.NewChunkDownloader(file.Dlink, targetPath, cacheDir)

	// 5. 设置并发数
	if isSync {
		d.SetConcurrency(1)
	} else {
		d.SetConcurrency(3)
	}

	// 无 TUI，不设置进度回调

	// 7. 开始下载
	if err := d.Start(); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	return nil
}

func (h *FileHandler) downloadSingleNoTUIWithAgg(
	ctx context.Context,
	file *bdpan.FileInfo,
	targetPath string,
	isSync bool,
	progWriter *downloader.ProgressWriter,
	globalDownloaded *int64,
	globalMu *sync.Mutex,
	totalBytes int64,
) error {
	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}
	// 创建缓存目录
	cacheDir := filepath.Join(config.GetCacheDir(), file.MD5)

	d := downloader.NewChunkDownloader(file.Dlink, targetPath, cacheDir)
	if ctx != nil {
		d.SetContext(ctx, nil)
	}
	if isSync {
		d.SetConcurrency(1)
	} else {
		d.SetConcurrency(3)
	}

	// 将该文件的进度累计到全局
	var last int64
	d.SetProgressFunc(func(downloaded, _ int64) {
		if progWriter == nil {
			return
		}
		delta := downloaded - last
		if delta <= 0 {
			return
		}
		last = downloaded
		globalMu.Lock()
		*globalDownloaded += delta
		current := *globalDownloaded
		globalMu.Unlock()
		progWriter.UpdateProgress(current, totalBytes)
	})

	return d.Start()
}

// 分片断点下载文件
//
// 参数：
// - file: 查询到的百度文件实例
// - req: 具体字段描述见 cmd/download.ga init() 中每个 cobra.Command 初始化 usage 字段
//
// 返回：
// - string: 下载的文件路径
// - error: 错误信息
//
// 实现逻辑:
//
// 1. 确定输出文件路径（优先级: OutputPath > OutputDir + filename）
// 2. 处理文件名冲突（数字后缀递增）
// 3. 创建缓存目录（使用 config.GetCacheDir() + file.MD5），下载完成后缓存目录需要一并删除
// 4. 创建分片下载器，设置分片大小为 5MB
// 5. 设置并发数（同步模式为 1，异步模式为 4）
// 6. 设置进度回调函数，显示下载进度
// 7. 开始下载，支持断点续传
// 8. 进度条样式使用 https://github.com/charmbracelet/bubbletea/tree/main/examples/progress-download
// 9. 使用file.Dlink时，必须在请求header中设置User-Agent字段为pan.baidu.com
// 10. 进度条与上下文字必须左对齐，禁止在信息行和提示行前添加前导空格
func (h *FileHandler) DownloadFile(file *bdpan.FileInfo, req *dto.DownloadReq) (string, error) {
	// 1. 确定输出文件路径
	var outputPath string
	if req.OutputPath != "" {
		outputPath = req.OutputPath
	} else {
		_, filename := filepath.Split(file.Path)
		outputPath = filepath.Join(req.OutputDir, filename)
	}

	// 2. 处理文件名冲突
	outputPath = h.resolveOutputPath(outputPath)

	// 3. 创建缓存目录
	cacheDir := filepath.Join(config.GetCacheDir(), file.MD5)

	// ===== Task detection & claim =====
	identity := taskstore.BuildIdentitySHA1("download", "file", file.Path, outputPath)
	tdata := taskstore.DownloadData{FSID: file.FSID, Path: file.Path, MD5: file.MD5, TargetPath: outputPath, OutputDir: req.OutputDir, IsDir: false}
	taskID, attached, err := taskstore.ClaimOrCreate(context.Background(), taskstore.TaskTypeDownload, identity, "", int64(file.Size), tdata)
	if err != nil {
		return "", err
	}
	if attached {
		fmt.Printf("已有下载任务正在运行: %s\n使用: bdpan task status %s 查看进度\n", taskID, taskID)
		return "", nil
	}

	// 4. 创建分片下载器
	d := downloader.NewChunkDownloader(file.Dlink, outputPath, cacheDir)

	// 5. 设置并发数
	if req.IsSync {
		d.SetConcurrency(1)
	} else {
		d.SetConcurrency(4)
	}

	// 6. 进度回调 + 启用 TUI
	var singleDownloaded int64
	d.SetProgressFunc(func(downloaded, _ int64) {
		singleDownloaded = downloaded
	})
	_, filename := filepath.Split(file.Path)
	d.EnableTUI(filename)

	// 7. Heartbeat 5s + cancel check
	ctx, cancel := context.WithCancel(context.Background())
	d.SetContext(ctx, cancel)
	hbTicker := time.NewTicker(5 * time.Second)
	defer hbTicker.Stop()
	go func() {
		var last int64
		for {
			select {
			case <-ctx.Done():
				return
			case <-hbTicker.C:
				cur := singleDownloaded
				delta := cur - last
				if delta < 0 {
					delta = 0
				}
				last = cur
				speed := delta / 5
				prog := 0.0
				if file.Size > 0 {
					prog = float64(cur) / float64(file.Size)
				}
				cancelReq, _ := taskstore.Heartbeat(context.Background(), taskID, taskstore.HeartbeatData{
					DownloadedBytes: cur,
					TotalBytes:      int64(file.Size),
					Progress:        prog,
					SpeedBPS:        speed,
					ETASeconds:      0,
				})
				if cancelReq {
					cancel()
				}
			}
		}
	}()

	// 8. 开始下载
	if err := d.Start(); err != nil {
		_ = taskstore.Fail(context.Background(), taskID, err.Error())
		return "", fmt.Errorf("下载失败: %w", err)
	}

	_ = taskstore.Complete(context.Background(), taskID)
	return outputPath, nil
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
		FSID:           createFileRes.FSID,
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
		file, err := bdtools.GetFileInfo(h.accessToken, createFileRes.FSID)
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

// resolveOutputPath 解决文件路径冲突
//
// 如果文件已存在，自动添加数字后缀
// 例如: file.txt -> file(1).txt -> file(2).txt
func (h *FileHandler) resolveOutputPath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	// 文件已存在，添加数字后缀
	ext := filepath.Ext(path)
	name := strings.TrimSuffix(path, ext)

	for i := 1; ; i++ {
		newPath := fmt.Sprintf("%s(%d)%s", name, i, ext)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}
