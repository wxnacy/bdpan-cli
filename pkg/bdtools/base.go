package bdtools

import (
	"path/filepath"

	"github.com/wxnacy/go-tools"
)

func GetUserDataRoot() (string, error) {
	root, err := tools.GetUserDataRoot()
	if err != nil {
		return "", nil
	}
	root = filepath.Join(root, "bdpan")
	tools.DirExistsOrCreate(root)
	return root, nil
}
