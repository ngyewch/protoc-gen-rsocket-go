package runtime

import (
	"context"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx"
	"github.com/rsocket/rsocket-go/rx/mono"
	"github.com/samber/mo"
)

func RSocketClientRequestResponseHandler(rs rsocket.RSocket) ClientRequestResponseHandler {
	return func(ctx context.Context, reqWrapperBytes []byte) ([]byte, error) {
		rspPayload, err := rs.RequestResponse(payload.New(reqWrapperBytes, nil)).
			Block(ctx)
		if err != nil {
			return nil, err
		}
		return rspPayload.Data(), nil
	}
}

func RSocketClientRequestResponseHandlerAsync(rs rsocket.RSocket) ClientRequestResponseHandlerAsync {
	return func(ctx context.Context, reqWrapperBytes []byte) *mo.Future[[]byte] {
		return mo.NewFuture(func(resolve func([]byte), reject func(error)) {
			rs.RequestResponse(payload.New(reqWrapperBytes, nil)).
				Subscribe(ctx,
					rx.OnNext(func(input payload.Payload) error {
						resolve(input.Data())
						return nil
					}),
					rx.OnError(func(err error) {
						reject(err)
					}),
				)
		})
	}
}

func RSocketServerRequestResponseHandler(servers *Servers) func(payload.Payload) mono.Mono {
	return func(msg payload.Payload) mono.Mono {
		return mono.FromFunc(func(ctx context.Context) (payload.Payload, error) {
			rspBytes, err := servers.HandleRequestResponse(ctx, msg.Data())
			if err != nil {
				return nil, err
			}
			return payload.New(rspBytes, nil), nil
		})
	}
}
