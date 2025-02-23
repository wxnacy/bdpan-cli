package common

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// 直接下载小文件的函数
func DownloadFile(url, outputPath string) error {
	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("User-Agent", "pan.baidu.com") // 设置 User-Agent

	// 创建 HTTP 客户端并处理 302 跳转
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			req.Header.Set("User-Agent", "pan.baidu.com") // 保持 User-Agent 在跳转中不变
			return nil
		},
	}

	// 发起请求
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file, status code: %d", resp.StatusCode)
	}

	// 创建输出文件
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	// 将响应体的内容写入文件
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	return nil
}
