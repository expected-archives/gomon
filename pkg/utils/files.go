package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path"

	"github.com/ghodss/yaml"
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

	file := path.Join(out, "pids.data")
	if _, err := os.Stat(file); os.IsNotExist(err) {
		os.Create(file)
	}

	return file
}

func InitConfig(file string, cfgHash *string, appList interface{}) error {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	hasher := md5.New()
	hasher.Write(bytes)
	*cfgHash = hex.EncodeToString(hasher.Sum(nil))

	if err := yaml.Unmarshal(bytes, appList); err != nil {
		return err
	}

	return nil
}

func InitConfigHash(file string, cfgHash *string) error {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	hasher := md5.New()
	hasher.Write(bytes)
	*cfgHash = hex.EncodeToString(hasher.Sum(nil))

	return nil
}
