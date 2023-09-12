package streamer

import "context"

type Client interface {
	Register(l Listener)
	Send(ctx context.Context, channel string, msg any)
}

type Listener interface {
	Channel() string
	OnCall(ctx context.Context, msg any)
}

type OnCallFn = func(ctx context.Context, msg any)
