package streamer

import "context"

type streamer struct {
	Listeners []Listener `wire:""`
	listeners map[string][]OnCallFn
}

func NewStreamer() Client {
	return &streamer{
		listeners: make(map[string][]OnCallFn),
	}
}

func (s *streamer) Init() error {
	for _, listener := range s.Listeners {
		for _, channel := range listener.Channel() {
			s.listeners[channel] = append(s.listeners[channel], listener.OnCall)
		}
	}
	return nil
}

func (s *streamer) Send(ctx context.Context, channel string, msg any) {
	for _, fn := range s.listeners[channel] {
		go fn(ctx, channel, msg)
	}
}
