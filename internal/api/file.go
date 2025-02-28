package api

import (
	"fmt"

	"github.com/wxnacy/go-bdpan"
)

func GetAllFileList(token, dir string) ([]*bdpan.FileInfo, error) {

	getListByPage := func(dir string, page int) ([]*bdpan.FileInfo, bool, error) {
		req := bdpan.NewGetFileListReq()
		req.Dir = dir
		req.SetPage(page)
		res, err := bdpan.GetFileList(token, req)
		if err != nil {
			return nil, false, err
		}
		// 返回是否有下一页
		if len(res.List) == 0 {
			return nil, false, nil
		}
		return res.List, true, nil
	}

	var totalList = make([]*bdpan.FileInfo, 0)
	for i := 0; i < 100000; i++ {
		files, hasMove, err := getListByPage(dir, i+1)
		if err != nil {
			return nil, err
		}
		if !hasMove {
			break
		}

		totalList = append(totalList, files...)
	}
	return totalList, nil
}

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
