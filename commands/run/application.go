package run

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/expectedsh/gomon/pkg/colors"
	"github.com/expectedsh/gomon/pkg/imports"
)

type application struct {
	config         applicationConfig
	paddingAppName int
	repo           string
	restart        chan bool

	mutex *sync.Mutex
	files map[string]bool
	cmd   *exec.Cmd
}

func newApplication(repo string, config applicationConfig, paddingAppName int) *application {
	app := &application{
		config:         config,
		paddingAppName: paddingAppName,
		repo:           repo,
		files:          make(map[string]bool),
		cmd:            nil,
		restart:        make(chan bool),
		mutex:          &sync.Mutex{},
	}

	app.updateFiles(app.config.Path)

	return app
}

func (a *application) build() error {
	buildBinaryCmd := exec.Command("go", "build", "-o", "bins/"+a.config.Name, a.config.Path)
	if err := buildBinaryCmd.Run(); err != nil {
		return err
	}

	return nil
}

func (a *application) run(exit chan bool) error {
	if err := a.build(); err != nil {
		return err
	}

	cmd := exec.Command("bins/" + a.config.Name)
	a.setCmd(cmd)

	// adding current environment variables
	for _, val := range os.Environ() {
		cmd.Env = append(cmd.Env, val)
	}

	// adding environment variables from the config
	for k, v := range a.config.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%v", k, v))
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	go a.handleLog(stderr, true)
	go a.handleLog(stdout, false)

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		_ = cmd.Wait()

		success := cmd.ProcessState.Success()
		exitCode := cmd.ProcessState.ExitCode()

		if success {
			a.log("successfully exited", false)
		} else {
			a.log(fmt.Sprintf("exited with code: %d", exitCode), true)
		}

		exit <- success
		close(exit)
	}()

	return nil
}

func (a *application) handleLog(r io.ReadCloser, error bool) {
	x := bufio.NewReader(r)
	for {
		line, err := x.ReadString('\n')
		if err != nil && (err == io.EOF || err == io.ErrClosedPipe ||
			strings.Contains(err.Error(), "file already closed")) {
			return
		} else if err != nil {
			a.log(fmt.Sprintf("cli: unable to read line for this process: %s", err.Error()), true)
		}

		a.log(line, error)
	}
}

func (a *application) log(line string, error bool) {
	lineToPrint := ""

	if fColors {
		lineToPrint += a.config.Color.String()
	}

	if fPid {
		if pid, err := a.getPid(); err == nil {
			lineToPrint += fmt.Sprintf("%05d ", pid)
		}
	}

	lineToPrint += fmt.Sprintf(fmt.Sprintf("%%-%ds |", a.paddingAppName), a.config.Name)

	if error {
		if fColors {
			lineToPrint += colors.Bold.String() + colors.Red.String()
		}

		lineToPrint += " ERR:"
	}

	lineToPrint += colors.Reset.String() + " " + line

	if !strings.HasSuffix(line, "\n") {
		lineToPrint += "\n"
	}

	fmt.Print(lineToPrint)
}

func (a *application) getPid() (int, error) {
	cmd := a.getCmd()

	if cmd == nil {
		return 0, errors.New("no command is running")
	}

	if cmd.Process == nil {
		return 0, errors.New("process is not running")
	}

	return cmd.Process.Pid, nil
}

func (a *application) getCmd() *exec.Cmd {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.cmd
}

func (a *application) setCmd(cmd *exec.Cmd) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.cmd = cmd
}

func (a *application) updateFiles(startFile string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	imports.FromFile(a.repo, startFile, a.files)
}

func (a *application) getFiles() map[string]bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.files
}
