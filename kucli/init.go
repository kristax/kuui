package kucli

import "github.com/go-kid/ioc"

func init() {
	ioc.Register(NewLocalCli())
}
