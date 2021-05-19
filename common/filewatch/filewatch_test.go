package filewatch

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type st struct {
	A string `json:"a"`
}

func TestWatch(t *testing.T) {
	watchInterval = time.Millisecond * 10
	log = func(msg string) {}
	path := "./test.json"
	defer os.Remove(path)
	firstVal := strings.Repeat("a", 10)
	nextVal := strings.Repeat("b", 20)
	err := ioutil.WriteFile(path, []byte("{\"a\": \""+firstVal+"\"}"), 0644)
	assert.Nil(t, err)

	v := &st{A: "xab"}
	err = WatchFile(v, path, func(val interface{}, isInit bool) {
		vv := val.(*st)
		if vv.A != firstVal && vv.A != nextVal {
			assert.Fail(t, "bad value")
		}
	})
	time.Sleep(time.Millisecond * 20)
	go func() {
		for i := 1; i < 1000; i++ {
			//bakPath := path + ".bak"
			err = ioutil.WriteFile(path, []byte("{\"a\": \""+nextVal+"\"}"), 0644)
			//os.Rename(bakPath, path)
			assert.Nil(t, err)
			time.Sleep(time.Millisecond * 10)
		}
	}()
	assert.Nil(t, err)
	time.Sleep(time.Second * 1)
}
