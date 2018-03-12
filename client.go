package httprpc

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/rpc"
	"sync/atomic"
)

// NewClientCodecFunc creates a new rpc.ClientCodec.
type NewClientCodecFunc func(io.ReadWriteCloser) rpc.ClientCodec

// A Client can make RPC requests over HTTP.
type Client struct {
	seq      uint64
	newCodec NewClientCodecFunc
	url      string
	tweak    func(*http.Request)
}

func (c *Client) do(call *rpc.Call) error {
	// write the RPC request into a buffer
	var buf bytes.Buffer
	rwc := &struct {
		io.ReadCloser
		io.Writer
	}{Writer: &buf}
	codec := c.newCodec(rwc)

	seq := atomic.AddUint64(&c.seq, 1)
	err := codec.WriteRequest(&rpc.Request{
		ServiceMethod: call.ServiceMethod,
		Seq:           seq,
	}, call.Args)
	if err != nil {
		return err
	}

	// create the HTTP request
	req, err := http.NewRequest(http.MethodPost, c.url, &buf)
	if err != nil {
		return err
	}
	if c.tweak != nil {
		c.tweak(req)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	} else if resp.StatusCode < 200 || 299 < resp.StatusCode {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("server returned non-200 status code %v: %s", resp.StatusCode, bytes.TrimSpace(body))
	}

	rwc.ReadCloser = ioutil.NopCloser(resp.Body)
	var rpcResp rpc.Response
	err = codec.ReadResponseHeader(&rpcResp)
	if err != nil {
		return err
	} else if rpcResp.Seq != seq {
		return errors.New("mismatched sequence number")
	} else if rpcResp.Error != "" {
		return rpc.ServerError(rpcResp.Error)
	}
	return codec.ReadResponseBody(call.Reply)
}

// Go invokes the function asynchronously. It returns the rpc.Call structure
// representing the invocation. The done channel will signal when the call is
// complete by returning the same rpc.Call object. If done is nil, Go will
// allocate a new channel.
func (c *Client) Go(method string, args interface{}, reply interface{}, done chan *rpc.Call) *rpc.Call {
	if done == nil {
		done = make(chan *rpc.Call, 1)
	}
	call := &rpc.Call{
		ServiceMethod: method,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	go func() {
		call.Error = c.do(call)
		call.Done <- call
	}()
	return call
}

// Call invokes the named function, waits for it to complete, and returns its
// error status.
func (c *Client) Call(method string, args interface{}, reply interface{}) error {
	call := <-c.Go(method, args, reply, make(chan *rpc.Call, 1)).Done
	return call.Error
}

// NewClient returns a Client that directs its requests at the specified url.
// If a tweak function is supplied, it is called on each http.Request
// immediately before the request is performed.
func NewClient(url string, codecFn NewClientCodecFunc, tweak func(*http.Request)) *Client {
	return &Client{
		url:      url,
		newCodec: codecFn,
		tweak:    tweak,
	}
}
