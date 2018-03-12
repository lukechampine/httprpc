package httprpc

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
)

// NewServerCodecFunc creates a new rpc.ServerCodec.
type NewServerCodecFunc func(io.ReadWriteCloser) rpc.ServerCodec

// NewServer returns an http.Handler that handles JSON-RPC requests. Only POST
// requests are allowed.
func NewServer(srv *rpc.Server, newCodec NewServerCodecFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		} else if req.Header.Get("Content-Type") != "application/json" {
			http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
			return
		} else if req.Header.Get("Accept") != "application/json" {
			http.Error(w, "Accept must be application/json", http.StatusBadRequest)
			return
		}

		var buf bytes.Buffer
		srv.ServeCodec(jsonrpc.NewServerCodec(struct {
			io.ReadCloser
			io.Writer
		}{
			ReadCloser: ioutil.NopCloser(req.Body),
			Writer:     &buf,
		}))

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", fmt.Sprint(buf.Len()))
		buf.WriteTo(w)
	})
}
