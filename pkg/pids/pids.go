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

var pidList []int
var pidListMutex = &sync.Mutex{}

func Add(cmd *exec.Cmd) {
	pidListMutex.Lock()
	defer pidListMutex.Unlock()

	if cmd != nil && cmd.Process != nil {
		pidList = append(pidList, cmd.Process.Pid)
	}
}

func Save(hash string) error {
	pidListMutex.Lock()
	defer pidListMutex.Unlock()

	file := utils.GetGomonPidListFile(hash)

	marshal, err := gobutils.Marshal(pidList)
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

	var oldPidList []int
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
