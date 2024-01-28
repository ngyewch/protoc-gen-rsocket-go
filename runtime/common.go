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

func HandleClientRequestResponse(ctx context.Context, selector uint64, methodName string, req proto.Message, handler func(ctx context.Context, reqWrapperBytes []byte) ([]byte, error)) ([]byte, error) {
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

func HandleServerRequestResponse(ctx context.Context, reqWrapperBytes []byte, handler func(ctx context.Context, reqWrapper *RequestWrapper) (proto.Message, error)) ([]byte, error) {
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
