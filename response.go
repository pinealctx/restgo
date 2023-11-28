package restgo

import (
	"encoding/xml"
	"github.com/pinealctx/neptune/jsonx"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

type IResponse interface {
	// GetResponse get http response
	GetResponse() *http.Response
	// StatusCode get http status code
	StatusCode() int
	// Data get response data
	// it will automatically close response body
	Data() ([]byte, error)
	// Pipe : pipe response data to writer
	// it will automatically close response body
	Pipe(writer io.Writer) error
	// JSONUnmarshal unmarshal response data to json
	// it will automatically close response body
	JSONUnmarshal(i interface{}) error
	// XMLUnmarshal unmarshal response data to xml
	// it will automatically close response body
	XMLUnmarshal(i interface{}) error
	// SaveFile save response data to file
	// it will automatically close response body
	SaveFile(fileName string) error
	// ExplicitCloseBody close response body
	// If you get a response but never access it's body,
	// you should call this method to close response body.
	// Otherwise, it will cause resource leak.
	// Actually, you can use GetResponse().Body.Close() to close response body,
	// but this method is more convenient and remind you to close response body.
	ExplicitCloseBody() error
}

type Response struct {
	rsp  *http.Response
	data []byte
}

func NewResponse(rsp *http.Response) IResponse {
	return &Response{rsp: rsp}
}

func (r *Response) GetResponse() *http.Response {
	return r.rsp
}

func (r *Response) StatusCode() int {
	return r.rsp.StatusCode
}

func (r *Response) Data() ([]byte, error) {
	if r.data != nil {
		return r.data, nil
	}
	defer r.rsp.Body.Close()
	var err error
	r.data, err = ioutil.ReadAll(r.rsp.Body)
	return r.data, err
}

func (r *Response) Pipe(writer io.Writer) error {
	defer r.rsp.Body.Close()
	var _, err = io.Copy(writer, r.rsp.Body)
	return err
}

func (r *Response) JSONUnmarshal(i interface{}) error {
	var data, err = r.Data()
	if err != nil {
		return err
	}
	return jsonx.JSONFastUnmarshal(data, i)
}

func (r *Response) XMLUnmarshal(i interface{}) error {
	var data, err = r.Data()
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, i)
}

func (r *Response) SaveFile(fileName string) error {
	var writer, err = os.Create(fileName)
	if err != nil {
		return err
	}
	defer writer.Close()
	return r.Pipe(writer)
}

func (r *Response) ExplicitCloseBody() error {
	return r.rsp.Body.Close()
}
