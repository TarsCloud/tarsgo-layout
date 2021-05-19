package main

import (
	"github.com/tarscloud/gopractice/apps/autogen/TestApp"
	"github.com/tarscloud/gopractice/apps/helloserver/config"
	"github.com/tarscloud/gopractice/apps/helloserver/logic"
	"github.com/tarscloud/gopractice/common/initserver"
	"github.com/TarsCloud/TarsGo/tars"
)

func main() {
	// Init server
	cfg := tars.GetServerConfig()
	if err := initserver.NewOption().
		WithRemoteConf(cfg.Server+".yaml", config.Init).
		DoInit(); err != nil {
		panic(err)
	}

	// Add servant
	imp := new(logic.ServerImp)
	app := new(TestApp.Main)
	app.AddServantWithContext(imp, cfg.App+"."+cfg.Server+".MainObj")

	// Run application
	tars.Run()
}
