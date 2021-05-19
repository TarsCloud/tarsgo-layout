package filewatch

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"time"

	"github.com/jinzhu/copier"
	"muzzammil.xyz/jsonc"
)

var (
	unMarshaller  = jsonc.Unmarshal
	watchInterval = time.Second * 5
)

type logger func(msg string)

// WatchFile ...
func WatchFile(initVal interface{}, path string, callback func(interface{}, bool)) error {
	// copy value first
	cloneInitVal := clonePtrValue(initVal)

	// read config file and callback first time
	cloneVal := clonePtrValue(cloneInitVal)
	lastVal, err := readFileToValue(path, cloneVal)
	if err != nil {
		return fmt.Errorf("read file error %v", err)
	}
	bs, _ := json.MarshalIndent(cloneVal, "", "\t")
	log("Init config: " + string(bs))
	callback(cloneVal, false)

	// loop update file
	if watchInterval > 0 {
		go func() {
			for range time.NewTicker(watchInterval).C {
				cloneVal := clonePtrValue(cloneInitVal)
				currVal, err := readFileToValue(path, cloneVal)
				if err != nil {
					log("Update value error: " + err.Error())
					continue
				}
				if lastVal == currVal {
					continue
				}
				bs, _ := json.MarshalIndent(cloneVal, "", "\t")
				log("Update Config: " + string(bs) + currVal)
				lastVal = currVal
				callback(cloneVal, true)
			}
		}()
	}
	return nil
}

// SetLogger ...
func SetLogger(lg logger) {
	log = lg
}

func readFileToValue(path string, val interface{}) (string, error) {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(bs), unMarshaller(bs, val)
}

func clonePtrValue(src interface{}) interface{} {
	v := reflect.New(reflect.ValueOf(src).Elem().Type()).Interface()
	copier.Copy(v, src)
	return v
}

var log logger = func(msg string) {
	fmt.Println(msg)
}
