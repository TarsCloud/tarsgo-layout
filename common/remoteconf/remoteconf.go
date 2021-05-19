package remoteconf

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/configf"
)

var (
	comm         = tars.NewCommunicator()
	configClient = new(configf.Config)
)

func init() {
	cfg := tars.GetServerConfig()
	if cfg != nil {
		obj := tars.GetServerConfig().Config
		comm.StringToProxy(obj, configClient)
	}
}

// DownloadConfig downloads config file from tars system
func DownloadConfig(rootDir, filename string) error {
	cfg := tars.GetServerConfig()
	info := configf.ConfigInfo{
		Appname:     cfg.App,
		Servername:  cfg.Server,
		Setdivision: cfg.Setdivision,
		Host:        cfg.LocalIP,
		Filename:    filename,
	}
	savePath := filepath.Join(rootDir, filename)
	var data string
	ret, err := configClient.LoadConfigByInfo(&info, &data)
	if err != nil {
		return fmt.Errorf("LoadConfigByInfo error %v", err)
	}
	if ret != 0 {
		return fmt.Errorf("LoadConfigByInfo returns non-zero values %d", ret)
	}
	err = ioutil.WriteFile(savePath, []byte(data), 0644)
	if err != nil {
		return fmt.Errorf("WriteFile error %v", err)
	}
	return nil
}
