//go:build windows && amd64

package winmic

/* This is what lack of packed structs in Go gets you into!

typedef struct tWAVEFORMATEX {
  WORD  wFormatTag;			// 2 bytes @ 0
  WORD  nChannels;			// 2 bytes @ 2
  DWORD nSamplesPerSec;		// 4 bytes @ 4
  DWORD nAvgBytesPerSec;	// 4 bytes @ 8
  WORD  nBlockAlign;		// 2 bytes @ 12
  WORD  wBitsPerSample;	 	// 2 bytes @ 14
  WORD  cbSize;				// 2 bytes @ 16
							// 18 bytes total
}
*/

import (
	"unsafe"
)

type WAVEFORMATEX struct {
	Storage [18]byte
}

// WORD at given offset
func (wfex *WAVEFORMATEX) getWORDAtOffset(o uintptr) uint16 {
	return *(*uint16)(unsafe.Pointer(&(wfex.Storage[o])))
}

func (wfex *WAVEFORMATEX) setWORDAtOffset(o uintptr, value uint16) {
	*(*uint16)(unsafe.Pointer(&(wfex.Storage[o]))) = value
}

// DWORD at given offset
func (wfex *WAVEFORMATEX) getDWORDAtOffset(o uintptr) uint32 {
	return *(*uint32)(unsafe.Pointer(&(wfex.Storage[o])))
}

func (wfex *WAVEFORMATEX) setDWORDAtOffset(o uintptr, value uint32) {
	*(*uint32)(unsafe.Pointer(&(wfex.Storage[o]))) = value
}

// wFormatTag
func (wfex *WAVEFORMATEX) setWFormatTag(wFormatTag uint16) {
	wfex.setWORDAtOffset(0, wFormatTag)
}

// nChannels
func (wfex *WAVEFORMATEX) getNChannels() uint16 {
	return wfex.getWORDAtOffset(2)
}

func (wfex *WAVEFORMATEX) setNChannels(nChannels uint16) {
	wfex.setWORDAtOffset(2, nChannels)
}

// nSamplesPerSec
func (wfex *WAVEFORMATEX) getNSamplesPerSec() uint32 {
	return wfex.getDWORDAtOffset(4)
}

func (wfex *WAVEFORMATEX) setNSamplesPerSec(nSamplesPerSec uint32) {
	wfex.setDWORDAtOffset(4, nSamplesPerSec)
}

// nAvgBytesPerSec
func (wfex *WAVEFORMATEX) setNAvgBytesPerSec(nAvgBytesPerSec uint32) {
	wfex.setDWORDAtOffset(8, nAvgBytesPerSec)
}

// nBlockAlign
func (wfex *WAVEFORMATEX) getNBlockAlign() uint16 {
	return wfex.getWORDAtOffset(12)
}

func (wfex *WAVEFORMATEX) setNBlockAlign(nBlockAlign uint16) {
	wfex.setWORDAtOffset(12, nBlockAlign)
}

// wBitsPerSample
func (wfex *WAVEFORMATEX) getWBitsPerSample() uint16 {
	return wfex.getWORDAtOffset(14)
}

func (wfex *WAVEFORMATEX) setWBitsPerSample(wBitsPerSample uint16) {
	wfex.setWORDAtOffset(14, wBitsPerSample)
}

// cbSize
func (wfex *WAVEFORMATEX) setCbSize(cbSize uint16) {
	wfex.setWORDAtOffset(16, cbSize)
}

// Auto-complete fields
func (wfex *WAVEFORMATEX) Complete() {
	wfex.setWFormatTag(wave_FORMAT_PCM)
	wfex.setNBlockAlign((wfex.getNChannels() * (wfex.getWBitsPerSample()) / 8))
	wfex.setNAvgBytesPerSec(wfex.getNSamplesPerSec() * uint32(wfex.getNBlockAlign()))
	wfex.setCbSize(0)
}
