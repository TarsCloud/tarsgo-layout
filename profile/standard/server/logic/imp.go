package logic

import (
	"context"

	"github.com/tarscloud/gopractice/common/errcode"

	"github.com/tarscloud/gopractice/apps/autogen/TestApp"
)

// ServerImp is the servant implementation
type ServerImp struct {
}

// Add ...
func (s *ServerImp) Add(ctx context.Context, req *TestApp.AddReq, rsp *TestApp.AddRsp) error {
	//Doing something in your function
	//...
	return errcode.Server("not implement")
}

// Sub ...
func (s *ServerImp) Sub(ctx context.Context, req *TestApp.SubReq, rsp *TestApp.SubRsp) error {
	//Doing something in your function
	//...
	return errcode.Server("not implement")
}
