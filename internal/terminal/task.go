package terminal

import (
	"fmt"
	"strconv"

	"github.com/wxnacy/bdpan-cli/internal/common"
	"github.com/wxnacy/bdpan-cli/internal/model"
)

type TaskType int
type TaskStatus int

const (
	TypeDelete TaskType = iota
	TypeDownload

	StatusWating TaskStatus = iota
	StatusRunning
	StatusSuccess
	StatusFailed
)

func NewTask(type_ TaskType, f *model.File) *Task {
	idStr := fmt.Sprintf("%s%d", common.FormatNumberWithTrailingZeros(int(type_), 3), f.FSID)
	id, _ := strconv.Atoi(idStr)
	return &Task{
		ID:     id,
		File:   f,
		Type:   type_,
		Status: StatusWating,
	}
}

type Task struct {
	ID        int
	File      *model.File
	Type      TaskType
	Status    TaskStatus
	IsConfirm bool
	err       error
}

func (t Task) GetTypeString() string {
	switch t.Type {
	case TypeDelete:
		return "Delete"
	case TypeDownload:
		return "Download"
	}
	panic("unkown type")
}

func (t Task) GetStatusString() string {
	switch t.Status {
	case StatusWating:
		return "Wating"
	case StatusRunning:
		return "Running"
	case StatusSuccess:
		return "Success"
	case StatusFailed:
		return "Failed"
	}
	panic("unkown status")
}

func (t Task) String() string {
	err := ""
	if t.err != nil {
		err = t.err.Error()
	}
	return fmt.Sprintf(
		"%s: %s %s %s",
		t.GetTypeString(),
		t.File.GetFilename(),
		t.GetStatusString(),
		err,
	)
}
