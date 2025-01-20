package main

import (
	"github.com/rendau/graylog_notify/internal/app"
)

func main() {
	a := &app.App{}
	a.Init()
	a.Start()
	a.Listen()
	a.Stop()
	a.Exit()
}
