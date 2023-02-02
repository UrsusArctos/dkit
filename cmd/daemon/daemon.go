package main

import (
	"UrsusArctos/dkit/pkg/daemonizer"
	"UrsusArctos/dkit/pkg/logmeow"
	"fmt"
	"time"
)

var meow logmeow.TLogMeow = logmeow.NewLogMeow("mymeow", logmeow.FacConsole)

func MeowInit() (err error) {
	meow.LogEventInfo("Init called")
	return nil
}

func MeowClose() (err error) {
	meow.LogEventInfo("Close called")
	return nil
}

func MeowRun() (err error) {
	time.Sleep(1 * time.Second)
	meow.LogEventInfo("Run called")
	return nil
}

func main() {
	defer meow.Close()
	ld := daemonizer.NewLinuxDaemon("mymeow")
	defer ld.Close()
	ld.FuncInit = MeowInit
	ld.FuncClose = MeowClose
	ld.FuncMain = MeowRun
	e := ld.Run()
	fmt.Printf("%+v\n", e)
	// ld.TestFunc()
}
