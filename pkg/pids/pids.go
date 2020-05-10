package pids

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/expectedsh/gomon/pkg/gobutils"
	"github.com/expectedsh/gomon/pkg/utils"
)

var pidMap = map[string]int{}
var pidListMutex = &sync.Mutex{}

func Add(name string, cmd *exec.Cmd) {
	pidListMutex.Lock()
	defer pidListMutex.Unlock()

	if cmd != nil && cmd.Process != nil {
		pidMap[name] = cmd.Process.Pid
	}
}

func Save(hash string) error {
	pidListMutex.Lock()
	defer pidListMutex.Unlock()

	file := utils.GetGomonPidListFile(hash)

	marshal, err := gobutils.Marshal(pidMap)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(file, marshal, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func Kill(hash string) error {
	pidListMutex.Lock()
	defer pidListMutex.Unlock()

	file := utils.GetGomonPidListFile(hash)
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	var oldPidList map[string]int
	if err := gobutils.Unmarshal(content, &oldPidList); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	for _, pid := range oldPidList {
		syscall.Kill(-pid, syscall.SIGKILL)
	}

	return nil
}

func Load(hash string) (map[string]int, error) {
	file := utils.GetGomonPidListFile(hash)
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var oldPidList map[string]int
	if err := gobutils.Unmarshal(content, &oldPidList); err != nil {
		if err == io.EOF {
			return map[string]int{}, nil
		}
		return nil, err
	}

	return oldPidList, nil
}
