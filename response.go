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
	GetResponse() *http.Response
	StatusCode() int
	Data() ([]byte, error)
	Pipe(writer io.Writer) error
	JSONUnmarshal(i interface{}) error
	XMLUnmarshal(i interface{}) error
	SaveFile(fileName string) error
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
