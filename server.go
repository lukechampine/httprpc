package httprpc

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/rpc"
)

// NewServerCodecFunc creates a new rpc.ServerCodec.
type NewServerCodecFunc func(io.ReadWriteCloser) rpc.ServerCodec

// NewServer returns an http.Handler that handles RPC requests.
func NewServer(srv *rpc.Server, newCodec NewServerCodecFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var buf bytes.Buffer
		srv.ServeCodec(newCodec(struct {
			io.ReadCloser
			io.Writer
		}{
			ReadCloser: ioutil.NopCloser(req.Body),
			Writer:     &buf,
		}))
		w.Header().Set("Content-Length", fmt.Sprint(buf.Len()))
		buf.WriteTo(w)
	})
}
