package runtime

import (
	"bytes"
	"context"
	"fmt"
	"google.golang.org/protobuf/proto"
)

type Server interface {
	Selector() uint64
	HandleRequestResponse(context.Context, *RequestWrapper) (proto.Message, error)
}

type Servers struct {
	serverMap map[uint64]Server
}

func NewServers(servers ...Server) *Servers {
	s := &Servers{
		serverMap: make(map[uint64]Server),
	}
	for _, server := range servers {
		s.serverMap[server.Selector()] = server
	}
	return s
}

func (servers *Servers) AddServer(server Server) {
	servers.serverMap[server.Selector()] = server
}

func (servers *Servers) HandleRequestResponse(ctx context.Context, reqWrapperBytes []byte) ([]byte, error) {
	inputBuffer := bytes.NewBuffer(reqWrapperBytes)
	var reqWrapper RequestWrapper
	err := reqWrapper.Unmarshal(inputBuffer)
	if err != nil {
		return nil, err
	}
	server, ok := servers.serverMap[reqWrapper.Selector]
	if !ok {
		return nil, fmt.Errorf("unknown selector")
	}
	rsp, err := server.HandleRequestResponse(ctx, &reqWrapper)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(rsp)
}
