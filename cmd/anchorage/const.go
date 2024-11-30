package main

import (
	"errors"
	"fmt"
	"github.com/peakedshout/anchorage-core/pkg/sdk"
	"github.com/peakedshout/go-pandorasbox/tool/wpath"
	"path"
)

const (
	projectName    = sdk.Name
	projectVersion = sdk.CoreVersion

	cfgDefaultDir = ".anchorage"
	logFile       = "run.log"
	cfgFile       = "cfg.yaml"
)

func GetProjectName() string {
	return projectName
}

func GetProjectVersion() string {
	return projectVersion
}

func GetProjectInfo() string {
	return fmt.Sprintf("\n%s\n%s:%s", sdk.CoreIcon, sdk.CoreName, sdk.CoreVersion)
}

func GetPath(paths ...string) string {
	return wpath.JoinHomePath(cfgDefaultDir, path.Join(paths...))
}

func ConfigPath() string {
	return GetPath(cfgFile)
}

func LogPath() string {
	return GetPath(logFile)
}

var ErrNotSupportCfg = errors.New("not support, please modify the configuration file")
