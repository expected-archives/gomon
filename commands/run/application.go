package run

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/expectedsh/gomon/pkg/colors"
	"github.com/expectedsh/gomon/pkg/imports"
	"github.com/expectedsh/gomon/pkg/pids"
	"github.com/expectedsh/gomon/pkg/utils"
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
	buildBinaryCmd := exec.Command("go", "build", "-o", a.getBin(), a.config.Path)

	stdout, err := buildBinaryCmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := buildBinaryCmd.StderrPipe()
	if err != nil {
		return err
	}

	go a.handleLog(stdout, false, "BUILDER")
	go a.handleLog(stderr, true, "BUILDER")

	if err := buildBinaryCmd.Start(); err != nil {
		return err
	}

	buildBinaryCmd.Wait()

	if !buildBinaryCmd.ProcessState.Success() {
		return errors.New("build unsuccessful")
	}

	return nil
}

func (a *application) run(exit chan bool) error {
	if err := a.build(); err != nil {
		return errors.Wrap(err, "unable to build application "+a.config.Name)
	}

	cmd := exec.Command(a.getBin())
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

	go a.handleLog(stderr, true, "")
	go a.handleLog(stdout, false, "")

	if err := cmd.Start(); err != nil {
		return err
	}

	time.AfterFunc(time.Millisecond*500, func() {
		pids.Add(cmd)
	})

	go func() {
		_ = cmd.Wait()

		success := cmd.ProcessState.Success()
		exitCode := cmd.ProcessState.ExitCode()

		if success {
			a.log("successfully exited", false, "")
		} else {
			a.log(fmt.Sprintf("exited with code: %d", exitCode), true, "")
		}

		exit <- success
		close(exit)
	}()

	return nil
}

func (a *application) handleLog(r io.ReadCloser, error bool, prefix string) {
	x := bufio.NewReader(r)
	for {
		line, err := x.ReadString('\n')
		if err != nil && (err == io.EOF || err == io.ErrClosedPipe ||
			strings.Contains(err.Error(), "file already closed")) {
			return
		} else if err != nil {
			a.log(fmt.Sprintf("cli: unable to read line for this process: %s", err.Error()), true, prefix)
		}

		a.log(line, error, prefix)
	}
}

func (a *application) log(line string, error bool, prefix string) {
	lineToPrint := ""

	if fColors {
		lineToPrint += a.config.Color.String()
	}

	hasPid := false
	if fPid {
		if pid, err := a.getPid(); err == nil {
			hasPid = true
			lineToPrint += fmt.Sprintf("%05d ", pid)
		}
	}

	if !hasPid {
		lineToPrint += "      "
	}

	lineToPrint += fmt.Sprintf(fmt.Sprintf("%%-%ds |", a.paddingAppName), a.config.Name)

	if error && !fIgnoreError {
		if fColors {
			lineToPrint += colors.Bold.String() + colors.Red.String()
		}

		lineToPrint += " ERR:"
	}

	if prefix != "" {
		if fColors {
			lineToPrint += colors.Reset.String() + colors.Bold.String()
		}

		lineToPrint += " " + strings.ToUpper(prefix) + ":"
	}

	if fColors {
		lineToPrint += colors.Reset.String()
	}

	lineToPrint += " " + line

	if !strings.HasSuffix(line, "\n") {
		lineToPrint += "\n"
	}

	fmt.Print(lineToPrint)
}

func (a application) getBin() string {
	return path.Join(utils.GetGomonBuilds(cfgHash), a.config.Name)
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
