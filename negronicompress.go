// Copyright 2016 Igor "Mocheryl" Zornik. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package negronicompress

import (
	`compress/flate`
	`compress/gzip`
	`io`
	`net/http`
	`regexp`
	`strconv`
	`strings`

	`github.com/codegangsta/negroni`
)

const (
	headerAcceptEncoding  string = `Accept-Encoding`
	headerContentEncoding string = `Content-Encoding`
	headerContentLength   string = `Content-Length`
	headerContentType     string = `Content-Type`
	headerDeflate         string = `deflate`
	headerGzip            string = `gzip`
	headerVary            string = `Vary`
	// Minimum data size in bytes the response body must have in order to be
	// considered for compression.
	mininumContentLength int = 2048
)

// compressRegEx is a regular expression for supported content encoding types
// that are checked against clients supported encoding types.
var compressRegEx = regexp.MustCompile(`(,` + headerGzip + `,|,` + headerDeflate + `,)`)

// compressResponseWriter is the ResponseWriter that negroni.ResponseWriter is
// wrapped in.
type compressResponseWriter struct {
	c []byte
	negroni.ResponseWriter
}

// Write appends any data to writers buffer.
func (m *compressResponseWriter) Write(b []byte) (int, error) {
	m.c = append(m.c, b...)
	return len(b), nil
}

// compress sends any output content back to client in a compressed format
// whenever possible.
type compress struct {
	// compressionLevel is the level of compression that should be performed on
	// the output content where higher level means better compression, but
	// longer processing time, while lower level is faster but yields lesser
	// compressed content.
	compressionLevel int
	// compressiableFileTypes is a list of file types that should be compressed.
	compressiableFileTypes []string
	// compressContentTypeRegEx is a list of file types that should be
	// compressed compiled into a regular expression.
	compressContentTypeRegEx *regexp.Regexp
}

// NewCompress returns a new compress middleware instance with default
// compression level set.
func NewCompress() *compress {
	return NewCompressWithCompressionLevel(flate.DefaultCompression)
}

// NewCompress returns a new compress middleware instance.
func NewCompressWithCompressionLevel(level int) *compress {
	return &compress{level, compressiableFileTypes, compressContentTypeRegEx}
}

// AddContentType adds a new file type to the middleware list of file types that
// can be compressed. c should match the form used of a value used in
// "Content-Type" HTTP header. If c is "*/*", it will reset the list to empty
// value making it match all types, including no type.
func (h *compress) AddContentType(c ...string) (err error) {
	// XXX: This is copied from the helper function. Somehow remove code
	// duplication. With pointers maybe?
	cList := h.compressiableFileTypes
	for _, t := range c {
		cList, err = appendFileType(cList, t)
		if err != nil {
			return
		}
	}

	// Compile new regular expression.
	// TODO: Compile only if slice has changed.
	var r *regexp.Regexp
	r, err = compileFileTypes(cList)
	if err != nil {
		return
	}

	h.compressiableFileTypes = cList
	h.compressContentTypeRegEx = r

	return
}

func (h *compress) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// Notify the user agent that we support content compression.
	if rw.Header().Get(headerVary) == `` {
		rw.Header().Set(headerVary, headerAcceptEncoding)
	}

	// Skip compression if content is already encoded.
	if rw.Header().Get(headerContentEncoding) != `` {
		next(rw, r)
		return
	}

	// Check if client supports any kind of content compression in response. Do
	// nothing and exit function if it doesn't.
	acceptedEncoding := r.Header.Get(headerAcceptEncoding)
	if acceptedEncoding == `` || !compressRegEx.MatchString(`,`+acceptedEncoding+`,`) {
		next(rw, r)
		return
	}

	// Wrap the original writer with a buffered one.
	crw := &compressResponseWriter{
		make([]byte, 0),
		negroni.NewResponseWriter(rw),
	}
	defer func() {
		crw.c = []byte{}
	}()
	next(crw, r)

	// Compress only if output content will benefit from compression and if we
	// are allowed to compress the output content type.
	if len(crw.c) > mininumContentLength && h.compressContentTypeRegEx.MatchString(crw.Header().Get(headerContentType)) {
		var (
			wc  io.WriteCloser
			old []byte
		)
		old, crw.c = crw.c, []byte{}
		for _, t := range strings.Split(acceptedEncoding, `,`) {
			// Find compression method with highest priority.
			switch t {
			case headerGzip:
				// TODO: Error checking.
				wc, _ = gzip.NewWriterLevel(crw, h.compressionLevel)
			case headerDeflate:
				// TODO: Error checking.
				wc, _ = flate.NewWriter(crw, h.compressionLevel)
			default:
				// TODO: Make a test for this.
			}
			// Check if any of the supported compression methods were found.
			if wc != nil {
				// Set response compression encoding based on the supported type
				// we found.
				rw.Header().Set(headerContentEncoding, t)
				// TODO: Error checking.
				wc.Write(old)
				old = []byte{}
				wc.Close()
				// Set size of the compressed content.
				rw.Header().Set(headerContentLength, strconv.FormatInt(int64(len(crw.c)), 10))
				break
			}
		}
	}

	// TODO: Error checking.
	rw.Write(crw.c)
}
