package restgo

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/pinealctx/neptune/jsonx"
	"go.uber.org/zap/zapcore"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
)

const (
	defaultURLSegmentFormat = ":%s"
)

type IRequest interface {
	AddParam(param IParam) IRequest
	AddParams(params ...IParam) IRequest
	AddCookie(name, value string) IRequest
	AddHeader(name, value string) IRequest
	AddURLQuery(name, value string) IRequest
	AddURLSegment(name, value, format string) IRequest
	AddFormItem(name, value, contentType string) IRequest
	AddFileBytes(fieldName, fileName string, bytes []byte) IRequest
	AddFilePath(fieldName, filePath string) IRequest
	SetBody(contentType string, value io.Reader) IRequest
	SetJSONBody(obj interface{}) IRequest
	SetXMLBody(obj interface{}) IRequest
	WithContentType(contentType string) IRequest

	MakeURL(baseURL *url.URL) (string, error)
	GetMethod() string
	GetRequestBody() (io.Reader, error)
	WrapperHTTPRequest(req *http.Request)
}

type Request struct {
	Resource string
	Method   string

	Cookies     []*CookieParam
	Headers     []*HeaderParam
	URLQueries  []*URLQueryParam
	URLSegments []*URLSegmentParam
	FormItems   []*FormDataParam
	Files       []*FileParam
	Body        *BodyParam
	ContentType string

	Err error
}

func NewRequest(method, resource string) *Request {
	return &Request{
		Method:   method,
		Resource: resource,
	}
}

func (r *Request) AddParam(param IParam) IRequest {
	switch p := param.(type) {
	case *CookieParam:
		r.Cookies = append(r.Cookies, p)
	case *HeaderParam:
		r.Headers = append(r.Headers, p)
	case *URLQueryParam:
		r.URLQueries = append(r.URLQueries, p)
	case *URLSegmentParam:
		r.URLSegments = append(r.URLSegments, p)
	case *FormDataParam:
		r.FormItems = append(r.FormItems, p)
	case *BodyParam:
		r.Body = p
	case *FileParam:
		r.Files = append(r.Files, p)
	}
	return r
}

func (r *Request) AddParams(params ...IParam) IRequest {
	for _, p := range params {
		r.AddParam(p)
	}
	return r
}

func (r *Request) AddCookie(name, value string) IRequest {
	r.Cookies = append(r.Cookies, &CookieParam{Cookie: http.Cookie{Name: name, Value: value}})
	return r
}

func (r *Request) AddHeader(name, value string) IRequest {
	r.Headers = append(r.Headers, &HeaderParam{
		BaseParam{Name: name, Value: value},
	})
	return r
}

func (r *Request) AddURLQuery(name, value string) IRequest {
	r.URLQueries = append(r.URLQueries, &URLQueryParam{
		BaseParam{Name: name, Value: value},
	})
	return r
}

func (r *Request) AddURLSegment(name, value, format string) IRequest {
	r.URLSegments = append(r.URLSegments, &URLSegmentParam{
		BaseParam: BaseParam{
			Name:  name,
			Value: value,
		},
		Format: format,
	})
	return r
}

func (r *Request) AddFormItem(name, value, contentType string) IRequest {
	r.FormItems = append(r.FormItems, &FormDataParam{
		BaseParam: BaseParam{
			Name:  name,
			Value: value,
		},
		ContentType: contentType,
	})
	return r
}

func (r *Request) SetBody(contentType string, value io.Reader) IRequest {
	r.Body = &BodyParam{
		ContentType: contentType,
		Value:       value,
	}
	return r
}

func (r *Request) SetJSONBody(obj interface{}) IRequest {
	var buff []byte
	var err error
	switch o := obj.(type) {
	case zapcore.ObjectMarshaler:
		buff, err = ZapJSONMarshal(o)
	default:
		buff, err = jsonx.JSONFastMarshal(o)
	}
	if err != nil {
		r.Err = err
		return r
	}
	r.Body = &BodyParam{
		ContentType: "application/json; charset=utf-8",
		Value:       bytes.NewBuffer(buff),
	}
	return r
}

func (r *Request) SetXMLBody(obj interface{}) IRequest {
	var buff, err = xml.Marshal(obj)
	if err != nil {
		r.Err = err
		return r
	}
	r.Body = &BodyParam{
		ContentType: "application/xml; charset=utf-8",
		Value:       bytes.NewBuffer(buff),
	}
	return r
}

func (r *Request) AddFileBytes(fieldName, fileName string, bytes []byte) IRequest {
	r.Files = append(r.Files, &FileParam{
		Name:           fieldName,
		FileName:       fileName,
		ContentType:    http.DetectContentType(bytes),
		ContentLength:  int64(len(bytes)),
		FileWriterFunc: BytesWriter(bytes),
	})
	return r
}

func (r *Request) AddFilePath(fieldName, filePath string) IRequest {
	var contentType, size, err = DetectContentTypeAndSize(filePath)
	if err != nil {
		r.Err = err
		return r
	}
	r.Files = append(r.Files, &FileParam{
		Name:           fieldName,
		FileName:       path.Base(filePath),
		ContentType:    contentType,
		ContentLength:  size,
		FileWriterFunc: FileWriter(filePath),
	})
	return r
}

func (r *Request) WithContentType(contentType string) IRequest {
	r.ContentType = contentType
	return r
}

func (r *Request) MakeURL(baseURL *url.URL) (string, error) {
	var err error
	if baseURL == nil {
		baseURL, err = url.ParseRequestURI(r.Resource)
		if err != nil {
			return "", err
		}
	} else {
		baseURL.Path = path.Join(baseURL.Path, r.Resource)
	}
	var q = baseURL.Query()
	for _, query := range r.URLQueries {
		q.Add(query.Name, query.Value)
	}
	baseURL.RawQuery = q.Encode()
	var outURL = baseURL.String()
	var segSize = len(r.URLSegments)
	if segSize != 0 {
		var replaces = make([]string, segSize*2)
		for i, seg := range r.URLSegments {
			replaces[i*2] = fmt.Sprintf(defaultURLSegmentFormat, seg.Name)
			replaces[i*2+1] = seg.Value
		}
		var replacer = strings.NewReplacer(replaces...)
		outURL = replacer.Replace(outURL)
	}
	return outURL, nil
}

func (r *Request) GetMethod() string {
	if r.Method == "" {
		return "GET"
	}
	return r.Method
}

func (r *Request) GetRequestBody() (io.Reader, error) {
	if r.Body != nil {
		if r.Body.ContentType != "" {
			r.ContentType = r.Body.ContentType
		}
		return r.Body.Value, nil
	}
	if len(r.Files) == 0 {
		return r.makeFormDataBody(), nil
	}
	return r.makeMultipartBody()
}

func (r *Request) WrapperHTTPRequest(req *http.Request) {
	for _, p := range r.Headers {
		req.Header.Add(p.Name, p.Value)
	}
	for _, p := range r.Cookies {
		req.AddCookie(&p.Cookie)
	}
	var ct = req.Header.Get(headerContentType)
	if ct == "" && r.ContentType != "" {
		req.Header.Set(headerContentType, r.ContentType)
	}
}

func (r *Request) makeFormDataBody() io.Reader {
	var values = url.Values{}
	for _, c := range r.FormItems {
		values.Add(c.Name, c.Value)
	}
	r.ContentType = "application/x-www-form-urlencoded"
	return strings.NewReader(values.Encode())
}

func (r *Request) makeMultipartBody() (io.Reader, error) {
	var body = new(bytes.Buffer)
	var writer = multipart.NewWriter(body)
	var fileWriter io.Writer
	var err error
	for _, f := range r.Files {
		fileWriter, err = writer.CreateFormFile(f.Name, f.FileName)
		if err != nil {
			return nil, err
		}
		err = f.FileWriterFunc(fileWriter)
		if err != nil {
			return nil, err
		}
	}
	for _, f := range r.FormItems {
		err = writer.WriteField(f.Name, f.Value)
		if err != nil {
			return nil, err
		}
	}
	r.ContentType = writer.FormDataContentType()
	return body, nil
}
