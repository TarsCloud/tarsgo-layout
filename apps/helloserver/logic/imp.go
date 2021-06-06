package logic

import (
	"context"

	"github.com/tarscloud/gopractice/common/tracing"

	"github.com/tarscloud/gopractice/apps/autogen/TestApp"
	"github.com/tarscloud/gopractice/apps/helloserver/config"
	"github.com/tarscloud/gopractice/common/ecode"
	"github.com/tarscloud/gopractice/common/log"
)

// ServerImp servant implementation
type ServerImp struct {
}

// Add ...
// curl -d '{"A":2, "B": 3}' "http://jsontarsproxy/apis/v1/Add"
func (s *ServerImp) Add(ctx context.Context, req *TestApp.AddReq, rsp *TestApp.AddRsp) error {
	//Doing something in your function
	log.Debug(ctx, "Config value is %v", config.Get().Value)

	rspx, err := tracing.Get(ctx, "http://qq.com")
	if err != nil {
		return err
	}
	rspx.Body.Close()

	rsp.C = req.A + req.B
	return nil
}

// Sub ...
func (s *ServerImp) Sub(ctx context.Context, req *TestApp.SubReq, rsp *TestApp.SubRsp) error {
	//Doing something in your function
	//...
	return ecode.Server("not implement")
}
