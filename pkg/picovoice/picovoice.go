//go:build windows && amd64

package picovoice

import (
	"C"
	"syscall"
	"unsafe"

	porcupine "github.com/Picovoice/porcupine/binding/go/v2"
)

const (
	dllName            = "libpv_cobra.dll"
	pvSampleRate       = "pv_sample_rate"
	pvCobraInit        = "pv_cobra_init"
	pvCobraDelete      = "pv_cobra_delete"
	pvCobraFrameLength = "pv_cobra_frame_length"
	pvCobraProcess     = "pv_cobra_process"
)

type (
	// Must be initialized to GetFrameLength()
	TPVRecordingFrame []int16

	// PV Instance
	TPicovoice struct {
		// internal data
		accessKey string
		hLib      *syscall.LazyDLL
		// subinstances
		pvPorcupine porcupine.Porcupine
		pvCobra     uintptr
	}
)

// Constructor
func NewInstance(aKey string) TPicovoice {
	return TPicovoice{accessKey: aKey, hLib: syscall.NewLazyDLL(dllName)}
}

// DLL function helper
func (pv TPicovoice) callProc(procname string, procargs ...uintptr) (ret0 uintptr, ret1 uintptr, err syscall.Errno) {
	return syscall.SyscallN(pv.hLib.NewProc(procname).Addr(), procargs...)
}

func (pv *TPicovoice) CreatePorcupine(modelFile string, keywordFile []string) error {
	pv.pvPorcupine = porcupine.Porcupine{AccessKey: pv.accessKey, ModelPath: modelFile, KeywordPaths: keywordFile}
	return pv.pvPorcupine.Init()
}

func (pv *TPicovoice) ClosePorcupine() error {
	return pv.pvPorcupine.Delete()
}

func (pv *TPicovoice) CreateCobra() porcupine.PvStatus {
	akey := []byte(pv.accessKey)
	r0, _, _ := pv.callProc(pvCobraInit,
		uintptr(unsafe.Pointer(&akey[0])),
		uintptr(unsafe.Pointer(&pv.pvCobra)))
	return porcupine.PvStatus(r0)
}

func (pv *TPicovoice) CloseCobra() error {
	r0, _, cerr := pv.callProc(pvCobraDelete, uintptr(pv.pvCobra))
	if r0 == 0 {
		return nil
	}
	return cerr
}

func (pv TPicovoice) GetSampleRate() uint32 {
	r0, _, _ := pv.callProc(pvSampleRate)
	return uint32(r0)
}

func (pv TPicovoice) GetFrameLength() uint32 {
	r0, _, _ := pv.callProc(pvCobraFrameLength)
	return uint32(r0)
}

// This function processes frame as Porcupine and Cobra
func (pv TPicovoice) ProcessFrame(frame *TPVRecordingFrame) (int, C.float, error) {
	kwindex, err := pv.pvPorcupine.Process(*frame)
	if err == nil {
		vprob, perr := pv.ProcessCobraFrame(frame)
		return kwindex, vprob, perr
	}
	return kwindex, 0.0, err
}

// This function processes frame as Cobra only
func (pv TPicovoice) ProcessCobraFrame(frame *TPVRecordingFrame) (C.float, error) {
	var vprob C.float
	r0, _, perr := pv.callProc(pvCobraProcess,
		pv.pvCobra,
		uintptr(unsafe.Pointer(&((*frame)[0]))),
		uintptr(unsafe.Pointer(&vprob)))
	if r0 == 0 {
		return vprob, nil
	}
	return 0.0, perr
}
