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
	   or defaults to /var/log (assuming name.log.gz there)
	   --foreground
	   or defaults to false (implying being started as a daemon)
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
	// Foreground
	optFore  = "foreground"
	defFore  = false
	descFore = "Start in a foreground mode (don't daemonize)"
)

type (
	TLinuxDaemon struct {
		// internal
		name    string
		pidFile string
		// exported
		Foreground bool
		LogPath    string
		ConfFile   string
		FuncInit   TDaemonCycle
		FuncClose  TDaemonCycle
		FuncMain   TDaemonCycle
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
	flag.StringVar(&ld.LogPath, optLog, defLog, descLog)
	flag.BoolVar(&ld.Foreground, optFore, defFore, descFore)
	flag.Parse()
}

func (ld TLinuxDaemon) writePidFile() {
	f, err := os.Create(ld.pidFile)
	if err == nil {
		defer f.Close()
		f.WriteString(fmt.Sprintf("%d", os.Getpid()))
	}
}

func (ld TLinuxDaemon) Run() error {
	// run initialization, if any
	if ld.FuncInit != nil {
		errInit := ld.FuncInit()
		if errInit != nil {
			return errInit
		}
	}
	// set this daemon to receive SIGINT
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt)
	// run main loop
	var sigint bool = false
	var errMain error
	if ld.FuncMain != nil {
		for errMain = ld.FuncMain(); (errMain == nil) && (!sigint); errMain = ld.FuncMain() {
			// check if this cycle failed
			// if errMain != nil {
			// 	break
			// }
			// check if SIGINT is received
			select {
			case <-kill:
				{
					sigint = true
					errMain = nil
				}
			default:
			}
		}
	} else {
		// no main function specified, that's an error
		errMain = fmt.Errorf("FuncMain() is not set")
	}

	// stop receiving signals
	signal.Stop(kill)

	// run finalization, if any
	if ld.FuncClose != nil {
		errClose := ld.FuncClose()
		if errClose != nil {
			return errClose
		}
	}
	// all done, exit
	return errMain
}

func (ld TLinuxDaemon) TestFunc() {
	fmt.Printf("%+v\n", ld)
}
