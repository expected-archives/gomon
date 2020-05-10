package run

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"github.com/expectedsh/gomon/pkg/colors"
	"github.com/expectedsh/gomon/pkg/gomodule"
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
	fConfig       string
	fDirectories  []string
	fWatchTimeout time.Duration
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

var applicationConfigList []applicationConfig
var applications = map[string]*application{}

func run(_ *cobra.Command, _ []string) error {
	if err := initConfig(); err != nil {
		return err
	}

	if len(applicationConfigList) == 0 {
		return errors.New("there is no application to run")
	}

	moduleName, err := gomodule.GetName()
	if err != nil {
		return err
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

	end := make(chan os.Signal, 1)
	signal.Notify(end, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	<-end

	cancelCtx()

	wg.Wait()

	return nil
}

func handleRunningApplication(ctx context.Context, wg *sync.WaitGroup, app *application) {
	wg.Add(1)

	for {
		exit := make(chan bool)
		if err := app.run(exit); err != nil {
			wg.Done()
			return
		}

		select {
		case <-ctx.Done():
			<-exit
			wg.Done()
			return
		case <-app.restart:
			if err := app.getCmd().Process.Signal(syscall.SIGINT); err != nil {
				if err := app.getCmd().Process.Kill(); err != nil {
					wg.Done()
					return
				}
			}
			<-exit
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
		"ignore errors")

	Command.Flags().StringVarP(
		&fConfig, "config",
		"C",
		"./.gomon.yaml",
		"the location of the config file that describe what to run")

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

}

func initConfig() error {
	bytes, err := ioutil.ReadFile(fConfig)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(bytes, &applicationConfigList); err != nil {
		return err
	}

	return nil
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
