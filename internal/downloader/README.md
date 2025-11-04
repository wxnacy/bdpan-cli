# 下载器模块

## 功能特性

### 1. 分片下载器 (ChunkDownloader)

支持断点续传的分片下载器，具有以下特性：

- ✅ **分片下载**：默认 5MB 分片大小，可配置
- ✅ **断点续传**：网络中断后可继续下载
- ✅ **并发控制**：支持多线程并发下载
- ✅ **进度显示**：两种进度显示模式
  - 简单模式：回调函数方式
  - TUI 模式：使用 bubbletea 的漂亮彩色进度条

### 2. TUI 进度条 (ProgressModel)

基于 charmbracelet/bubbletea 和 charmbracelet/bubbles 实现的漂亮进度条：

**显示内容**：
- 文件名（蓝色加粗）
- 彩色渐变进度条
- 下载信息：已下载/总大小、百分比
- 下载速度（实时计算）
- 预计剩余时间（ETA）
- 操作提示（按 q 或 Ctrl+C 取消）

**交互操作**：
- 按 `q` 或 `Ctrl+C` 立即取消下载
- 取消后会显示友好的提示：`✗ 下载已取消`
- 已下载的分片会保留，支持下次断点续传
- 无需按两次，立即响应

**样式特点**：
- 使用 lipgloss 样式库
- 彩色渐变进度条
- 清晰的信息层次
- 支持终端大小自适应

## 使用示例

### 基础用法

```go
import "github.com/wxnacy/bdpan-cli/internal/downloader"

// 创建下载器
d := downloader.NewChunkDownloader(
    "https://example.com/file.zip",  // 下载 URL
    "/path/to/output/file.zip",      // 输出路径
    "/path/to/cache/dir",            // 缓存目录
)

// 设置分片大小（可选，默认 5MB）
d.SetChunkSize(10 * 1024 * 1024) // 10MB

// 设置并发数（可选，默认 3）
d.SetConcurrency(5)

// 启用 TUI 进度条（推荐）
d.EnableTUI("file.zip")

// 开始下载
if err := d.Start(); err != nil {
    log.Fatal(err)
}
```

### 使用简单进度回调

```go
// 不使用 TUI，使用简单的回调函数
d.SetProgressFunc(func(downloaded, total int64) {
    percent := float64(downloaded) / float64(total) * 100
    fmt.Printf("\r%.2f%%", percent)
})

d.Start()
```

## 进度条效果预览

```
下载: 无欲则刚.mp4

████████████████████████████████░░░░░░░░

  45.2 MB / 100.0 MB  45.2%  速度: 2.3 MB/s  剩余: 00:23

  按 q 或 Ctrl+C 取消
```

## 依赖

- `github.com/charmbracelet/bubbletea` - TUI 框架
- `github.com/charmbracelet/bubbles` - TUI 组件库
- `github.com/charmbracelet/lipgloss` - 样式库

安装依赖：
```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/bubbles
go get github.com/charmbracelet/lipgloss
```

或运行：
```bash
go mod tidy
```

## 技术细节

### 断点续传实现

1. 将文件分成多个分片
2. 每个分片独立下载到缓存目录
3. 下载前检查缓存目录中已下载的分片
4. 只下载未完成的分片
5. 下载完成后合并所有分片
6. 清理缓存文件

### 缓存文件结构

```
cache_dir/
  ├── chunk_0    # 第 0 个分片
  ├── chunk_1    # 第 1 个分片
  ├── chunk_2    # 第 2 个分片
  └── ...
```

### 并发控制

使用信号量（semaphore）控制并发数：
- 文件内并发：多个分片同时下载
- 最大并发数可配置（默认 3）
- 避免过多连接导致的资源浪费

## 性能优化

1. **分片大小**：默认 5MB，适合大多数场景
2. **并发数**：默认 3，平衡速度和资源消耗
3. **缓冲区**：32KB 读写缓冲，提高 I/O 效率
4. **进度更新**：使用 channel 异步更新，不阻塞下载
