// Copyright 2016 Igor "Mocheryl" Zornik. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package negronicompress

import `errors`

// ErrBadContentTypeFormat is returned when a file type in incorrect format is
// used.
var ErrBadContentTypeFormat = errors.New(`Syntax error in content type`)
