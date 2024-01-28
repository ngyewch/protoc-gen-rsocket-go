package runtime

import (
	"bytes"
	"context"
	"fmt"
	"google.golang.org/protobuf/proto"
)

var (
	ErrorSelectorMismatch = fmt.Errorf("selector mismatch")
)

type ClientRequestResponseHandler func(context.Context, []byte) ([]byte, error)

type ServerRequestResponseHandler func(context.Context, *RequestWrapper) (proto.Message, error)

func HandleClientRequestResponse(ctx context.Context, selector uint64, methodName string, req proto.Message, handler ClientRequestResponseHandler) ([]byte, error) {
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	reqWrapper := &RequestWrapper{
		Selector:   selector,
		MethodName: methodName,
		Payload:    reqBytes,
	}
	outputBuffer := bytes.NewBuffer(nil)
	err = reqWrapper.Marshal(outputBuffer)
	if err != nil {
		return nil, err
	}
	reqWrapperBytes := outputBuffer.Bytes()
	return handler(ctx, reqWrapperBytes)
}

func HandleServerRequestResponse(ctx context.Context, reqWrapperBytes []byte, handler ServerRequestResponseHandler) ([]byte, error) {
	inputBuffer := bytes.NewBuffer(reqWrapperBytes)
	var reqWrapper RequestWrapper
	err := reqWrapper.Unmarshal(inputBuffer)
	if err != nil {
		return nil, err
	}
	rsp, err := handler(ctx, &reqWrapper)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(rsp)
}
