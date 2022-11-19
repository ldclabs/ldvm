// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
//
// This file is a derived work, based on the https://github.com/klauspost/compress/blob/master/gzhttp/transport.go whose original
// notices appear below.
//
// **********
// Copyright (c) 2021 Klaus Post. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package compress

import (
	"io"
	"sync"

	"github.com/klauspost/compress/zstd"
)

// zstdReaderPool pools zstd decoders.
var zstdReaderPool sync.Pool

// zstdReader wraps a response body so it can lazily
// call zstd.NewReader on the first call to Read
type ZstdReader struct {
	R    io.Reader     // underlying Reader stream
	zr   *zstd.Decoder // lazily-initialized zstd reader
	zerr error         // any error from zstd.NewReader
}

func (zr *ZstdReader) Read(p []byte) (n int, err error) {
	if zr.zerr != nil {
		return 0, zr.zerr
	}
	if zr.zr == nil {
		if zr.zerr == nil {
			reader, ok := zstdReaderPool.Get().(*zstd.Decoder)
			if ok {
				zr.zerr = reader.Reset(zr.R)
				zr.zr = reader
			} else {
				zr.zr, zr.zerr = zstd.NewReader(zr.R,
					zstd.WithDecoderLowmem(true),
					zstd.WithDecoderMaxWindow(1<<20),
					zstd.WithDecoderConcurrency(1))
			}
		}
		if zr.zerr != nil {
			return 0, zr.zerr
		}
	}
	n, err = zr.zr.Read(p)
	if err != nil {
		// Usually this will be io.EOF,
		// stash the decoder and keep the error.
		zr.Reset()
		zr.zerr = err
	}
	return
}

func (zr *ZstdReader) Close() error {
	zr.Reset()
	if c, ok := zr.R.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func (zr *ZstdReader) Reset() {
	if zr.zr != nil {
		zr.zr.Reset(nil)
		zstdReaderPool.Put(zr.zr)
		zr.zr = nil
	}
}

// zstdReaderPool pools zstd decoders.
var zstdWriterPool sync.Pool

// zstdReader wraps a response body so it can lazily
// call zstd.NewReader on the first call to Read
type ZstdWriter struct {
	W    io.Writer     // underlying Writer stream
	zw   *zstd.Encoder // lazily-initialized zstd writer
	zerr error         // any error from zstd.NewReader
}

func (zw *ZstdWriter) Write(p []byte) (n int, err error) {
	if zw.zerr != nil {
		return 0, zw.zerr
	}
	if zw.zw == nil {
		if zw.zerr == nil {
			writer, ok := zstdWriterPool.Get().(*zstd.Encoder)
			if ok {
				writer.Reset(zw.W)
				zw.zw = writer
			} else {
				zw.zw, zw.zerr = zstd.NewWriter(zw.W,
					zstd.WithLowerEncoderMem(true),
					zstd.WithWindowSize(1<<20),
					zstd.WithEncoderConcurrency(1))
			}
		}
		if zw.zerr != nil {
			return 0, zw.zerr
		}
	}
	n, err = zw.zw.Write(p)
	if err != nil {
		// Usually this will be io.EOF,
		// stash the decoder and keep the error.
		zw.Reset()
		zw.zerr = err
	}
	return
}

func (zw *ZstdWriter) Close() error {
	zw.Reset()
	if c, ok := zw.W.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func (zw *ZstdWriter) Reset() {
	if zw.zw != nil {
		zw.zw.Close()
		zw.zw.Reset(nil)
		zstdReaderPool.Put(zw.zw)
		zw.zw = nil
	}
}
