package main

import (
	"github.com/go-kid/ioc"
	_ "github.com/kristax/kuui/gui"
	//_ "github.com/kristax/kuui/kucli"
	"log"
)

func main() {
	log.Fatal(ioc.Run())
}
