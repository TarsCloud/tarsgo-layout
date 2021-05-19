package config

import (
	"sync/atomic"
	"unsafe"

	"github.com/tarscloud/gopractice/common/filewatch"
)

// Action各接口配置
type Action struct {
	Name      string
	Addr      string
	Cluster   string
	TimeoutMS int
	HashKey   string
}

type config struct {
	Logging    *logging
	ActionList []Action
	ActionMap  map[string]Action
}

type logging struct {
	ReqFields map[string]bool
}

// init the default value
var conf = &config{
	Logging: &logging{ReqFields: map[string]bool{"ReqId": true}},
}

// -------------- Do not edit the code below  --------------
var pConf = unsafe.Pointer(conf)

// Init...
func Init(path string) error {
	return filewatch.WatchFile(conf, path, func(val interface{}, isUpdate bool) {
		newVal := val.(*config)
		// custom logic
		actionMap := make(map[string]Action)
		for _, v := range newVal.ActionList {
			if v.TimeoutMS <= 0 {
				v.TimeoutMS = 5000
			}
			actionMap[v.Name] = v
		}
		newVal.ActionMap = actionMap

		atomic.StorePointer(&pConf, unsafe.Pointer(newVal))
	})
}

// Get ...
func Get() *config {
	val := atomic.LoadPointer(&pConf)
	return (*config)(val)
}
