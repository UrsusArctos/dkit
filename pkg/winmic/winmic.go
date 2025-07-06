//go:build windows && amd64

package winmic

// extern void GoWIPCallBack(uintptr_t hwi, uintptr_t uMsg, void* dwInstance, void* dwParam1, void* dwParam2);
// static void CWIPCallBack(uintptr_t hwi, uintptr_t uMsg, void* dwInstance, void* dwParam1, void* dwParam2) {
//   GoWIPCallBack(hwi,uMsg,dwInstance,dwParam1,dwParam2);
// }
// static void* GetCWIPCBackPtr() {
//	return &CWIPCallBack;
// }
import "C"

import (
	"fmt"
	"sync"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

const (
	dllName               = "winmm.dll"
	waveInGetNumDevs      = "waveInGetNumDevs"
	waveInGetDevCaps      = "waveInGetDevCapsW"
	waveInOpen            = "waveInOpen"
	waveInClose           = "waveInClose"
	waveInPrepareHeader   = "waveInPrepareHeader"
	waveInUnprepareHeader = "waveInUnprepareHeader"
	waveInAddBuffer       = "waveInAddBuffer"
	waveInStart           = "waveInStart"
	waveInReset           = "waveInReset"
	waveInStop            = "waveInStop"
	WAVE_MAPPER           = -1
	//
	sampleWidth = 16
	//
	nBuffers    = 2
	syncBacklog = 0x100
)

type (
	// We will only ever record in signed 16bit little-endian samples
	TRecordingBuffer = []int16

	TRecordingHandler func(buf *TRecordingBuffer)

	TWinMicrophone struct {
		// Handles
		hLib    *syscall.LazyDLL
		hWaveIn uintptr
		// Sync
		recSignal chan bool
		Handler   TRecordingHandler
		wgroup    sync.WaitGroup
		// Data
		wfex     WAVEFORMATEX
		stopFlag bool
		// Recording buffers
		recBufIndex uint8
		recBuf      [nBuffers]TRecordingBuffer
		waveHDR     [nBuffers]WAVEHDR
		storedBuf   TRecordingBuffer
	}
)

func NewInstance() TWinMicrophone {
	return TWinMicrophone{hLib: syscall.NewLazyDLL(dllName), recSignal: make(chan bool, syncBacklog), Handler: nil}
}

// DLL function helper
func (wmic *TWinMicrophone) callProc(procname string, procargs ...uintptr) (ret0 uintptr, ret1 uintptr, err syscall.Errno) {
	return syscall.SyscallN(wmic.hLib.NewProc(procname).Addr(), procargs...)
}

func (wmic *TWinMicrophone) GetNumDevs() uint32 {
	ret0, _, _ := wmic.callProc(waveInGetNumDevs)
	return uint32(ret0)
}

func (wmic *TWinMicrophone) GetDevName(devnum uint32) string {
	if devnum < wmic.GetNumDevs() {
		var pwic WAVEINCAPS
		ret0, _, _ := wmic.callProc(waveInGetDevCaps,
			uintptr(devnum),
			uintptr(unsafe.Pointer(&pwic)),
			uintptr(unsafe.Sizeof(pwic)))
		if ret0 == 0 {
			return string(utf16.Decode(pwic.SzPname[:]))
		}
	}
	return ""
}

func (wmic *TWinMicrophone) SetAudioFormat(nChans uint16, nFreq uint32) {
	wmic.wfex.setNChannels(nChans)
	wmic.wfex.setNSamplesPerSec(nFreq)
	wmic.wfex.setWBitsPerSample(sampleWidth)
	wmic.wfex.Complete()
}

func (wmic *TWinMicrophone) Open(devnum int32) error {
	ret0, _, _ := wmic.callProc(waveInOpen,
		uintptr(unsafe.Pointer(&wmic.hWaveIn)),
		uintptr(devnum),
		uintptr(unsafe.Pointer(&wmic.wfex)),
		uintptr(C.GetCWIPCBackPtr()),
		uintptr(unsafe.Pointer(wmic)),
		uintptr(callBACK_FUNCTION))
	if ret0 != 0 {
		wmic.hWaveIn = 0
		return fmt.Errorf("cannot open, error 0x%x (%d)", ret0, ret0)
	}
	return nil
}

func (wmic *TWinMicrophone) Close() {
	if wmic.IsOpened() {
		wmic.callProc(waveInClose, uintptr(wmic.hWaveIn))
	}
	wmic.hWaveIn = 0
}

func (wmic *TWinMicrophone) IsOpened() bool {
	return wmic.hWaveIn != 0
}

// WaveIn CALLBACK bridge ===

//export GoWIPCallBack
func GoWIPCallBack(hwi C.uintptr_t, uMsg C.uintptr_t, dwInstance *C.void, dwParam1 *C.void, dwParam2 *C.void) {
	(*TWinMicrophone)(unsafe.Pointer(dwInstance)).waveInCallBack(uMsg)
}

// WaveIn CALLBACK bridge ===

func (wmic *TWinMicrophone) waveInCallBack(uMsg C.uintptr_t) {
	switch uMsg {
	case mm_WIM_OPEN: // do nothing
	case mm_WIM_CLOSE: // do nothing
	case mm_WIM_DATA:
		{
			// unqueue current buffer from device
			wmic.unqueueRecBuffer(wmic.recBufIndex)
			// store recorded content into storege buffer
			copy(wmic.storedBuf, wmic.recBuf[wmic.recBufIndex])
			// signal for Sync Loop to call storage buffer handler
			wmic.recSignal <- true
			// queue this same buffer again
			if !wmic.stopFlag {
				wmic.queueRecBuffer(wmic.recBufIndex)
				// make sure next time the index is different
				wmic.recBufIndex = (wmic.recBufIndex + 1) % nBuffers
			}
		}
	}
}

func (wmic *TWinMicrophone) AllocateRecordingBuffers(SamplesInBuffer uint32) {
	for bufidx := range wmic.recBuf {
		wmic.recBuf[bufidx] = make(TRecordingBuffer, SamplesInBuffer)
	}
	wmic.storedBuf = make(TRecordingBuffer, SamplesInBuffer)
}

// Duration of the recording buffer in seconds
func (wmic *TWinMicrophone) GetBufferDuration() float64 {
	return float64(len(wmic.recBuf[0])) / (float64(wmic.wfex.getNChannels()) * float64(wmic.wfex.getNSamplesPerSec()))
}

func (wmic *TWinMicrophone) initWaveHDR(bufIdx uint8) {
	wmic.waveHDR[bufIdx] = WAVEHDR{
		LpData:         &wmic.recBuf[bufIdx][0],
		DwBufferLength: uint32(len(wmic.recBuf[bufIdx])) * uint32(unsafe.Sizeof(wmic.recBuf[bufIdx][0])),
		DwFlags:        0,
	}
}

func (wmic *TWinMicrophone) unqueueRecBuffer(bufIdx uint8) {
	wmic.callProc(waveInUnprepareHeader,
		uintptr(wmic.hWaveIn),
		uintptr(unsafe.Pointer(&wmic.waveHDR[bufIdx])),
		uintptr(unsafe.Sizeof(wmic.waveHDR[bufIdx])))
}

func (wmic *TWinMicrophone) queueRecBuffer(bufIdx uint8) {
	wmic.initWaveHDR(bufIdx)
	ret0, _, _ := wmic.callProc(waveInPrepareHeader,
		uintptr(wmic.hWaveIn),
		uintptr(unsafe.Pointer(&wmic.waveHDR[bufIdx])),
		uintptr(unsafe.Sizeof(wmic.waveHDR[bufIdx])))
	if ret0 == 0 {
		wmic.callProc(waveInAddBuffer,
			uintptr(wmic.hWaveIn),
			uintptr(unsafe.Pointer(&wmic.waveHDR[bufIdx])),
			uintptr(unsafe.Sizeof(wmic.waveHDR[bufIdx])))
	}
}

func (wmic *TWinMicrophone) StartRecording() error {
	for bufidx := range wmic.recBuf {
		wmic.queueRecBuffer(uint8(bufidx))
	}
	ret0, _, err := wmic.callProc(waveInStart, uintptr(wmic.hWaveIn))
	wmic.stopFlag = (ret0 != 0)
	if ret0 == 0 {
		wmic.recBufIndex = 0
		wmic.wgroup.Add(1)
		go wmic.syncLoop()
		return nil
	} else {
		return err
	}
}

func (wmic *TWinMicrophone) StopRecording() {
	wmic.stopFlag = true
	wmic.callProc(waveInStop, uintptr(wmic.hWaveIn))
	wmic.wgroup.Wait()
}

func (wmic *TWinMicrophone) syncLoop() {
	for !wmic.stopFlag {
		<-wmic.recSignal
		if wmic.Handler != nil {
			wmic.Handler(&wmic.storedBuf)
		}
	}
	wmic.wgroup.Done()
}
