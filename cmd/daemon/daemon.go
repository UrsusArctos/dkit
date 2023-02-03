package main

import (
	"fmt"
	"time"

	"github.com/UrsusArctos/dkit/pkg/daemonizer"
	"github.com/UrsusArctos/dkit/pkg/logmeow"
)

type TMeowDaemon struct {
	LinuxDaemon daemonizer.TLinuxDaemon
	MeowLogger  logmeow.TLogMeow
}

func (MD TMeowDaemon) MeowInit() (err error) {
	MD.MeowLogger.LogEventInfo("Init called")
	return nil
}

func (MD TMeowDaemon) MeowClose() (err error) {
	MD.MeowLogger.LogEventInfo("Close called")
	MD.MeowLogger.Close()
	MD.LinuxDaemon.Close()
	return nil
}

func (MD TMeowDaemon) MeowRun() (err error) {
	time.Sleep(500 * time.Millisecond)
	MD.MeowLogger.LogEventInfo("Run called")
	return nil
}

func main() {
	md := TMeowDaemon{LinuxDaemon: daemonizer.NewLinuxDaemon("mymeow")}
	md.MeowLogger = logmeow.NewLogMeow("mymeow", logmeow.FacConsole|logmeow.FacFile, md.LinuxDaemon.LogPath)
	md.LinuxDaemon.FuncInit = md.MeowInit
	md.LinuxDaemon.FuncClose = md.MeowClose
	md.LinuxDaemon.FuncMain = md.MeowRun
	md.LinuxDaemon.TestFunc()
	e := md.LinuxDaemon.Run()
	fmt.Printf("exit %+v\n", e)
}
