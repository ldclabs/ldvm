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

	"github.com/klauspost/compress/zstd"

	"github.com/ldclabs/ldvm/util/sync"
)

var (
	zstdReaderPool sync.Pool[*zstd.Decoder]
	zstdWriterPool sync.Pool[*zstd.Encoder]
)

// zstdReader is an optimized wrapper of zstd.Decoder with sync.Pool support.
type ZstdReader struct {
	r    io.Reader     // underlying Reader stream
	zr   *zstd.Decoder // lazily-initialized zstd reader
	zerr error         // any error from zstd.NewReader
}

// NewZstdReader returns a new zstd Reader with the given io.Reader.
func NewZstdReader(r io.Reader) *ZstdReader {
	return &ZstdReader{r: r}
}

// Read reads data from the underlying io.Reader and decompresses it.
func (zr *ZstdReader) Read(p []byte) (n int, err error) {
	if zr.zerr != nil {
		return 0, zr.zerr
	}
	if zr.zr == nil {
		if zr.zerr == nil {
			reader, ok := zstdReaderPool.Get()
			if ok {
				zr.zerr = reader.Reset(zr.r)
				zr.zr = reader
			} else {
				zr.zr, zr.zerr = zstd.NewReader(zr.r,
					zstd.WithDecoderLowmem(false),
					zstd.WithDecoderMaxWindow(16<<20),
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

// Reset will reset the decoder the supplied stream after the current has finished processing,
// it does not close the underlying io.Writer.
func (zr *ZstdReader) Reset() {
	if zr.zr != nil {
		zr.zr.Reset(nil)
		zstdReaderPool.Put(zr.zr)
		zr.zr = nil
	}
}

// Close will reset the decoder the supplied stream after the current has finished processing,
// and close the underlying io.Reader if it implemented io.Closer.
// The Reader can not be reused after Close.
func (zr *ZstdReader) Close() error {
	zr.Reset()
	if c, ok := zr.r.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// ZstdWriter is an optimized wrapper of zstd.Encoder with sync.Pool support.
type ZstdWriter struct {
	w    io.Writer     // underlying Writer stream
	zw   *zstd.Encoder // lazily-initialized zstd writer
	zerr error         // any error from zstd.NewReader
}

// NewZstdWriter returns a new zstd Writer with the given io.Writer.
func NewZstdWriter(w io.Writer) *ZstdWriter {
	return &ZstdWriter{w: w}
}

// Write compresses data and writes it to the underlying io.Writer.
func (zw *ZstdWriter) Write(p []byte) (n int, err error) {
	if zw.zerr != nil {
		return 0, zw.zerr
	}
	if zw.zw == nil {
		if zw.zerr == nil {
			writer, ok := zstdWriterPool.Get()
			if ok {
				writer.Reset(zw.w)
				zw.zw = writer
			} else {
				zw.zw, zw.zerr = zstd.NewWriter(zw.w,
					zstd.WithLowerEncoderMem(false),
					zstd.WithWindowSize(16<<20),
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
		// stash the encoder and keep the error.
		zw.Reset()
		zw.zerr = err
	}
	return
}

// Reset closes the Writer, flushing any unwritten data to the underlying io.Writer.
// but does not close the underlying io.Writer.
// The Writer can not be reused after Reset.
func (zw *ZstdWriter) Reset() {
	if zw.zw != nil {
		zw.zw.Close()
		zw.zw.Reset(nil)
		zstdWriterPool.Put(zw.zw)
		zw.zw = nil
	}
}

// Close closes the Writer, flushing any unwritten data to the underlying io.Writer.
// and close the underlying io.Writer if it implemented io.Closer.
// The Writer can not be reused after Close.
func (zw *ZstdWriter) Close() error {
	zw.Reset()
	if c, ok := zw.w.(io.Closer); ok {
		return c.Close()
	}
	return nil
}
