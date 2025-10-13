package bdtools

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/wxnacy/bdpan-cli/internal/logger"
	"github.com/wxnacy/bdpan-cli/pkg/whitetea"
	"github.com/wxnacy/go-bdpan"
	"github.com/wxnacy/go-tools"
	"golang.org/x/term"
)

func GetFileInfo(token string, fsid uint64) (*bdpan.FileInfo, error) {
	req := bdpan.NewGetFileInfoReq(fsid)
	infoRes, err := bdpan.GetFileInfo(token, req)
	if err != nil {
		return nil, err
	}
	info := &infoRes.FileInfo
	info.Dlink = fmt.Sprintf("%s&access_token=%s", info.Dlink, token)
	return info, nil
}

func BatchGetFileInfos(accessToken string, fsids []uint64) ([]*bdpan.FileInfo, error) {
	chunkSize := 100
	var allBatchFiles []*bdpan.FileInfo
	for i := 0; i < len(fsids); i += chunkSize {
		end := i + chunkSize
		if end > len(fsids) {
			end = len(fsids)
		}
		chunk := fsids[i:end]

		batchReq := &bdpan.BatchGetFileInfoReq{
			FSIDs: chunk,
			Dlink: 1,
		}
		batchRes, err := bdpan.BatchGetFileInfo(accessToken, batchReq)
		if err != nil {
			return nil, err // Or handle error more gracefully
		}
		for _, info := range batchRes.List {
			info.Dlink = fmt.Sprintf("%s&access_token=%s", info.Dlink, accessToken)
			allBatchFiles = append(allBatchFiles, info)
		}
		// allBatchFiles = append(allBatchFiles, batchRes.List...)
	}
	return allBatchFiles, nil
}

// 获取文件的真实md5
func GetFileContentMD5(file *bdpan.FileInfo) (string, error) {
	if file.Dlink == "" {
		return "", fmt.Errorf("file %s not found Dlink", file.Path)
	}
	// 创建一个HTTP客户端，允许重定向
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 允许重定向，但限制重定向次数
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			return nil
		},
	}

	// 发送HEAD请求以获取文件头信息，而不下载整个文件
	resp, err := client.Head(file.Dlink)
	if err != nil {
		return "", fmt.Errorf("HEAD request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// 尝试从响应头中获取Content-MD5
	contentMD5 := resp.Header.Get("Content-MD5")
	if contentMD5 != "" {
		return contentMD5, nil
	}

	// 如果响应头中没有Content-MD5，记录一条日志并返回空字符串
	logger.Debugf("Content-MD5 not found in headers for file: %s", file.GetFilename())
	return "", nil
}

// GetDirAllFiles 获取目录下的所有文件（通过轮询方式）
func GetDirAllFiles(accessToken, dir string) ([]*bdpan.FileInfo, error) {
	// req := bdpan.NewGetFileListReq()
	// req.Dir = dir
	// totalList := []*bdpan.FileInfo{}
	// fileList := []*bdpan.FileInfo{}
	// page := 1

	// for {
	// req.SetPage(page)
	// res, err := bdpan.GetFileList(accessToken, req)
	// if err != nil {
	// return nil, err
	// }
	// fileList = res.List
	// totalList = append(totalList, fileList...)

	// if len(fileList) <= 0 || len(fileList) < int(req.Limit) {
	// break
	// }
	// page++
	// }

	req := bdpan.NewGetFileListAllReq(dir)

	var allFiles []*bdpan.FileInfo

	for {
		res, err := bdpan.GetFileListAll(accessToken, req)
		if err != nil {
			return nil, err
		}

		if res.IsError() {
			return nil, res
		}

		allFiles = append(allFiles, res.List...)

		if res.HasMore == 0 {
			break
		}
		req.Start = res.Cursor
	}

	return allFiles, nil
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
			info, err := GetFileInfo(accessToken, f.FSID)
			if err != nil {
				return nil, err
			}
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

func GetFileInfoView(f *bdpan.FileInfo, args ...any) (string, error) {
	width := 100
	var height int

	for _, arg := range args {
		switch val := arg.(type) {
		case whitetea.Width:
			width = int(val)
		case whitetea.Height:
			height = int(val)
		}
	}

	keyW := 12
	valueW := width - keyW

	// keyStyle := lipgloss.NewStyle().Bold(true).Width(keyW)
	// valueStyle := lipgloss.NewStyle().Width(valueW)
	keyStyle := lipgloss.NewStyle().
		// BorderStyle(lipgloss.NormalBorder()).
		// BorderForeground(lipgloss.Color("240")).
		// BorderLeft(true).
		Align(lipgloss.Left).
		// Foreground(lipgloss.Color("#FAFAFA")).
		// Background(lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}).
		// Margin(1, 3, 0, 0).
		Padding(0, 1).
		Height(1).
		Width(keyW)
	valueStyle := lipgloss.NewStyle().
		// BorderStyle(lipgloss.NormalBorder()).
		// BorderForeground(lipgloss.Color("240")).
		// BorderRight(true).
		Align(lipgloss.Left).
		// Foreground(lipgloss.Color("#FAFAFA")).
		// Background(lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}).
		// Margin(1, 3, 0, 0).
		Padding(0, 1).
		Height(1).
		Width(valueW)

	var rows []string

	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top,
		keyStyle.Height(2).Render("字段名"),
		valueStyle.Height(2).Render("内容"),
	))
	if f != nil {

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top,
			keyStyle.Render("FSID"),
			valueStyle.Render(fmt.Sprintf("%d", f.FSID)),
		))
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top,
			keyStyle.Render("文件名"),
			valueStyle.Render(fmt.Sprintf("%s %s", f.GetFileTypeEmoji(), f.GetFilename())),
		))
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top,
			keyStyle.Render("大小"),
			valueStyle.Render(f.GetSize()),
		))
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top,
			keyStyle.Render("类型"),
			valueStyle.Render(f.GetFileType()),
		))
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top,
			keyStyle.Render("地址"),
			valueStyle.Render(f.Path),
		))
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top,
			keyStyle.Render("MD5"),
			valueStyle.Render(f.MD5),
		))
		if f.Dlink != "" {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top,
				keyStyle.Render("下载地址"),
				valueStyle.Render(f.Dlink),
			))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top,
			keyStyle.Render("创建时间"),
			valueStyle.Render(f.GetServerCTime()),
		))
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top,
			keyStyle.Render("修改时间"),
			valueStyle.Render(f.GetServerMTime()),
		))
	} else {
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top,
			keyStyle.Render("空"),
			valueStyle.Render("空"),
		))
	}

	curHeight := lipgloss.Height(lipgloss.JoinVertical(lipgloss.Left, rows...))
	if height > curHeight {
		blankH := height - curHeight
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top,
			keyStyle.Height(blankH).Render(""),
			valueStyle.Height(blankH).Render(""),
		))

	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...), nil
}

func PrintFileInfo(f *bdpan.FileInfo, args ...any) error {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	width /= 2
	if width < 100 {
		width = 100
	}
	width -= 2
	view, err := GetFileInfoView(f, whitetea.Width(width))
	baseStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))
	view = baseStyle.Render(view)
	logger.Printf("%s", view)
	return err
}

// 获取文件本地地址
func GetFileLocalPath(f *bdpan.FileInfo) (string, error) {
	root, err := GetUserDataRoot()
	if err != nil {
		return "", err
	}
	p := filepath.Join(root, "tmpfile", f.MD5, f.GetFilename())
	return p, nil
}

func HasLocalFile(f *bdpan.FileInfo) bool {
	p, err := GetFileLocalPath(f)
	if err != nil {
		return false
	}
	return tools.FileExists(p)
}

// 下载文件到本地
// 不超过 1 M的文件
func DownloadFileToLocal(accessToken string, f *bdpan.FileInfo) (string, error) {
	p, err := GetFileLocalPath(f)
	if err != nil {
		return "", err
	}
	if HasLocalFile(f) {
		return p, nil
	}
	dir := filepath.Dir(p)
	tools.DirExistsOrCreate(dir)
	if f.Dlink == "" {
		f, err = GetFileInfo(accessToken, f.FSID)
		if err != nil {
			return "", err
		}
	}
	err = tools.Download(f.Dlink, p)
	if err != nil {
		return "", err
	}
	if err != nil {
		return "", err
	}
	return p, nil
}
