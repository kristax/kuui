package gui

import "github.com/go-kid/ioc"

func init() {
	ioc.Register(NewUI())
}
