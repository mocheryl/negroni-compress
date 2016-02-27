// Copyright 2016 Igor "Mocheryl" Zornik. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package negronicompress implements a Negroni middleware handler for various
HTTP content compression methods.

Basics

A lot of content that gets sent out form the HTTP server is usually in text
format. This kind of output content can be large in size and has very good
compression potential. In HTTP we can take advantage of this potential with
the use of various compression encoding schemes. Most notably with "deflate"
and "gzip" (more others are still in the work) methods of content encoding.
If the user agent supports any such mechanisms, it specifies its support with
the use of the "Accept-Encoding" HTTP header value. The middleware can pick this
up and respond accordingly by encoding any output content in one of the user
agents supported compression formats. By doing this, any content sent back to
client is greatly reduced in size and thus saving on bandwidth and
consequentially loading time.

Usage

	package main

	import (
		`fmt`
		`net/http`

		`github.com/codegangsta/negroni`
		`github.com/mocheryl/negroni-compress`
	)

	func main() {
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
		n.Use(negronimodified.NewCompress())
		n.UseHandler(mux)
		n.Run(`:3000`)
	}

The above code initializes the middleware with default settings. These include
compressing only content types that are the most widely used in rendering a web
page at a level that is a good compromise between speed of
compression/decompression and compression ratio.

You can define your own level of compression by initializing the middleware like
this:

	NewCompressWithCompressionLevel(9)

Where higher value means better compression but also more processing time and
power while lower number outputs encoded content faster but yields worse
compression ratio. Keep in mind that the value cannot go below 1 nor above 9.

You can specify additional content types to check for compression.

	m.AddContentType(`application/pdf`, `image/*`)

This will compress all .pdf files and images when the middeleware comes across
them in the "Content-Type" HTTP header usually set by the other backend
services.

Tips

If you have multiple instances of this middleware and all share the same custom
list of content types allowed to compress, you can alter the global list with a
helper function that works in the same way as the middlerware method.

	AddContentType(`application/pdf`, `image/*`)

Now all new middleware instances instantiated after this function call will have
this new content list set as the default list. The list belonging to the
middeware itself can be then further altered with the method call without
affecting any other lists.

To clear the list, you can call the same function by specifying all types.

	AddContentType(`*\/*`)

Empty list means match any type and thus compress it. After the function call
you can then freely create your own custom list of types.

*/
package negronicompress
