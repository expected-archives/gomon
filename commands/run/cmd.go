package run

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/expectedsh/gomon/pkg/colors"
	"github.com/expectedsh/gomon/pkg/gomodule"
	"github.com/expectedsh/gomon/pkg/pids"
	"github.com/expectedsh/gomon/pkg/utils"
)

var Command = &cobra.Command{
	Use:          "run",
	Short:        "Run services described in the .gomon.yaml",
	Example:      "gomon run",
	RunE:         run,
	SilenceUsage: true,
}

var (
	fPid          bool
	fColors       bool
	fIgnoreError  bool
	fDirectories  []string
	fWatchTimeout time.Duration
	fKillTimeout  time.Duration
)

type applicationConfig struct {
	Name           string            `yaml:"name"`
	Path           string            `yaml:"path"`
	Env            map[string]string `yaml:"env"`
	Color          colors.Color      `yaml:"color"`
	MustNotRestart bool              `yaml:"must_not_restart"`

	// \/ \/ theses options are not handled currently \/ \/

	DirectoriesToWatch   []string `yaml:"directories_to_watch"`
	DirectoriesToExclude []string `yaml:"directories_to_exclude"`

	FilesToWatch   []string `yaml:"files_to_watch"`
	FilesToExclude []string `yaml:"files_to_exclude"`
}

var cfgHash string
var applicationConfigList []applicationConfig
var applications = map[string]*application{}

func run(c *cobra.Command, _ []string) error {
	cfg, err := c.Root().Flags().GetString("config")
	if err != nil {
		return err
	}

	if err := utils.InitConfig(cfg, &cfgHash, &applicationConfigList); err != nil {
		return errors.Wrap(err, "unable to get config file")
	}

	if err := pids.Kill(cfgHash); err != nil {
		return errors.Wrap(err, "unable to kill old pid run by gomon")
	}

	if len(applicationConfigList) == 0 {
		return errors.New("there is no application to run")
	}

	moduleName, err := gomodule.GetName()
	if err != nil {
		return errors.Wrap(err, "unable to get gomodule")
	}

	var (
		appPadding     = getAppPadding()
		wg             = sync.WaitGroup{}
		ctx, cancelCtx = context.WithCancel(context.Background())
	)

	for _, config := range applicationConfigList {
		app := newApplication(moduleName, config, appPadding)
		applications[config.Name] = app

		go handleRunningApplication(ctx, &wg, app)
	}

	if err := newWatcher(ctx).watchForRestarts(); err != nil {
		return err
	}

	go pids.SaveAtInterval(cfgHash)

	end := make(chan os.Signal, 1)
	signal.Notify(end, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	<-end

	cancelCtx()

	wg.Wait()

	if err := pids.Save(cfgHash); err != nil {
		return errors.Wrap(err, "unable to save pidList")
	}

	return nil
}

func handleRunningApplication(ctx context.Context, wg *sync.WaitGroup, app *application) {
	wg.Add(1)

	for {
		exit := make(chan bool)
		if err := app.run(exit); err != nil {
			app.log("This application could not be run because it encountered an error.", true, "GOMON")
			app.log("Waiting for a restart", true, "GOMON")
		}

		select {
		case <-ctx.Done():
			stopApp(false, app, exit)
			wg.Done()
			return
		case <-app.restart:
			stopApp(true, app, exit)
		case <-exit:
			if app.config.MustNotRestart {
				wg.Done()
				return
			}
		}
	}
}

func init() {
	Command.Flags().BoolVarP(
		&fColors, "colors",
		"c",
		true,
		"show the output with colors")

	Command.Flags().BoolVarP(
		&fPid, "pid",
		"p",
		false,
		"show the pid for each app")

	Command.Flags().BoolVarP(
		&fIgnoreError, "ignore-error",
		"i",
		false,
		"output error like normal message")

	Command.Flags().BoolVarP(
		&fIgnoreError, "ignore-build",
		"b",
		false,
		"ignore build output")

	Command.Flags().StringArrayVarP(
		&fDirectories, "directories",
		"d",
		[]string{"cmd", "pkg", "internal"},
		"list of directories to watch by default")

	Command.Flags().DurationVarP(
		&fWatchTimeout, "watch-timeout",
		"t",
		time.Second,
		"the duration after a change to restart an app")

	Command.Flags().DurationVarP(
		&fWatchTimeout, "kill-timeout",
		"k",
		time.Second*2,
		"kill the program after this duration if it is living after a sigint")

}

func stopApp(sigint bool, app *application, exit chan bool) {

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*2)
	defer cancelFunc()

	cmd := app.getCmd()
	if cmd == nil || cmd != nil && cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		return
	}

	if sigint && cmd != nil {
		if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
			if cmd.ProcessState.Exited() {
				return
			}
		}
	}

	alreadyForceKill := false

	for {
		if alreadyForceKill {
			<-exit
			return
		}

		select {
		case <-ctx.Done():
			app.log("This app take too long to be interrupted, so gomon will kill it.", false, "GOMON")
			cmd.Process.Kill()
			alreadyForceKill = true
			break
		case <-exit:
			return
		}
	}
}

func getAppPadding() int {
	max := -1
	for _, a := range applicationConfigList {
		if len(a.Name) > max {
			max = len(a.Name)
		}
	}

	return max
}
