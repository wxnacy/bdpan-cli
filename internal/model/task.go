package model

import (
	"time"
)

// Task 表示保存在 SQLite 中的通用后台任务（下载、上传等）。
//
// 字段保持通用化，类型相关的数据放在 Data(JSON 字符串) 中。
// identity 是任务的幂等键，例如 sha1("download|/src|/out")。
//
// 状态流转：queued -> running -> completed | failed | canceled | stale。
// stale 由外部逻辑判断（心跳超时 + 进程不存在）。
//
// CancelRequested 是协作取消标记，下载/上传循环需主动检查以便优雅退出。
// Progress 取值 [0,1]，配合字节字段用于计算速率和 ETA。
//
// GORM 标签建立必要索引，便于列表/状态查询。
// 注意：json 标签需稳定，未来若输出 JSON 不会变动字段名。
//
// 表名：tasks
// 索引：
//   - unique(identity)
//   - idx_tasks_status(status)
//   - idx_tasks_update_time(update_time)
//
// 时间字段序列化使用 RFC3339，数据库内保存 TEXT。
// Version 预留给未来的结构演进。
//
// 若存在子任务（如文件夹下载），通过外键关联到 Task.ID。

type Task struct {
	ID               string  `gorm:"primaryKey;column:id" json:"id"`
	Type             string  `gorm:"column:type;index" json:"type"`
	Identity         string  `gorm:"column:identity;uniqueIndex" json:"identity"`
	Status           string  `gorm:"column:status;index" json:"status"`
	Progress         float64 `gorm:"column:progress" json:"progress"`
	TotalBytes       int64   `gorm:"column:total_bytes" json:"total_bytes"`
	DownloadedBytes  int64   `gorm:"column:downloaded_bytes" json:"downloaded_bytes"`
	PID              int     `gorm:"column:pid" json:"pid"`
	Hostname         string  `gorm:"column:hostname" json:"hostname"`
	Concurrency      int     `gorm:"column:concurrency" json:"concurrency"`
	StartTime        string  `gorm:"column:start_time;index" json:"start_time"`
	UpdateTime       string  `gorm:"column:update_time;index" json:"update_time"`
	SpeedBPS         int64   `gorm:"column:speed_bps" json:"speed_bps"`
	ETASeconds       int64   `gorm:"column:eta_seconds" json:"eta_seconds"`
	Error            string  `gorm:"column:error" json:"error"`
	CancelRequested  int     `gorm:"column:cancel_requested" json:"cancel_requested"`
	Data             string  `gorm:"column:data;type:text" json:"data"`
	Version          int     `gorm:"column:version" json:"version"`
}

func (Task) TableName() string { return "tasks" }

// NewTask 构造一个运行中的 Task，填充起始字段。
func NewTask(id, typ, identity string, totalBytes int64, pid int, hostname string, concurrency int, data string) *Task {
	now := time.Now().Format(time.RFC3339)
	return &Task{
		ID:              id,
		Type:            typ,
		Identity:        identity,
		// 注意：初始状态需与 taskstore 中的常量保持一致（"运行中"），
		// 否则运行中任务在二次调用 ClaimOrCreate 时无法被判定为已运行，导致重复启动。
		Status:          "运行中",
		Progress:        0,
		TotalBytes:      totalBytes,
		DownloadedBytes: 0,
		PID:             pid,
		Hostname:        hostname,
		Concurrency:     concurrency,
		StartTime:       now,
		UpdateTime:      now,
		SpeedBPS:        0,
		ETASeconds:      0,
		CancelRequested: 0,
		Data:            data,
		Version:         1,
	}
}

// TaskChild 描述目录类任务的子文件条目，仅在需要跟踪子项时使用。

type TaskChild struct {
	ID          int64  `gorm:"primaryKey;column:id" json:"id"`
	TaskID      string `gorm:"column:task_id;index" json:"task_id"`
	Name        string `gorm:"column:name" json:"name"`
	Path        string `gorm:"column:path" json:"path"`
	Size        int64  `gorm:"column:size" json:"size"`
	Downloaded  int64  `gorm:"column:downloaded" json:"downloaded"`
	Status      string `gorm:"column:status;index" json:"status"`
	Error       string `gorm:"column:error" json:"error"`
	UpdateTime  string `gorm:"column:update_time;index" json:"update_time"`
}

func (TaskChild) TableName() string { return "task_children" }
