package restgo

import "net/http"

type IResponse interface {
	Close()
}

type Response struct {
	rsp *http.Response
}

func NewResponse(rsp *http.Response) IResponse {
	return &Response{rsp: rsp}
}

func (r *Response) Close() {
	r.rsp.Body.Close()
}
