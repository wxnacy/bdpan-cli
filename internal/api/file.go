package api

import (
	"fmt"

	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/bdpan/file"
)

func GetAllFileList(token, dir string) ([]*bdpan.FileInfoDto, error) {

	getListByPage := func(dir string, page int) ([]*bdpan.FileInfoDto, bool, error) {
		req := file.NewGetFileListReq()
		req.Dir = dir
		req.SetPage(page)
		res, err := file.GetFileList(token, req)
		if err != nil {
			return nil, false, err
		}
		// 返回是否有下一页
		if len(res.List) == 0 {
			return nil, false, nil
		}
		return res.List, true, nil
	}

	var totalList = make([]*bdpan.FileInfoDto, 0)
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

func GetFileInfo(token string, fsid uint64) (*bdpan.FileInfoDto, error) {
	req := file.NewGetFileInfoReq(fsid)
	infoRes, err := file.GetFileInfo(token, req)
	if err != nil {
		return nil, err
	}
	info := &infoRes.FileInfoDto
	info.Dlink = fmt.Sprintf("%s&access_token=%s", info.Dlink, token)
	return info, nil
}
