package upsdclient

import (
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

type (
	// Defines UPS entity
	TUPS struct {
		Description string
	}

	// Defines connection to UPSD instance
	TUPSDClientConnection struct {
		tcpconn *net.TCPConn
		ups     map[string]TUPS
		defups  string
	}
)

func NewConnection(raddr net.TCPAddr) (uconn TUPSDClientConnection, err error) {
	tconn, err := net.DialTCP("tcp", nil, &raddr)
	if err == nil {
		uconn.tcpconn = tconn
		uconn.ups = make(map[string]TUPS)
	} else {
		uconn.tcpconn = nil
		uconn.ups = nil
	}
	return uconn, err
}

func (UPSDC *TUPSDClientConnection) CloseConnection() {
	UPSDC.transactionRAW(kwLOGOUT)
	UPSDC.tcpconn.Close()
	UPSDC.tcpconn = nil
}

func (UPSDC TUPSDClientConnection) transactionRAW(cmd string) (response []string, err error) {
	// Add finalizing LF
	cmd = fmt.Sprintf("%s\n", cmd)
	// Send command to UPSD
	_, writeerr := fmt.Fprint(UPSDC.tcpconn, cmd)
	if writeerr == nil {
		// Prepare stream reader
		rereader := bufio.NewReader(UPSDC.tcpconn)
		multiline := false
		for {
			// read another line
			reline, readerr := rereader.ReadString('\n')
			if readerr == nil {
				// check if it is a beginning of multiline response
				if reline == fmt.Sprintf("%s %s", kwBEGIN, cmd) {
					multiline = true
					continue
				}
				// check if multiline response ends
				if reline == fmt.Sprintf("%s %s", kwEND, cmd) {
					return response, nil
				}
				// trim essential lines from trailing CR
				reline = strings.TrimRight(reline, "\n")
				if multiline {
					// accumulate lines
					response = append(response, reline)
				} else { // return single line
					return []string{reline}, nil
				}
			} else {
				return nil, readerr
			}
		}
	}
	//
	return nil, writeerr
}

func (UPSDC *TUPSDClientConnection) Login(username string, password string) (err error) {
	// Sending username
	resp, trerr := UPSDC.transactionRAW(fmt.Sprintf("%s %s", kwUSERNAME, username))
	if trerr == nil {
		if resp[0] == kwOK {
			// Sending password
			resp, trerr = UPSDC.transactionRAW(fmt.Sprintf("%s %s", kwPASSWORD, password))
			if trerr == nil {
				if resp[0] == kwOK {
					return UPSDC.loadUPSlist()
				} else {
					return fmt.Errorf("login failed: %v", resp[0])
				}
			}
		} else {
			return fmt.Errorf("login failed: %v", resp[0])
		}
	}
	//
	return trerr
}

func breakListLine(listline string, rex string) []string {
	re := regexp.MustCompile(rex)
	found := re.FindAllStringSubmatch(listline, -1)
	if found != nil {
		return found[0][1:]
	}
	return nil
}

func (UPSDC *TUPSDClientConnection) loadUPSlist() (err error) {
	resp, trerr := UPSDC.transactionRAW(fmt.Sprintf("%s %s", kwLIST, kwUPS))
	if trerr == nil {
		for ri := range resp {
			upsdetails := breakListLine(resp[ri], reLISTLINE3F)
			if upsdetails != nil {
				if upsdetails[0] == kwUPS {
					tups := TUPS{Description: upsdetails[2]}
					UPSDC.ups[upsdetails[1]] = tups
					UPSDC.defups = upsdetails[1]
				}
			}
		}
	}
	//
	return trerr
}

func (UPSDC TUPSDClientConnection) getFloatValue(valuename string) (v float64) {
	resp, trerr := UPSDC.transactionRAW(fmt.Sprintf("%s %s %s %s", kwGET, kwVAR, UPSDC.defups, valuename))
	if (trerr == nil) && (len(resp) == 1) {
		vardetails := breakListLine(resp[0], reLISTLINE4F)
		if len(vardetails) == 4 {
			f, _ := strconv.ParseFloat(vardetails[3], 64)
			return f
		}
	}
	//
	return 0
}

func (UPSDC TUPSDClientConnection) getStringValue(valuename string) (v string) {
	resp, trerr := UPSDC.transactionRAW(fmt.Sprintf("%s %s %s %s", kwGET, kwVAR, UPSDC.defups, valuename))
	if (trerr == nil) && (len(resp) == 1) {
		vardetails := breakListLine(resp[0], reLISTLINE4F)
		if len(vardetails) == 4 {
			return vardetails[3]
		}
	}
	//
	return ""
}

func (UPSDC TUPSDClientConnection) IsPowerMainsGood() bool {
	voltage := UPSDC.getFloatValue(varVOLTAGE)
	// return (UPSDC.getFloatValue(varTRANSFERLO) < voltage) && (voltage < UPSDC.getFloatValue(varTRANSFERHI))
	return voltage > 0
}

func (UPSDC *TUPSDClientConnection) GetStatusOn(st string) bool {
	spst := strings.Split(UPSDC.getStringValue(varUPSSTATUS), " ")
	if len(spst) > 0 {
		for i := range spst {
			if spst[i] == st {
				return true
			}
		}
	}
	return false
}

func (UPSDC TUPSDClientConnection) ScheduleLoadOffIn(seconds uint) (err error) {
	_, err = UPSDC.transactionRAW(fmt.Sprintf("%s %s %s %d", kwINSTCMD, UPSDC.defups, cmdLOADOFFDELAY, seconds))
	return err
}
