package logic

import (
	"context"

	"github.com/TarsCloud/TarsGo/tars"

	"github.com/tarscloud/gopractice/common/tracing"

	"github.com/tarscloud/gopractice/apps/helloserver/config"
	"github.com/tarscloud/gopractice/apps/helloserver/proto/stub/Base"
	"github.com/tarscloud/gopractice/common/log"
)

var (
	myClient = &Base.Main{}
)

func init() {
	comm := tars.NewCommunicator()
	comm.StringToProxy("Base.HelloServer.MainObj", myClient)
}

// ServerImp servant implementation
type ServerImp struct {
}

// Add ...
// curl -d '{"A":2, "B": 3}' "http://jsontarsproxy/apis/v1/Add"
func (s *ServerImp) Add(ctx context.Context, req *Base.AddReq, rsp *Base.AddRsp) error {
	//Doing something in your function
	log.Debug(ctx, "Config value is %v", config.Get().Value)

	sRsp := &Base.SubRsp{}
	myClient.SubWithContext(ctx, &Base.SubReq{
		A: req.A,
		B: req.B,
	}, sRsp)
	rspx, err := tracing.Get(ctx, "http://qqxxxxxxx.com")
	if err != nil {
		return err
	}
	rspx.Body.Close()

	rsp.C = req.A + req.B
	return nil
}

// Sub ...
func (s *ServerImp) Sub(ctx context.Context, req *Base.SubReq, rsp *Base.SubRsp) error {
	//Doing something in your function
	//...
	return nil
}
