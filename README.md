httprpc
=======

[![GoDoc](https://godoc.org/github.com/lukechampine/httprpc?status.svg)](https://godoc.org/github.com/lukechampine/httprpc)
[![Go Report Card](http://goreportcard.com/badge/github.com/lukechampine/httprpc)](https://goreportcard.com/report/github.com/lukechampine/httprpc)

```
go get github.com/lukechampine/httprpc
```

`httprpc` provides HTTP wrappers for the `net/rpc` package.

While the `net/rpc` package does provide `DialHTTP` and `ServeHTTP` functions,
these only support Go's `gob` encoding format, and do so by hijacking the
underlying HTTP connection. `httprpc` supports any text-based format
satisfying the `rpc.ServerCodec`/`rpc.ClientCodec` interfaces (most notably
JSON-RPC) by storing request/response payloads in the request/response body.

`httprpc` has only been tested with JSON-RPC.

Usage
-----

```go
type Args struct {
	A, B int
}

type Arith int

func (t *Arith) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

func basicAuthHandler(next http.Handler, password string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		_, pass, ok := req.BasicAuth()
		if !ok || pass != password {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"MyRealm\"")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, req)
	}
}

func main() {
	// create JSON-RPC HTTP server, wrapped with Basic Auth middleware
	rpc.Register(new(Arith))
	srv := httprpc.NewServer(rpc.DefaultServer, jsonrpc.NewServerCodec)
	srv = basicAuthHandler(srv, "foo")
	go http.ListenAndServe(":5555", srv)

	// create JSON-RPC HTTP client, wrapped with Basic Auth "tweak"
	c := httprpc.NewClient("http://127.0.0.1:5555", jsonrpc.NewClientCodec, func(req *http.Request) {
		req.SetBasicAuth("", "foo")
	})

	// invoke RPC
	args := &Args{7, 8}
	var reply int
	err := c.Call("Arith.Multiply", args, &reply)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Arith: %d*%d=%d\n", args.A, args.B, reply)
}
```
