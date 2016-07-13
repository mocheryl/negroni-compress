// Copyright 2016 Igor "Mocheryl" Zornik. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package negronicompress

import (
	"regexp"
	"testing"
)

func TestCompileFileTypes(t *testing.T) {
	if r, err := compileFileTypes(compressiableFileTypes); err != nil || r.String() != `^(text/.+|application/x-javascript|application/xhtml+xml)$` {
		t.Errorf(`negronimodified.compileFile(%v) = %q, %v; want %q, nil`, compressiableFileTypes, r.String(), err, `^(text/.+|application/x-javascript|application/xhtml+xml)$`)
	}

	if r, err := compileFileTypes([]string{}); err != nil || r.String() != `.*` {
		t.Errorf(`negronimodified.compileFile(%v) = %q, %v; want %q, nil`, []string{}, r.String(), err, `.*`)
	}
}

func TestAppendFileType(t *testing.T) {
	cLen, cOrig := len(compressiableFileTypes), compressiableFileTypes
	l, err := appendFileType(compressiableFileTypes, `*/*`)
	if len(l) != 0 || err != nil {
		t.Errorf(`negronimodified.appendFileType(%v, %q) = %v, %v; want %v, nil`, compressiableFileTypes, `*/*`, l, err, []string{})
	}

	l, err = appendFileType(compressiableFileTypes, `xyz`)
	if err != ErrBadContentTypeFormat {
		t.Errorf(`negronimodified.appendFileType(%v, %q) = %v, %v; want %v, %v`, compressiableFileTypes, `xyz`, l, err, compressiableFileTypes, ErrBadContentTypeFormat)
	}
	if len(l) != cLen {
		t.Fatalf(`len(negronimodified.appendFileType(%v, %q)) = %d, want %d`, compressiableFileTypes, `xyz`, len(l), cLen)
	}
	for i := range compressiableFileTypes {
		if compressiableFileTypes[i] != l[i] {
			t.Errorf(`negronimodified.appendFileType(%v, %q)[%d] = %q, want %q`, compressiableFileTypes, `xyz`, i, l[i], compressiableFileTypes[i])
		}
	}

	l, err = appendFileType(compressiableFileTypes, `text/*`)
	if len(l) != cLen || err != nil {
		t.Fatalf(`negronimodified.appendFileType(%v, %q) = %v, %v; want %v, nil`, compressiableFileTypes, `text/*`, l, err, compressiableFileTypes)
	}
	for i := range compressiableFileTypes {
		if compressiableFileTypes[i] != l[i] {
			t.Errorf(`negronimodified.appendFileType(%v, %q)[%d] = %q, want %q`, compressiableFileTypes, `text/*`, i, l[i], compressiableFileTypes[i])
		}
	}

	l, err = appendFileType(compressiableFileTypes, `application/octet-stream`)
	if len(l) != (cLen+1) || err != nil {
		t.Fatalf(`negronimodified.appendFileType(%v, %q) = %v, %v; want %v, nil`, compressiableFileTypes, `application/octet-stream`, l, err, compressiableFileTypes)
	}
	for i := range compressiableFileTypes {
		if compressiableFileTypes[i] != l[i] {
			t.Errorf(`negronimodified.appendFileType(%v, %q)[%d] = %q, want %q`, compressiableFileTypes, `application/octet-stream`, i, l[i], compressiableFileTypes[i])
		}
	}
	if l[cLen] != `application/octet-stream` {
		t.Errorf(`negronimodified.appendFileType(%v, %q)[%d] = %q, want %q`, compressiableFileTypes, `application/octet-stream`, cLen, l[cLen], `application/octet-stream`)
	}

	compressiableFileTypes = cOrig
	compressContentTypeRegEx, _ = compileFileTypes(compressiableFileTypes)
}

func TestAddContentType(t *testing.T) {
	cLen, cOrig, cExOrig := len(compressiableFileTypes), compressiableFileTypes, *contentTypeRegEx
	if err := AddContentType(`xyz`); err == nil {
		t.Errorf(`negronimodified.AddContentType(%q) = nil, want err`, `xyz`)
	}

	if err := AddContentType(`application/octet-stream`); err != nil {
		t.Fatalf(`negronimodified.AddContentType(%q) = %v, want nil`, `application/octet-stream`, err)
	}
	if l := len(compressiableFileTypes); l != cLen+1 {
		t.Fatalf(`len(negronimodified.compressiableFileTypes) = %d, want %d`, l, cLen+1)
	}
	for i := range cOrig {
		if cOrig[i] != compressiableFileTypes[i] {
			t.Errorf(`negronimodified.compressiableFileTypes[%d] = %q, want %q`, i, compressiableFileTypes[i], cOrig[i])
		}
	}
	if compressiableFileTypes[cLen] != `application/octet-stream` {
		t.Errorf(`negronimodified.compressiableFileTypes[%d] = %q, want %q`, cLen, compressiableFileTypes[cLen], `application/octet-stream`)
	}
	if r := compressContentTypeRegEx.String(); r != `^(text/.+|application/x-javascript|application/xhtml+xml|application/octet-stream)$` {
		t.Errorf(`negronimodified.compressContentTypeRegEx.String() = %q, want %q`, r, `^(text/.+|application/x-javascript|application/xhtml+xml|application/octet-stream)$`)
	}

	contentTypeRegEx = regexp.MustCompile(`.*`)
	if err := AddContentType(`\x`); err == nil {
		t.Errorf(`negronimodified.AddContentType(%q) = nil, want err`, `\x`)
	}

	compressiableFileTypes = cOrig
	compressContentTypeRegEx, _ = compileFileTypes(compressiableFileTypes)
	contentTypeRegEx = &cExOrig
}
