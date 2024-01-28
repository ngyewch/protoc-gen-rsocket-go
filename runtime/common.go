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

type Server interface {
	HandleRequestResponse(context.Context, []byte) ([]byte, error)
}

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

func HandleServerRequestResponseAsync(ctx context.Context, reqWrapperBytes []byte, handler ServerRequestResponseHandlerAsync) *mo.Future[[]byte] {
	inputBuffer := bytes.NewBuffer(reqWrapperBytes)
	var reqWrapper RequestWrapper
	err := reqWrapper.Unmarshal(inputBuffer)
	if err != nil {
		return mo.NewFuture(func(resolve func([]byte), reject func(error)) {
			reject(err)
		})
	}
	return mo.NewFuture(func(resolve func([]byte), reject func(error)) {
		handler(ctx, &reqWrapper).
			Then(func(rsp proto.Message) (proto.Message, error) {
				b, err := proto.Marshal(rsp)
				if err != nil {
					reject(err)
					return nil, err
				}
				resolve(b)
				return rsp, nil
			}).
			Catch(func(err error) (proto.Message, error) {
				reject(err)
				return nil, err
			})
	})
}
