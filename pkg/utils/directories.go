package utils

import (
	"io/ioutil"
	"path"
)

func GetSubDirectories(dir string, directories *[]string) {
	readDir, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}

	for _, info := range readDir {
		dir := path.Join(dir, info.Name())
		if info.IsDir() {
			*directories = append(*directories, dir)
			GetSubDirectories(dir, directories)
		}
	}
}
