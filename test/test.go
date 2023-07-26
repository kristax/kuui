package main

//import (
//	"context"
//	"errors"
//	"fyne.io/fyne/v2"
//	"fyne.io/fyne/v2/app"
//	"github.com/fyne-io/terminal"
//	"github.com/go-kid/ioc"
//	ap "github.com/go-kid/ioc/app"
//	"github.com/kristax/kuui/kucli"
//	"io"
//	"time"
//)
//
//type Stream struct {
//	io.ReadWriteCloser
//	ch   chan []byte
//	Next *Stream
//}
//
//func (s *Stream) Read(p []byte) (n int, err error) {
//	by, ok := <-s.ch
//	if !ok {
//		return 0, errors.New("I/O read failed")
//	}
//	copy(p, by)
//	return len(by), nil
//}
//
//func (s *Stream) Write(p []byte) (n int, err error) {
//	return s.Next.Read(p)
//}
//
//func (s *Stream) Scan(p []byte) {
//	s.ch <- p
//}
//
//func (s *Stream) Close() error {
//	close(s.ch)
//	return nil
//}
//
//func NewStream(next *Stream) *Stream {
//	return &Stream{
//		ch:   make(chan []byte, 0),
//		Next: next,
//	}
//}
//
//func main() {
//	var a = &struct {
//		Cli kucli.KuCli `wire:""`
//	}{}
//	ioc.Run(ap.SetComponents(a))
//	ctx := context.Background()
//
//	v1 := NewStream(nil)
//	v2 := NewStream(v1)
//	go func() {
//		err := a.Cli.ExecPod(ctx, "api-adaptor-dev", "api-adaptor-5c8c958f9b-v4skt", v2, v2)
//		if err != nil {
//			panic(err)
//		}
//	}()
//
//	f := app.New()
//	w := f.NewWindow("hello")
//
//	t := terminal.New()
//	w.SetContent(t)
//	size := fyne.Size{
//		Width:  800,
//		Height: 400,
//	}
//	w.Resize(size)
//	t.Resize(size)
//
//	go func() {
//		_ = t.RunWithConnection(v2, v2)
//		//t.RunLocalShell()
//		f.Quit()
//	}()
//
//	go func() {
//		for range time.Tick(time.Second) {
//			v1.Scan([]byte("ls\n"))
//		}
//	}()
//	w.ShowAndRun()
//}
