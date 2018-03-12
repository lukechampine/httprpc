package jsonrpc

import (
	"mime"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/lukechampine/httprpc"
)

// NewServer returns an http.Handler that handles JSON-RPC requests.
func NewServer(srv *rpc.Server) http.Handler {
	rh := httprpc.NewServer(srv, jsonrpc.NewServerCodec)
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		mediaType, _, _ := mime.ParseMediaType(req.Header.Get("Content-Type"))
		if mediaType != "application/json" || req.Header.Get("Accept") != "application/json" {
			http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		rh.ServeHTTP(w, req)
	})
}

// NewClient returns an httprpc.Client that directs JSON-RPC requests at the
// specified url. If a tweak function is supplied, it is called on each
// http.Request immediately before the request is performed.
func NewClient(url string, tweak func(*http.Request)) *httprpc.Client {
	return httprpc.NewClient(url, jsonrpc.NewClientCodec, func(req *http.Request) {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		tweak(req)
	})
}
