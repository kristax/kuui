package main

import (
	"github.com/go-kid/ioc"
	_ "github.com/kristax/kuui/gui"
	"log"
)

func main() {
	log.Fatal(ioc.Run())
}
