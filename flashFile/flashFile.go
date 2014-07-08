package flashFile

import "github.com/mikispag/rosettaflash/zlibStream"

type FlashFile struct {
	signature   []byte
	fileVersion []byte
	fileLength  []byte
	data        zlibStream.ZlibStream
}

func (f *FlashFile) WriteSignature() []byte {
	return []byte("CWS")
}

func (f *FlashFile) WriteFileVersion() []byte {
	return []byte("M")
}

func (f *FlashFile) WriteFileLength() []byte {
	// First three bytes are ignored, last byte shouldn't be
	// too high for this to work on PepperFlash in Chrome.
	return []byte("IKI0")
}

func (f *FlashFile) GetBytes() []byte {
	bytes := make([]byte, 0, 1024*128)
	bytes = append(bytes, f.signature...)
	bytes = append(bytes, f.fileVersion...)
	bytes = append(bytes, f.fileLength...)
	bytes = append(bytes, f.data.GetBytes()...)
	return bytes
}

func (f *FlashFile) WriteFile(stream *zlibStream.ZlibStream) {
	f.signature = f.WriteSignature()
	f.fileVersion = f.WriteFileVersion()
	f.fileLength = f.WriteFileLength()
	f.data = *stream
}
