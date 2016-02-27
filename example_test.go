// Copyright 2016 Igor "Mocheryl" Zornik. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package negronicompress

import (
	`fmt`
	`net/http`

	`github.com/codegangsta/negroni`
)

// NewCompress basic usage.
func ExampleNewCompress() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set(`Content-Type`, `text/plain`)
		s := `Large enough compressiable content will be encoded based on client encoding support.`
		for i := 0; i <= 2048; i++ {
			s += `.`
		}
		fmt.Fprintf(w, s)
	})

	n := negroni.Classic()
	n.Use(NewCompress())
	n.UseHandler(mux)
	n.Run(`:3000`)
}
