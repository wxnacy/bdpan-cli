package taskstore

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"syscall"
	"time"

	"github.com/wxnacy/bdpan-cli/internal/model"
	"gorm.io/gorm"
)

// taskstore 封装了基于 SQLite（通过 gorm）的轻量持久层，用于管理下载、上传等长时间运行的任务。
// 主要职责：统一领取任务、心跳上报、列表查询、取消与恢复等操作。
//
// 设计说明：
// - 使用 model.Task 保存通用字段，类型差异数据以 JSON 形式写入 Task.Data。
// - 调用方需传入 identity（如 sha1("download|path|output")) 确保幂等，一个 identity 同时只允许一个运行任务。
// - ClaimOrCreate 负责抢占/创建任务：
//     * 如存在“运行中且仍存活”的任务 => attached=true，直接复用
//     * 如任务陈旧或不存在 => 接管为运行中或新建
// - Heartbeat 更新进度、时间等字段，并返回 CancelRequested，用于协作式取消。
// - Cancel 将 CancelRequested 置为 1，提示执行端尽快退出。
//
// 注意：本模块依赖 WAL 模式和小连接池配置（见 model.InitSqlite）。

const (
	TaskTypeDownload = "下载"
	statusRunning    = "运行中"
	statusCompleted  = "已完成"
	statusFailed     = "失败"
	statusCanceled   = "已取消"
	statusStale      = "陈旧"
)

type HeartbeatData struct {
	DownloadedBytes int64
	TotalBytes      int64
	Progress        float64
	SpeedBPS        int64
	ETASeconds      int64
}

type DownloadData struct {
	FSID       uint64 `json:"fsid,omitempty"`
	Path       string `json:"path"`
	MD5        string `json:"md5,omitempty"`
	OutputDir  string `json:"output_dir,omitempty"`
	TargetPath string `json:"target_path,omitempty"`
	IsDir      bool   `json:"is_dir"`
}

// BuildIdentitySHA1 根据输入片段生成幂等用的 identity。
func BuildIdentitySHA1(parts ...string) string {
	h := sha1.New()
	for _, p := range parts {
		_, _ = h.Write([]byte(p))
		_, _ = h.Write([]byte{'|'})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// ClaimOrCreate 尝试以 identity 附着到已有运行任务；若不存在或已陈旧则创建/接管。
// attached==true 表示已存在活跃任务，调用方不应重复启动，而是提示用户查看状态。
func ClaimOrCreate(ctx context.Context, typ, identity, id string, totalBytes int64, data any) (taskID string, attached bool, err error) {
	db := model.GetDB()
	returnID := id

	// serialize data
	var dataStr string
	if data != nil {
		b, _ := json.Marshal(data)
		dataStr = string(b)
	}

	// one-shot tx
	err = db.Transaction(func(tx *gorm.DB) error {
		var task model.Task
		errq := tx.Where("identity = ?", identity).Take(&task).Error
		if errq == nil {
			// found
			alive := isAlive(task)
			if task.Status == statusRunning && alive {
				taskID = task.ID
				attached = true
				return nil
			}
			// stale takeover
			host, _ := os.Hostname()
			now := time.Now().Format(time.RFC3339)
			updates := map[string]any{
				"status":      statusRunning,
				"pid":         os.Getpid(),
				"hostname":    host,
				"update_time": now,
			}
			if err := tx.Model(&model.Task{}).Where("id = ?", task.ID).Updates(updates).Error; err != nil {
				return err
			}
			taskID = task.ID
			attached = false
			return nil
		} else if !errors.Is(errq, gorm.ErrRecordNotFound) {
			return errq
		}

		// create
		host, _ := os.Hostname()
		if returnID == "" {
			returnID = identity // default id = identity
		}
		nt := model.NewTask(returnID, typ, identity, totalBytes, os.Getpid(), host, 0, dataStr)
		if err := tx.Create(nt).Error; err != nil {
			return err
		}
		taskID = nt.ID
		attached = false
		return nil
	})
	return taskID, attached, err
}

func Heartbeat(ctx context.Context, taskID string, hb HeartbeatData) (cancelRequested bool, err error) {
	db := model.GetDB()
	now := time.Now().Format(time.RFC3339)
	r := db.Model(&model.Task{}).
		Where("id = ?", taskID).
		Updates(map[string]any{
			"downloaded_bytes": hb.DownloadedBytes,
			"total_bytes":      hb.TotalBytes,
			"progress":         hb.Progress,
			"speed_bps":        hb.SpeedBPS,
			"eta_seconds":      hb.ETASeconds,
			"update_time":      now,
		})
	if r.Error != nil {
		return false, r.Error
	}
	var t model.Task
	if err := db.First(&t, "id = ?", taskID).Error; err != nil {
		return false, err
	}
	return t.CancelRequested == 1, nil
}

func Get(ctx context.Context, taskID string) (*model.Task, error) {
	var t model.Task
	if err := model.GetDB().First(&t, "id = ?", taskID).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func ListRunning(ctx context.Context) ([]model.Task, error) {
	var list []model.Task
	if err := model.GetDB().Where("status = ?", statusRunning).Order("update_time DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func Cancel(ctx context.Context, taskID string) error {
	return model.GetDB().Model(&model.Task{}).Where("id = ?", taskID).
		Update("cancel_requested", 1).Error
}

func Complete(ctx context.Context, taskID string) error {
	return model.GetDB().Model(&model.Task{}).Where("id = ?", taskID).
		Updates(map[string]any{
			"status":      statusCompleted,
			"update_time": time.Now().Format(time.RFC3339),
		}).Error
}

func Fail(ctx context.Context, taskID string, errMsg string) error {
	return model.GetDB().Model(&model.Task{}).Where("id = ?", taskID).
		Updates(map[string]any{
			"status":      statusFailed,
			"error":       errMsg,
			"update_time": time.Now().Format(time.RFC3339),
		}).Error
}

func SetCanceled(ctx context.Context, taskID string) error {
	return model.GetDB().Model(&model.Task{}).Where("id = ?", taskID).
		Updates(map[string]any{
			"status":      statusCanceled,
			"update_time": time.Now().Format(time.RFC3339),
		}).Error
}

func isAlive(t model.Task) bool {
    // 优先依据心跳时间判断，15 秒内视为存活
    if ut, err := time.Parse(time.RFC3339, t.UpdateTime); err == nil {
        if time.Since(ut) <= 15*time.Second {
            return true
        }
    }

    // 进程探测作为补充。macOS 上不同权限可能返回 EPERM，遇到 EPERM 也视为进程存在。
    if t.PID > 0 {
        if err := syscall.Kill(t.PID, 0); err == nil {
            return true
        } else if errors.Is(err, syscall.EPERM) {
            return true
        }
    }
    return false
}
