package model

import (
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/Tech-Arch1tect/rkllmopenapi/config"
)

var (
	ModelList = []Model{}
)

func RefreshModelList() {
	ModelList = ModelList[:0]
	root := config.C.StoragePath

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".rkllm") {
			ModelList = append(ModelList, Model{
				ModelName: d.Name(),
				ModelPath: path,
				ModelDir:  filepath.Dir(path),
			})
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
