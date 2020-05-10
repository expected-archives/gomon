package utils

import (
	"os"
	"path"
)

func GetGomonBuilds(hash string) string {
	out := path.Join(os.TempDir(), "gomon", hash, "builds")

	if _, err := os.Stat(out); os.IsNotExist(err) {
		os.MkdirAll(out, os.ModePerm)
	}

	return out
}

func GetGomonPidListFile(hash string) string {
	out := path.Join(os.TempDir(), "gomon", hash)

	if _, err := os.Stat(out); os.IsNotExist(err) {
		os.MkdirAll(out, os.ModePerm)
	}

	file := path.Join(out, "pidlist")
	if _, err := os.Stat(file); os.IsNotExist(err) {
		os.Create(file)
	}

	return file
}
