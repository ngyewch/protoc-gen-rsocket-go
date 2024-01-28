package runtime

import (
	"bytes"
	"context"
	"fmt"
	"github.com/samber/mo"
	"google.golang.org/protobuf/proto"
)

var (
	ErrorSelectorMismatch = fmt.Errorf("selector mismatch")
)

type ClientRequestResponseHandler func(context.Context, []byte) ([]byte, error)

type ClientRequestResponseHandlerAsync func(context.Context, []byte) *mo.Future[[]byte]

type ServerRequestResponseHandler func(context.Context, *RequestWrapper) (proto.Message, error)

type ServerRequestResponseHandlerAsync func(context.Context, *RequestWrapper) *mo.Future[proto.Message]

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

func HandleClientRequestResponseAsync(ctx context.Context, selector uint64, methodName string, req proto.Message, handler ClientRequestResponseHandlerAsync) *mo.Future[[]byte] {
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return mo.NewFuture(func(resolve func([]byte), reject func(error)) {
			reject(err)
		})
	}
	reqWrapper := &RequestWrapper{
		Selector:   selector,
		MethodName: methodName,
		Payload:    reqBytes,
	}
	outputBuffer := bytes.NewBuffer(nil)
	err = reqWrapper.Marshal(outputBuffer)
	if err != nil {
		return mo.NewFuture(func(resolve func([]byte), reject func(error)) {
			reject(err)
		})
	}
	reqWrapperBytes := outputBuffer.Bytes()
	return handler(ctx, reqWrapperBytes)
}
