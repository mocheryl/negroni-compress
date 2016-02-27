# negroni-compress

Content-Encoding middleware for [Negroni](https://github.com/codegangsta/negroni).

## Usage

~~~ go
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
~~~

See [godoc.org](http://godoc.org/github.com/mocheryl/negroni-compress) for more information.

## License

negroni-compress is released under the 3-Clause BSD license.
See [LICENSE](https://github.com/mocheryl/negroni-compress/blob/master/LICENSE).
