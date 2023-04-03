//go:build windows && amd64

package winmic

const (
	wave_FORMAT_PCM   = 1
	maxPNAMELEN       = 32
	callBACK_FUNCTION = 0x30000
	mm_WIM_OPEN       = 0x03BE
	mm_WIM_CLOSE      = 0x03BF
	mm_WIM_DATA       = 0x03C0
)

type (
	// This structure is not packed
	WAVEINCAPS struct {
		WMid           uint16
		WPid           uint16
		VDriverVersion uint32
		SzPname        [maxPNAMELEN]uint16
		DwFormats      uint32
		WChannels      uint16
		WReserved1     uint16
	}

	// This structure is not packed either
	WAVEHDR struct {
		LpData          *int16 // *byte
		DwBufferLength  uint32
		DwBytesRecorded uint32
		DwUser          *uint32
		DwFlags         uint32
		DwLoops         uint32
		LpNext          *WAVEHDR
		Reserved        *uint32
	}
)
