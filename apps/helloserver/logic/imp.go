package logic

import (
	"context"

	"github.com/TarsCloud/TarsGo/tars"

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

// SayHello ...
// curl -d '{"msg": "bob"}' "http://172.25.0.3:8082/apis/v1/sayHello"
func (s *ServerImp) SayHello(ctx context.Context, req *Base.SayHelloRequest) (rsp Base.SayHelloReply, err error) {
	//Doing something in your function
	log.Debug(ctx, "Config value is %v", config.Get().Value)

	hiRsp, err := myClient.SayHiWithContext(ctx, &Base.SayHiRequest{
		Name: req.Msg,
	})
	if code := tars.GetErrorCode(err); code != 0 {
		log.Error(ctx, "SayHi error %v", err)
		return
	}

	rsp.Reply = "reply message:" + hiRsp.Reply
	return
}

// SayHi ...
func (s *ServerImp) SayHi(ctx context.Context, req *Base.SayHiRequest) (rsp Base.SayHiReply, err error) {
	if req.Name == "defoo" {
		err = tars.Errorf(4004, "bad gay")
		return
	}
	rsp.Reply = "hi " + req.Name
	return
}
