package logic

import (
	"context"

	"github.com/tarscloud/gopractice/apps/helloserver/config"
	"github.com/tarscloud/gopractice/common/log"

	"github.com/tarscloud/gopractice/common/ecode"

	"github.com/tarscloud/gopractice/apps/autogen/TestApp"
)

// ServerImp servant implementation
type ServerImp struct {
}

// Add ...
func (s *ServerImp) Add(ctx context.Context, req *TestApp.AddReq, rsp *TestApp.AddRsp) error {
	//Doing something in your function
	log.Debug("Value is %v", config.Get().Value)
	return nil
}

// Sub ...
func (s *ServerImp) Sub(ctx context.Context, req *TestApp.SubReq, rsp *TestApp.SubRsp) error {
	//Doing something in your function
	//...
	return ecode.Server("not implement")
}
