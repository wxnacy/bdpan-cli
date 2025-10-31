package bdtools

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/wxnacy/go-tools"
)

func GetUserDataRoot() (string, error) {
	root, err := getUserDataRoot()
	if err != nil {
		return "", nil
	}
	root = filepath.Join(root, "bdpan")
	tools.DirExistsOrCreate(root)
	return root, nil
}

func getUserDataRoot() (string, error) {
	// 1. 优先检查XDG_DATA_HOME环境变量（Linux/macOS）
	if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		return xdgDataHome, nil
	}

	// 2. 未设置则使用系统默认路径
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("无法获取用户主目录: %w", err)
	}

	// 根据操作系统拼接默认路径
	switch os := runtime.GOOS; os {
	case "linux", "darwin":
		return filepath.Join(homeDir, ".local", "share"), nil
	case "windows":
		return filepath.Join(homeDir, "AppData", "Local"), nil
	default:
		return "", fmt.Errorf("不支持的操作系统: %s", os)
	}
}
