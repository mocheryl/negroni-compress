// Copyright 2016 Igor "Mocheryl" Zornik. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package negronicompress

import (
	`compress/flate`
	`net/http`
	`net/http/httptest`
	`regexp`
	`testing`

	`github.com/codegangsta/negroni`
)

type respWriteTest struct {
	h http.Header
}

func (w *respWriteTest) Header() http.Header {
	return w.h
}

func (w *respWriteTest) WriteHeader(code int) {}

func (w *respWriteTest) Write(data []byte) (n int, err error) {
	return
}

func TestCompressResponseWriter_Write(t *testing.T) {
	rw := &respWriteTest{
		h: make(http.Header),
	}
	nrw := negroni.NewResponseWriter(rw)
	crw := &compressResponseWriter{
		make([]byte, 0),
		nrw,
	}
	if n, err := crw.Write([]byte(`test`)); n != 4 || err != nil {
		t.Errorf(`negronicompress.compressResponseWriter.Write(%s) = %d, %v; want %d, nil`, []byte(`test`), n, err, 4)
	}
	if crw.c[0] != []byte(`t`)[0] {
		t.Errorf(`negronicompress.compressResponseWriter.c[0] = %v, want %v`, crw.c[0], []byte(`t`)[0])
	}
	if crw.c[1] != []byte(`e`)[0] {
		t.Errorf(`negronicompress.compressResponseWriter.c[1] = %v, want %v`, crw.c[1], []byte(`e`)[0])
	}
	if crw.c[2] != []byte(`s`)[0] {
		t.Errorf(`negronicompress.compressResponseWriter.c[2] = %v, want %v`, crw.c[2], []byte(`s`)[0])
	}
	if crw.c[3] != []byte(`t`)[0] {
		t.Errorf(`negronicompress.compressResponseWriter.c[3] = %v, want %v`, crw.c[3], []byte(`t`)[0])
	}
}

func TestNewCompress(t *testing.T) {
	handler := NewCompress()
	if handler == nil {
		t.Fatal(`negronicompress.NewCompress() cannot return nil`)
	}

	if handler.compressionLevel != flate.DefaultCompression {
		t.Errorf(`negronicompress.NewCompress().compressionLevel = %d, want %d`, handler.compressionLevel, flate.DefaultCompression)
	}
}

func TestNewCompressWithCompressionLevel(t *testing.T) {
	handler := NewCompressWithCompressionLevel(1)
	if handler == nil {
		t.Fatal(`negronicompress.NewCompressWithCompressionLevel() cannot return nil`)
	}

	if handler.compressionLevel != 1 {
		t.Errorf(`negronicompress.NewCompressWithCompressionLevel().compressionLevel = %d, want %d`, handler.compressionLevel, 1)
	}

	if l, e := len(handler.compressiableFileTypes), len(compressiableFileTypes); l != e {
		t.Fatalf(`len(negronicompress.NewCompressWithCompressionLevel().compressiableFileTypes) = %d, want %d`, l, e)
	}
	for i := range handler.compressiableFileTypes {
		if handler.compressiableFileTypes[i] != compressiableFileTypes[i] {
			t.Errorf(`negronicompress.NewCompressWithCompressionLevel().compressiableFileTypes[%d] = %q, want %q`, i, handler.compressiableFileTypes[i], compressiableFileTypes[i])
		}
	}

	if r, e := handler.compressContentTypeRegEx.String(), compressContentTypeRegEx.String(); r != e {
		t.Errorf(`negronicompress.NewCompressWithCompressionLevel().compressContentTypeRegEx.String() = %q, want %q`, r, e)
	}
}

func TestCompress_AddContentType(t *testing.T) {
	handler := NewCompress()
	if handler == nil {
		t.Fatal(`negronicompress.NewCompress() cannot return nil`)
	}

	cLen, cOrig, cExOrig := len(handler.compressiableFileTypes), handler.compressiableFileTypes, *contentTypeRegEx
	if err := handler.AddContentType(`xyz`); err == nil {
		t.Errorf(`negronicompress.AddContentType(%q) = nil, want err`, `xyz`)
	}

	if err := handler.AddContentType(`application/octet-stream`); err != nil {
		t.Fatalf(`negronicompress.AddContentType(%q) = %v, want nil`, `application/octet-stream`, err)
	}
	if l := len(handler.compressiableFileTypes); l != cLen+1 {
		t.Fatalf(`len(negronicompress.compressiableFileTypes) = %d, want %d`, l, cLen+1)
	}
	for i := range cOrig {
		if cOrig[i] != handler.compressiableFileTypes[i] {
			t.Errorf(`negronicompress.compressiableFileTypes[%d] = %q, want %q`, i, handler.compressiableFileTypes[i], cOrig[i])
		}
	}
	if handler.compressiableFileTypes[cLen] != `application/octet-stream` {
		t.Errorf(`negronicompress.compressiableFileTypes[%d] = %q, want %q`, cLen, handler.compressiableFileTypes[cLen], `application/octet-stream`)
	}
	if r := handler.compressContentTypeRegEx.String(); r != `^(text/.+|application/x-javascript|application/xhtml+xml|application/octet-stream)$` {
		t.Errorf(`negronicompress.compressContentTypeRegEx.String() = %q, want %q`, r, `^(text/.+|application/x-javascript|application/xhtml+xml|application/octet-stream)$`)
	}

	contentTypeRegEx = regexp.MustCompile(`.*`)
	if err := handler.AddContentType(`\x`); err == nil {
		t.Errorf(`negronicompress.AddContentType(%q) = nil, want err`, `\x`)
	}
	contentTypeRegEx = &cExOrig
}

func TestCompress_ServeHTTP(t *testing.T) {
	cnt := ``
	for i := 0; i <= mininumContentLength; i++ {
		cnt += `.`
	}

	w := httptest.NewRecorder()
	if _, err := w.Write([]byte(cnt)); err != nil {
		t.Fatalf(`httptest.NewRecorder().Write(%q) = _, %v; want _, nil`, cnt, err)
	}
	w.Body.Reset()
	req, err := http.NewRequest(`GET`, `http://localhost/foo`, nil)
	if err != nil {
		t.Fatalf(`http.NewRequest(%q, %q, nil) = _, %v; want _, nil`, `GET`, `http://localhost/foo`, err)
	}

	// Test with encoding vary header set.
	handler := NewCompress()

	w.Header().Set(headerVary, `test`)
	handler.ServeHTTP(w, req, func(w http.ResponseWriter, r *http.Request) {

	})
	if h := w.Header().Get(headerVary); h != `test` {
		t.Errorf(`httputil.ResponseRecorder.Header().Get(%q) = %q, want %q`, headerVary, h, `test`)
	}

	// Test with empty encoding vary header.
	w.Header().Set(headerVary, ``)
	handler.ServeHTTP(w, req, func(w http.ResponseWriter, r *http.Request) {

	})
	if h := w.Header().Get(headerVary); h != headerAcceptEncoding {
		t.Errorf(`httputil.ResponseRecorder.Header().Get(%q) = %q, want %q`, headerVary, h, headerAcceptEncoding)
	}

	// Test with headers for already encoded content set.
	w.Header().Set(headerVary, ``)
	w.Header().Set(headerContentEncoding, `encoded`)
	handler.ServeHTTP(w, req, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(cnt))
	})
	if w.Body.String() != cnt {
		t.Errorf(`httptest.NewRecorder().Body.String() = %q, want %q`, w.Body.String(), cnt)
	}
	w.Body.Reset()
	w.Header().Set(headerContentEncoding, ``)

	// Test with unknown supported client encoding scheme.
	w.Header().Set(headerVary, ``)
	req.Header.Set(headerAcceptEncoding, `unknown`)
	handler.ServeHTTP(w, req, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(cnt))
	})
	if w.Body.String() != cnt {
		t.Errorf(`httptest.NewRecorder().Body.String() = %q, want %q`, w.Body.String(), cnt)
	}
	w.Body.Reset()

	// Test output content.
	for _, e := range [2][2]interface{}{
		{`gzip`, []byte{31, 139, 8, 0, 0, 9, 110, 136, 0, 255, 210, 27, 5, 163, 96, 20, 140, 130, 81, 48, 10, 70, 193, 200, 3, 128, 0, 0, 0, 255, 255, 109, 39, 42, 126, 1, 8, 0, 0}},
		{`deflate`, []byte{210, 27, 5, 163, 96, 20, 140, 130, 81, 48, 10, 70, 193, 200, 3, 128, 0, 0, 0, 255, 255}},
	} {
		req.Header.Set(headerAcceptEncoding, e[0].(string))
		for _, c := range [4][3]string{{cnt[:len(cnt)-2], `text/plain`, `0`}, {cnt[:len(cnt)-2], `application/octet-stream`, `0`}, {cnt, `application/octet-stream`, `0`}, {cnt, `text/plain`, `1`}} {
			w.Header().Set(headerVary, ``)
			handler.ServeHTTP(w, req, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set(headerContentType, c[1])
				w.Write([]byte(c[0]))
			})
			if c[2] == `0` {
				if w.Body.String() != c[0] {
					t.Errorf(`httptest.NewRecorder().Body.String() = %q, want %q`, w.Body.String(), c[0])
				}
			} else {
				b := w.Body.Bytes()
				if len(b) != len(e[1].([]byte)) {
					t.Errorf(`httptest.NewRecorder().Body.Bytes() = %v, want %v`, w.Body.Bytes(), e[1].([]byte))
					continue
				}
				for i := range b {
					if b[i] != e[1].([]byte)[i] {
						t.Errorf(`httptest.NewRecorder().Body.Bytes() = %v, want %v`, w.Body.Bytes(), e[1].([]byte))
						break
					}
				}
			}
			w.Header().Set(headerContentEncoding, ``)
			w.Body.Reset()
		}
	}
}
