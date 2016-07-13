// Copyright 2016 Igor "Mocheryl" Zornik. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package negronicompress

import (
	"regexp"
	"strings"
)

// compressiableFileTypes is a list of "Content-Type" file types that should be
// compressed. By default it uses some common file types passed as output on
// modern HTML webpages.
var compressiableFileTypes = []string{`text/.+`, `application/x-javascript`, `application/xhtml+xml`}

// contentTypeRegEx is a regular expression for the allowed value to be passed
// as compressionable file type.
var contentTypeRegEx = regexp.MustCompile(`^(\w+|\*)/([-+.\w]|\*)+$`)

// compressContentTypeRegEx is list of compresionable file types in a regular
// expression format.
// TODO: Check for error.
var compressContentTypeRegEx, _ = compileFileTypes(compressiableFileTypes)

// compileFileTypes turns a slice of file types into a regular expression fit
// to be used for checking "Content-Type" headers.
func compileFileTypes(fileTypes []string) (*regexp.Regexp, error) {
	if len(fileTypes) == 0 {
		return regexp.Compile(`.*`)
	}

	return regexp.Compile(`^(` + strings.Join(fileTypes, `|`) + `)$`)
}

// appendFileType adds c to a list of file types if c is not already present in
// the list. If c is "*/*", it will return an empty list.
func appendFileType(fileTypes []string, c string) ([]string, error) {
	if c == `*/*` {
		return []string{}, nil
	}

	if !contentTypeRegEx.MatchString(c) {
		return fileTypes, ErrBadContentTypeFormat
	}

	c = strings.Replace(c, `*`, `.+`, -1)
	for _, f := range fileTypes {
		if f == c {
			return fileTypes, nil
		}
	}

	return append(fileTypes, c), nil
}

// AddContentType adds a new file type to the global list of file types that can
// be compressed. c should match the form used of a value used in "Content-Type"
// HTTP header. If c is "*/*", it will reset the list to empty value making it
// match all types, including no type.
func AddContentType(c ...string) (err error) {
	// Setup new content type list.
	cList := compressiableFileTypes
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

	compressiableFileTypes = cList
	compressContentTypeRegEx = r

	return
}
