package restgo

import (
	"io"
	"net/http"
	"os"
	"path"
)

type IParam interface {
	ParamName() string
}

// BaseParam 参数基类
type BaseParam struct {
	Name  string
	Value string
}

func (p BaseParam) ParamName() string {
	return p.Name
}

// CookieParam 将参数通过cookie携带
type CookieParam struct {
	http.Cookie
}

func NewCookieParam(cookie *http.Cookie) *CookieParam {
	return &CookieParam{Cookie: *cookie}
}

func (p CookieParam) ParamName() string {
	return p.Name
}

// HeaderParam 将参数通过HTTP Header携带
type HeaderParam struct {
	BaseParam
}

func NewHeaderParam(name, value string) *HeaderParam {
	return &HeaderParam{BaseParam{
		Name:  name,
		Value: value,
	}}
}

// URLQueryParam 将参数通过URL Query携带
type URLQueryParam struct {
	BaseParam
}

func NewURLQueryParam(name, value string) *URLQueryParam {
	return &URLQueryParam{BaseParam{
		Name:  name,
		Value: value,
	}}
}

// URLSegmentParam 将参数通过URL segment携带
// eg: path = "tag/:Resource"，name = "Resource"
type URLSegmentParam struct {
	BaseParam
	Format string
}

func NewURLSegmentParam(name, value, format string) *URLSegmentParam {
	return &URLSegmentParam{BaseParam: BaseParam{
		Name:  name,
		Value: value,
	}, Format: format}
}

// FormDataParam 将参数通过Form表单（multipart/form-data或application/x-www-form-urlencoded）携带
type FormDataParam struct {
	BaseParam
	ContentType string
}

func NewFormDataParam(name, value, contentType string) *FormDataParam {
	return &FormDataParam{BaseParam: BaseParam{
		Name:  name,
		Value: value,
	}, ContentType: contentType}
}

// BodyParam 将参数作为HTTP Body携带，具体序列化方式通过参数内容类型而定
// 同一个request有且仅有一个BodyParam
type BodyParam struct {
	ContentType string
	Value       io.Reader
}

func NewBodyParam(contentType string, value io.Reader) *BodyParam {
	return &BodyParam{Value: value, ContentType: contentType}
}

func (p BodyParam) ParamName() string {
	return p.ContentType
}

// FileParam 将文件作为参数携带（multipart/form-data）
type FileParam struct {
	Name           string
	FileName       string
	ContentType    string
	ContentLength  int64
	FileWriterFunc WriterFunc
}

func NewBytesFileParam(fieldName, fileName string, bytes []byte) *FileParam {
	return &FileParam{
		Name:           fieldName,
		FileName:       fileName,
		ContentType:    http.DetectContentType(bytes),
		ContentLength:  int64(len(bytes)),
		FileWriterFunc: BytesWriter(bytes),
	}
}

func NewPathFileParam(fieldName, filePath string) (*FileParam, error) {
	var contentType, size, err = DetectContentTypeAndSize(filePath)
	if err != nil {
		return nil, err
	}
	return &FileParam{
		Name:           fieldName,
		FileName:       path.Base(filePath),
		ContentType:    contentType,
		ContentLength:  size,
		FileWriterFunc: FileWriter(filePath),
	}, nil
}

func (p FileParam) ParamName() string {
	return p.Name
}

type WriterFunc func(w io.Writer) error

func BytesWriter(buff []byte) WriterFunc {
	return func(w io.Writer) error {
		var _, err = w.Write(buff)
		return err
	}
}

func FileWriter(filePath string) WriterFunc {
	return func(w io.Writer) error {
		var file, err = os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()
		var buf = make([]byte, 1024)
		for {
			var n int
			n, err = file.Read(buf)
			if err != nil && err != io.EOF {
				return err
			}
			if n == 0 {
				break
			}
			_, err = w.Write(buf[:n])
			if err != nil {
				return err
			}
		}
		return nil
	}
}
