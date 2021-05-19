package config

import (
	"sync/atomic"
	"unsafe"

	"github.com/tarscloud/gopractice/common/filewatch"
)

type config struct {
	val int
}

// init the default value
var conf = &config{
	val: 2,
}

// -------------- Do not edit the code below  --------------
var pConf = unsafe.Pointer(conf)

// Init ...
func Init(path string) error {
	return filewatch.WatchFile(conf, path, func(val interface{}, isUpdate bool) {
		newVal := val.(*config)

		// custom logic

		atomic.StorePointer(&pConf, unsafe.Pointer(newVal))
	})
}

// Get returns the global config
func Get() *config {
	val := atomic.LoadPointer(&pConf)
	return (*config)(val)
}
