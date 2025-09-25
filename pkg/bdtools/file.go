package bdtools

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/wxnacy/bdpan-cli/internal/logger"
	"github.com/wxnacy/go-bdpan"
)

// GetDirAllFiles 获取目录下的所有文件（通过轮询方式）
func GetDirAllFiles(accessToken, dir string) ([]*bdpan.FileInfo, error) {
	req := bdpan.NewGetFileListReq()
	req.Dir = dir
	totalList := []*bdpan.FileInfo{}
	fileList := []*bdpan.FileInfo{}
	page := 1

	for {
		req.SetPage(page)
		res, err := bdpan.GetFileList(accessToken, req)
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

// 根据地址查找文件
// 在文件目录中循环查找是否有该名称文件
func GetFileByPath(accessToken, path string) (*bdpan.FileInfo, error) {
	getFileByPage := func(dir, name string, page int) (*bdpan.FileInfo, bool, error) {
		req := bdpan.NewGetFileListReq()
		req.Dir = dir
		req.SetPage(page)
		res, err := bdpan.GetFileList(accessToken, req)
		// fmt.Println(req)
		// fmt.Println(page, res, err)
		if err != nil {
			return nil, false, err
		}
		// 返回是否有下一页
		if len(res.List) == 0 {
			return nil, false, nil
		}
		// 过滤文件
		for _, f := range res.List {
			if f.GetFilename() == name {
				return f, false, nil
			}
		}
		return nil, true, nil
	}
	dir, name := filepath.Split(path)
	// 10万页循环
	for i := range 100000 {
		f, hasMove, err := getFileByPage(dir, name, i+1)
		if err != nil {
			return nil, err
		}
		if f != nil {
			req := bdpan.NewGetFileInfoReq(f.FSID)

			// 获取带有下载地址的文件详情
			infoRes, err := bdpan.GetFileInfo(accessToken, req)
			if err != nil {
				return nil, err
			}
			info := &infoRes.FileInfo
			info.Dlink = fmt.Sprintf("%s&access_token=%s", info.Dlink, accessToken)
			return info, nil
		} else {
			// 判断是否有下一页，没有直接返回
			if !hasMove {
				break
			}
		}
	}
	return nil, errors.New("file not found")
}

func PrintFileInfo(f *bdpan.FileInfo, height int) error {
	valueW := 50
	columns := []table.Column{
		{Title: "字段", Width: 10},
		{Title: "详情", Width: valueW},
	}

	rows := make([]table.Row, 0)
	rows = append(rows, table.Row{
		"FSID",
		fmt.Sprintf("%d", f.FSID),
	})
	filename := fmt.Sprintf("%s %s", f.GetFileTypeEmoji(), f.GetFilename()) + "\n"
	// newfilename := ""
	// for _, s := range filename {
	// w, _ := lipgloss.Size(newfilename)
	// }
	nameW, nameH := lipgloss.Size(filename)
	logger.Infof("名称尺寸 %dx%d", nameW, nameH)
	rows = append(rows, table.Row{
		"文件名",
		fmt.Sprintf("%s %s", f.GetFileTypeEmoji(), f.GetFilename()),
		// "你好\nss",
	})
	rows = append(rows, table.Row{
		"大小",
		f.GetSize(),
	})
	rows = append(rows, table.Row{
		"类型",
		f.GetFileType(),
	})
	rows = append(rows, table.Row{
		"地址",
		f.Path,
	})
	rows = append(rows, table.Row{
		"创建时间",
		f.GetServerCTime(),
	})
	rows = append(rows, table.Row{
		"修改时间",
		f.GetServerMTime(),
	})

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(height),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	// s.Selected = s.Selected.
	// Foreground(lipgloss.Color("229")).
	// Background(lipgloss.Color("57")).
	// Bold(false)
	t.SetStyles(s)
	fmt.Println(t.View())
	return nil
}
