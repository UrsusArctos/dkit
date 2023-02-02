//go:build linux

package daemonizer

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
)

const (
	/*
	   	The conventions for daemons are as such:

	   --conf /full/path/to/daemon.conf
	   or defaults /etc/name.conf
	   --pid /full/path/to/daemon.pid
	   or defaults to /run/name.pid
	   --logpath /full/path/to				(nb: no filename!)
	   or defaults to /var/log/name.log.gz
	*/
	// Config file
	optConf  = "conf"
	defConf  = "/etc/%s.conf"
	descConf = "Full pathname of config file"
	// PID file
	optPID  = "pid"
	defPID  = "/run/%s.pid"
	descPID = "Full pathname of pid file"
	// Log path
	optLog  = "logpath"
	defLog  = "/var/log/"
	descLog = "Directory for log files"
)

type (
	TLinuxDaemon struct {
		// internal
		name    string
		pidFile string
		logPath string
		// exported
		ConfFile  string
		FuncInit  TDaemonCycle
		FuncClose TDaemonCycle
		FuncMain  TDaemonCycle
	}

	TDaemonCycle func() (err error)
)

func NewLinuxDaemon(dname string) (ld TLinuxDaemon) {
	ld.name = dname
	ld.parseCmdLine()
	ld.writePidFile()
	ld.FuncInit = nil
	ld.FuncClose = nil
	ld.FuncMain = nil
	return ld
}

func (ld *TLinuxDaemon) Close() {
	os.Remove(ld.pidFile)
}

func (ld *TLinuxDaemon) parseCmdLine() {
	flag.StringVar(&ld.ConfFile, optConf, fmt.Sprintf(defConf, ld.name), descConf)
	flag.StringVar(&ld.pidFile, optPID, fmt.Sprintf(defPID, ld.name), descPID)
	flag.StringVar(&ld.logPath, optLog, defLog, descLog)
	flag.Parse()
}

func (ld TLinuxDaemon) writePidFile() {
	f, err := os.Create(ld.pidFile)
	if err == nil {
		defer f.Close()
		f.WriteString(fmt.Sprintf("%d", os.Getpid()))
	}
}

func (ld TLinuxDaemon) Run() (err error) {
	// run initialization, if any
	if ld.FuncInit != nil {
		err = ld.FuncInit()
		if err != nil {
			return err
		}
	}
	// set this daemon to receive SIGINT
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt)
	// run main loop
	var sigint bool = false
	for interr := ld.FuncMain(); (interr == nil) && (!sigint); interr = ld.FuncMain() {
		// check if this cycle failed
		if interr != nil {
			err = interr
			break
		}
		// check if SIGINT is received
		select {
		case <-kill:
			{
				sigint = true
				err = nil
			}
		default:
		}
	}

	// stop receiving signals
	signal.Stop(kill)

	// run finalization, if any
	if ld.FuncClose != nil {
		err = ld.FuncClose()
		if err != nil {
			return err
		}
	}
	// all done, exit
	return err
}

func (ld TLinuxDaemon) TestFunc() {
	fmt.Printf("%+v\n", ld)
}
