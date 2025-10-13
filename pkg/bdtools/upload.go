package bdtools

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/wxnacy/go-bdpan"
	"github.com/wxnacy/go-tools"
)

const (
	// 分片大小，4MB，普通用户的最大分片大小
	ChunkSize = 4 * 1024 * 1024
)

type (
	Printf    func(format string, v ...any)
	IsRewrite bool
)

// uploadFile 实现文件上传的完整流程
func UploadFile(accessToken, localFilePath, remoteFilePath string, args ...any) (*bdpan.CreateFileRes, error) {
	uPrintf := log.Printf
	var progressBar tools.ProgressBar
	var isRewrite IsRewrite

	for _, arg := range args {
		switch val := arg.(type) {
		case tools.ProgressBar:
			progressBar = val
		case Printf:
			uPrintf = val
		case IsRewrite:
			isRewrite = val
		}
	}

	// 1. 打开本地文件
	file, err := os.Open(localFilePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 2. 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 3. 计算文件大小
	fileSize := fileInfo.Size()

	// 4. 计算文件的MD5分块列表
	blockList, err := calculateBlockList(file, fileSize)
	if err != nil {
		return nil, fmt.Errorf("计算文件MD5失败: %w", err)
	}

	// 5. 预上传
	preCreateReq := bdpan.NewPreCreateFileReq(remoteFilePath, int32(fileSize), blockList)
	if isRewrite {
		preCreateReq.SetRtype(3)
	}
	uPrintf("预上传参数 %#v", preCreateReq)
	preCreateRes, err := bdpan.PreCreateFile(accessToken, preCreateReq)
	if err != nil {
		return nil, fmt.Errorf("预上传失败: %w", err)
	}

	if preCreateRes.IsError() {
		return nil, fmt.Errorf("预上传失败，%s", preCreateRes.Error())
	}

	uPrintf("预上传成功，uploadid: %s", preCreateRes.Uploadid)

	// 6. 分片上传
	// 如果文件小于等于4MB，只需要上传一个分片
	// 如果文件大于4MB，需要按照4MB大小分片上传
	var remoteBlockList []string

	if fileSize <= ChunkSize {
		// 小文件上传
		if _, err := file.Seek(0, 0); err != nil {
			return nil, fmt.Errorf("文件指针重置失败: %w", err)
		}

		uploadPartReq := bdpan.NewUploadFilePartReq(remoteFilePath, preCreateRes.Uploadid, 0)
		uploadPartReq.File = file

		uploadPartRes, err := bdpan.UploadFilePart(accessToken, uploadPartReq)
		if err != nil {
			return nil, fmt.Errorf("分片上传失败: %w", err)
		}

		if uploadPartRes.IsError() {
			return nil, fmt.Errorf("分片 %d 上传失败，%s", 0, uploadPartRes.Error())
		}

		remoteBlockList = append(remoteBlockList, uploadPartRes.Md5)
		uPrintf("分片 0 上传成功，md5: %s", uploadPartRes.Md5)
	} else {
		// 大文件分片上传
		chunkCount := int(fileSize / ChunkSize)
		if fileSize%ChunkSize != 0 {
			chunkCount++
		}
		if progressBar != nil {
			progressBar.Start(chunkCount)
		}

		for i := range chunkCount {
			if _, err := file.Seek(int64(i)*ChunkSize, 0); err != nil {
				return nil, fmt.Errorf("文件指针定位失败: %w", err)
			}

			// 创建临时文件存储分片数据
			tempFile, err := os.CreateTemp("", "upload_chunk_*")
			if err != nil {
				return nil, fmt.Errorf("创建临时文件失败: %w", err)
			}
			tempFilePath := tempFile.Name()
			defer os.Remove(tempFilePath)

			// 读取分片数据
			writer := bufio.NewWriter(tempFile)
			r := io.LimitReader(file, ChunkSize)
			if _, err := io.Copy(writer, r); err != nil {
				tempFile.Close()
				return nil, fmt.Errorf("写入分片数据失败: %w", err)
			}
			writer.Flush()
			tempFile.Close()

			// 重新打开临时文件用于上传
			tempFile, err = os.Open(tempFilePath)
			if err != nil {
				return nil, fmt.Errorf("打开临时文件失败: %w", err)
			}
			defer tempFile.Close()

			// 上传分片
			uploadPartReq := bdpan.NewUploadFilePartReq(remoteFilePath, preCreateRes.Uploadid, i)
			uploadPartReq.File = tempFile

			uploadPartRes, err := bdpan.UploadFilePart(accessToken, uploadPartReq)
			if err != nil {
				tempFile.Close()
				return nil, fmt.Errorf("分片 %d 上传失败: %w", i, err)
			}
			tempFile.Close()

			if uploadPartRes.IsError() {
				if progressBar != nil {
					progressBar.Finish()
				}
				return nil, fmt.Errorf("分片 %d 上传失败，%s", i, uploadPartRes.Error())
			}

			remoteBlockList = append(remoteBlockList, uploadPartRes.Md5)
			uPrintf("分片 %d/%d 上传成功，md5: %s path: %s", i+1, chunkCount, uploadPartRes.Md5, tempFilePath)
			if progressBar != nil {
				progressBar.Increment()
			}
		}
	}

	// 7. 创建文件
	createFileReq := bdpan.NewCreateFileReq(remoteFilePath, int32(fileSize), 0, preCreateRes.Uploadid, remoteBlockList)
	if isRewrite {
		createFileReq.SetRtype(3)
	}
	createFileRes, err := bdpan.CreateFile(accessToken, createFileReq)
	if err != nil {
		return nil, fmt.Errorf("创建文件失败: %w", err)
	}

	if createFileRes.IsError() {
		if progressBar != nil {
			progressBar.Finish()
		}
		return nil, fmt.Errorf("创建文件失败，%s", createFileRes.Error())
	}
	if progressBar != nil {
		progressBar.Finish()
	}

	uPrintf("文件创建成功，fs_id: %d name: %s", createFileRes.FSID, createFileRes.ServerFilename)
	return createFileRes, nil
}

// calculateBlockList 计算文件的MD5分块列表
func calculateBlockList(file *os.File, fileSize int64) ([]string, error) {
	blockList := make([]string, 0)

	// 如果文件小于等于4MB，只需要计算一个MD5
	if fileSize <= ChunkSize {
		hasher := md5.New()
		if _, err := io.Copy(hasher, file); err != nil {
			return nil, err
		}
		md5Str := hex.EncodeToString(hasher.Sum(nil))
		blockList = append(blockList, md5Str)
	} else {
		// 如果文件大于4MB，需要按照4MB大小分片计算MD5
		chunkCount := int(fileSize / ChunkSize)
		if fileSize%ChunkSize != 0 {
			chunkCount++
		}

		for i := 0; i < chunkCount; i++ {
			offset, err := file.Seek(int64(i)*ChunkSize, 0)
			if err != nil {
				return nil, err
			}
			if offset != int64(i)*ChunkSize {
				return nil, fmt.Errorf("failed to seek to expected position: got %d, want %d", offset, int64(i)*ChunkSize)
			}

			hasher := md5.New()
			r := io.LimitReader(file, ChunkSize)
			if _, err := io.Copy(hasher, r); err != nil {
				return nil, err
			}

			md5Str := hex.EncodeToString(hasher.Sum(nil))
			blockList = append(blockList, md5Str)
		}
	}

	// 重置文件指针
	offset, err := file.Seek(0, 0)
	if err != nil {
		return nil, err
	}
	if offset != 0 {
		return nil, fmt.Errorf("failed to reset file pointer: got offset %d", offset)
	}

	return blockList, nil
}
