package req

import "sync"

const copyBufPoolSize = 32 * 1024

var copyBufPool = sync.Pool{New: func() any { return new([copyBufPoolSize]byte) }}

func getCopyBuf() []byte {
	return copyBufPool.Get().(*[copyBufPoolSize]byte)[:]
}

func putCopyBuf(b []byte) {
	if len(b) != copyBufPoolSize {
		panic("trying to put back buffer of the wrong size in the copyBufPool")
	}
	copyBufPool.Put((*[copyBufPoolSize]byte)(b))
}
