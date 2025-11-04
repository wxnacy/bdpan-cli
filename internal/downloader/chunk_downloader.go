package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wxnacy/bdpan-cli/internal/logger"
)

const (
	// DefaultChunkSize 默认分片大小 - 5MB
	// 这是平衡性能、稳定性和重传成本的最佳选择
	DefaultChunkSize = 5 * 1024 * 1024

	// MinChunkSize 最小分片大小 - 1MB
	MinChunkSize = 1 * 1024 * 1024

	// MaxChunkSize 最大分片大小 - 50MB
	MaxChunkSize = 50 * 1024 * 1024

	// DefaultConcurrency 默认并发数
	DefaultConcurrency = 3

	// userAgent 统一的 User-Agent 字符串
	userAgent = "pan.baidu.com"
)

// ChunkInfo 分片信息
type ChunkInfo struct {
	Index      int   // 分片索引（从0开始）
	Start      int64 // 起始位置
	End        int64 // 结束位置
	Size       int64 // 分片大小
	Downloaded int64 // 已下载大小
	Completed  bool  // 是否完成
}

// ChunkDownloader 分片下载器
//
// 功能：
// - 支持多线程分片下载
// - 支持断点续传
// - 支持进度回调
//
// 实现细节：
// - 使用 HTTP Range 头实现分片下载
// - 将分片缓存保存到临时目录
// - 下载完成后合并所有分片到目标文件
// - 断点续传通过检查缓存文件的大小实现
type ChunkDownloader struct {
	URL           string           // 下载 URL
	OutputPath    string           // 输出文件路径
	CacheDir      string           // 缓存目录
	ChunkSize     int64            // 分片大小
	Concurrency   int              // 并发数
	TotalSize     int64            // 文件总大小
	Chunks        []*ChunkInfo     // 分片列表
	ProgressFunc  func(int64, int64) // 进度回调函数 (已下载, 总大小)
	useTUI        bool             // 是否启用 TUI 进度条
	Filename      string           // 文件名（用于显示）
	client        *http.Client     // HTTP 客户端
	ctx           context.Context  // 上下文
	cancel        context.CancelFunc // 取消函数
	mu            sync.Mutex       // 互斥锁
	downloadedSum int64            // 已下载总量
	progressWriter *ProgressWriter // 进度写入器
}

// NewChunkDownloader 创建分片下载器
func NewChunkDownloader(url, outputPath, cacheDir string) *ChunkDownloader {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &ChunkDownloader{
		URL:         url,
		OutputPath:  outputPath,
		CacheDir:    cacheDir,
		ChunkSize:   DefaultChunkSize,
		Concurrency: DefaultConcurrency,
		client: &http.Client{
			Timeout: 30 * time.Minute,
			Transport: &http.Transport{
				// 关闭 keep-alive，避免空闲连接上服务端的额外输出触发标准库日志
				DisableKeepAlives: true,
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// 保持 User-Agent 在跳转中不变
				req.Header.Set("User-Agent", userAgent)
				// 保证重定向请求也不复用连接
				req.Close = true
				req.Header.Set("Connection", "close")
				return nil
			},
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// SetChunkSize 设置分片大小
func (d *ChunkDownloader) SetChunkSize(size int64) *ChunkDownloader {
	if size < MinChunkSize {
		size = MinChunkSize
	}
	if size > MaxChunkSize {
		size = MaxChunkSize
	}
	d.ChunkSize = size
	return d
}

// SetConcurrency 设置并发数
func (d *ChunkDownloader) SetConcurrency(n int) *ChunkDownloader {
	if n < 1 {
		n = 1
	}
	if n > 10 {
		n = 10
	}
	d.Concurrency = n
	return d
}

// SetProgressFunc 设置进度回调函数
func (d *ChunkDownloader) SetProgressFunc(f func(int64, int64)) *ChunkDownloader {
	d.ProgressFunc = f
	return d
}

// EnableTUI 启用 TUI 进度条
func (d *ChunkDownloader) EnableTUI(filename string) *ChunkDownloader {
	d.useTUI = true
	d.Filename = filename
	return d
}

// Start 开始下载
//
// 实现步骤：
// 1. 获取文件总大小，确定是否支持 Range 下载
// 2. 计算分片信息
// 3. 如果启用 TUI，创建进度条程序
// 4. 检查断点续传：扫描缓存目录，恢复已下载的分片进度
// 5. 并发下载未完成的分片
// 6. 合并所有分片到目标文件
// 7. 清理缓存文件
func (d *ChunkDownloader) Start() error {
	// 1. 获取文件总大小
	totalSize, supportsRange, err := d.getFileSize()
	if err != nil {
		return fmt.Errorf("获取文件大小失败: %w", err)
	}
	d.TotalSize = totalSize
	logger.Infof("文件总大小: %d bytes (%.2f MB)", totalSize, float64(totalSize)/1024/1024)

	// 如果不支持 Range 或文件很小，直接下载
	if !supportsRange || totalSize < d.ChunkSize {
		logger.Infof("不支持分片下载或文件过小，使用直接下载")
		return d.downloadDirect()
	}

	// 2. 计算分片
	d.calculateChunks()
	logger.Infof("分片数量: %d, 每片大小: %.2f MB", len(d.Chunks), float64(d.ChunkSize)/1024/1024)

	// 3. 创建缓存目录
	if err := os.MkdirAll(d.CacheDir, 0755); err != nil {
		return fmt.Errorf("创建缓存目录失败: %w", err)
	}

	// 4. 如果启用 TUI，创建进度条程序
	var p *tea.Program
	if d.useTUI {
		filename := d.Filename
		if filename == "" {
			filename = filepath.Base(d.OutputPath)
		}
		model := NewProgressModel(filename, totalSize, d.cancel)
		p = tea.NewProgram(model)
		d.progressWriter = NewProgressWriter(p, totalSize)
		
		// 在 goroutine 中运行 TUI
		go func() {
			if _, err := p.Run(); err != nil {
				logger.Errorf("进度条运行错误: %v", err)
			}
		}()
	}

	// 5. 检查断点续传
	d.checkResume()

	// 6. 并发下载分片
	err = d.downloadChunks()
	if err != nil {
        // 对于用户取消，不向 TUI 输出错误，直接返回
        if err == context.Canceled {
            return err
        }
        if d.progressWriter != nil {
            d.progressWriter.Error(err)
            time.Sleep(200 * time.Millisecond) // 等待 UI 显示错误消息
        }
        return err
    }

	// 7. 合并分片
	logger.Infof("开始合并分片...")
	if err := d.mergeChunks(); err != nil {
		if d.progressWriter != nil {
			d.progressWriter.Error(err)
		}
		return fmt.Errorf("合并分片失败: %w", err)
	}

	// 8. 清理缓存
	d.cleanup()
	
	// 9. 通知完成
	if d.progressWriter != nil {
		d.progressWriter.Complete()
		time.Sleep(100 * time.Millisecond) // 等待 UI 更新
	}
	
	logger.Infof("下载完成: %s", d.OutputPath)
	return nil
}

// Cancel 取消下载
func (d *ChunkDownloader) Cancel() {
	d.cancel()
}

// SetContext 允许外部覆盖上下文与取消函数，用于批量任务统一取消
func (d *ChunkDownloader) SetContext(ctx context.Context, cancel context.CancelFunc) *ChunkDownloader {
    if ctx != nil {
        d.ctx = ctx
    }
    if cancel != nil {
        d.cancel = cancel
    }
    return d
}

// getFileSize 获取文件大小并检查是否支持 Range 下载
func (d *ChunkDownloader) getFileSize() (int64, bool, error) {
	req, err := http.NewRequestWithContext(d.ctx, "HEAD", d.URL, nil)
	if err != nil {
		return 0, false, err
	}
	req.Header.Set("User-Agent", userAgent)
	// 使用短连接，避免空闲连接引发标准库日志
	req.Close = true
	req.Header.Set("Connection", "close")

	resp, err := d.client.Do(req)
	if err != nil {
		return 0, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, false, fmt.Errorf("HTTP status: %d", resp.StatusCode)
	}

	size := resp.ContentLength
	supportsRange := resp.Header.Get("Accept-Ranges") == "bytes"
	
	return size, supportsRange, nil
}

// calculateChunks 计算分片信息
func (d *ChunkDownloader) calculateChunks() {
	totalChunks := int((d.TotalSize + d.ChunkSize - 1) / d.ChunkSize)
	d.Chunks = make([]*ChunkInfo, totalChunks)

	for i := 0; i < totalChunks; i++ {
		start := int64(i) * d.ChunkSize
		end := start + d.ChunkSize - 1
		if end >= d.TotalSize {
			end = d.TotalSize - 1
		}

		d.Chunks[i] = &ChunkInfo{
			Index: i,
			Start: start,
			End:   end,
			Size:  end - start + 1,
		}
	}
}

// checkResume 检查断点续传
func (d *ChunkDownloader) checkResume() {
	for _, chunk := range d.Chunks {
		chunkPath := d.getChunkPath(chunk.Index)
		if info, err := os.Stat(chunkPath); err == nil {
			// 文件存在，检查大小
			if info.Size() == chunk.Size {
				chunk.Completed = true
				chunk.Downloaded = chunk.Size
				d.updateProgress(chunk.Size)
				logger.Infof("分片 %d 已完成，跳过下载", chunk.Index)
			} else if info.Size() > 0 {
				// 部分下载，记录已下载大小
				chunk.Downloaded = info.Size()
				d.updateProgress(info.Size())
				logger.Infof("分片 %d 已下载 %.2f%%", chunk.Index, float64(info.Size())/float64(chunk.Size)*100)
			}
		}
	}
}

// downloadChunks 并发下载所有未完成的分片
func (d *ChunkDownloader) downloadChunks() error {
	var wg sync.WaitGroup
	sem := make(chan struct{}, d.Concurrency) // 限制并发数
	errChan := make(chan error, len(d.Chunks))

	for _, chunk := range d.Chunks {
		if chunk.Completed {
			continue
		}

		wg.Add(1)
		go func(c *ChunkInfo) {
			defer wg.Done()
			sem <- struct{}{}        // 获取信号量
			defer func() { <-sem }() // 释放信号量

			if err := d.downloadChunk(c); err != nil {
				errChan <- fmt.Errorf("分片 %d 下载失败: %w", c.Index, err)
			}
		}(chunk)
	}

	wg.Wait()
	close(errChan)

	// 检查是否有错误
	if len(errChan) > 0 {
		return <-errChan
	}

	return nil
}

// downloadChunk 下载单个分片
func (d *ChunkDownloader) downloadChunk(chunk *ChunkInfo) error {
	chunkPath := d.getChunkPath(chunk.Index)
	
	// 计算实际的下载范围（考虑已下载部分）
	start := chunk.Start + chunk.Downloaded
	end := chunk.End

	// 如果已经下载完成，跳过
	if start > end {
		chunk.Completed = true
		return nil
	}

	// 创建请求
	req, err := http.NewRequestWithContext(d.ctx, "GET", d.URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	// 使用短连接
	req.Close = true
	req.Header.Set("Connection", "close")

	// 发起请求
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status: %d", resp.StatusCode)
	}

	// 打开文件（追加模式）
	file, err := os.OpenFile(chunkPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// 写入数据并更新进度
	buf := make([]byte, 32*1024) // 32KB 缓冲区
	for {
		select {
		case <-d.ctx.Done():
			return d.ctx.Err()
		default:
		}

		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := file.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			chunk.Downloaded += int64(n)
			d.updateProgress(int64(n))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	chunk.Completed = true
	return nil
}

// downloadDirect 直接下载（不分片）
func (d *ChunkDownloader) downloadDirect() error {
	req, err := http.NewRequestWithContext(d.ctx, "GET", d.URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	// 使用短连接
	req.Close = true
	req.Header.Set("Connection", "close")

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status: %d", resp.StatusCode)
	}

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(d.OutputPath), 0755); err != nil {
		return err
	}

	file, err := os.Create(d.OutputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 写入数据并更新进度
	buf := make([]byte, 32*1024)
	for {
		select {
		case <-d.ctx.Done():
			return d.ctx.Err()
		default:
		}

		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := file.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			d.updateProgress(int64(n))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// mergeChunks 合并所有分片到目标文件
func (d *ChunkDownloader) mergeChunks() error {
	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(d.OutputPath), 0755); err != nil {
		return err
	}

	// 创建输出文件
	outFile, err := os.Create(d.OutputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// 按顺序读取并写入每个分片
	for _, chunk := range d.Chunks {
		chunkPath := d.getChunkPath(chunk.Index)
		
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return fmt.Errorf("打开分片 %d 失败: %w", chunk.Index, err)
		}

		if _, err := io.Copy(outFile, chunkFile); err != nil {
			chunkFile.Close()
			return fmt.Errorf("写入分片 %d 失败: %w", chunk.Index, err)
		}
		chunkFile.Close()
	}

	return nil
}

// cleanup 清理缓存文件和缓存目录
//
// 下载完成后需要删除所有分片文件和整个缓存目录，释放磁盘空间
func (d *ChunkDownloader) cleanup() {
	// 删除所有分片文件
	for _, chunk := range d.Chunks {
		chunkPath := d.getChunkPath(chunk.Index)
		if err := os.Remove(chunkPath); err != nil {
			logger.Infof("删除分片文件失败 %s: %v", chunkPath, err)
		}
	}
	
	// 删除整个缓存目录
	if err := os.RemoveAll(d.CacheDir); err != nil {
		logger.Infof("删除缓存目录失败 %s: %v", d.CacheDir, err)
	} else {
		logger.Infof("缓存目录已清理: %s", d.CacheDir)
	}
}

// getChunkPath 获取分片文件路径
func (d *ChunkDownloader) getChunkPath(index int) string {
	return filepath.Join(d.CacheDir, fmt.Sprintf("chunk_%d", index))
}

// updateProgress 更新进度
func (d *ChunkDownloader) updateProgress(delta int64) {
	d.mu.Lock()
	d.downloadedSum += delta
	downloaded := d.downloadedSum
	d.mu.Unlock()

	// 优先使用 TUI 进度条
	if d.progressWriter != nil {
		d.progressWriter.UpdateProgress(downloaded, d.TotalSize)
	} else if d.ProgressFunc != nil {
		d.ProgressFunc(downloaded, d.TotalSize)
	}
}
