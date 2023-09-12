package streamer

import "context"

type streamer struct {
	listeners map[string][]OnCallFn
}

func NewStreamer() Client {
	return &streamer{
		listeners: make(map[string][]OnCallFn),
	}
}

func (s *streamer) Register(l Listener) {
	s.listeners[l.Channel()] = append(s.listeners[l.Channel()], l.OnCall)
}

func (s *streamer) Send(ctx context.Context, channel string, msg any) {
	for _, fn := range s.listeners[channel] {
		fn(ctx, msg)
	}
}
