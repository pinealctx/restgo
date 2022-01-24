package restgo

import (
	"bytes"
	"encoding/xml"
	"github.com/fatih/structtag"
	"github.com/pinealctx/neptune/jsonx"
	"github.com/pinealctx/neptune/tex"
	"go.uber.org/zap/zapcore"
	"io"
	"net/http"
	"os"
	"path"
	"reflect"
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
}

func NewFormDataParam(name, value string) *FormDataParam {
	return &FormDataParam{BaseParam{Name: name, Value: value}}
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

func NewJSONBody(obj interface{}) (*BodyParam, error) {
	var buff []byte
	var err error
	switch o := obj.(type) {
	case zapcore.ObjectMarshaler:
		buff, err = ZapJSONMarshal(o)
	default:
		buff, err = jsonx.JSONFastMarshal(o)
	}
	if err != nil {
		return nil, err
	}
	return &BodyParam{
		ContentType: "application/json; charset=utf-8",
		Value:       bytes.NewBuffer(buff),
	}, nil
}

func NewXMLBody(obj interface{}) (*BodyParam, error) {
	var buff, err = xml.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return &BodyParam{
		ContentType: "application/xml; charset=utf-8",
		Value:       bytes.NewBuffer(buff),
	}, nil
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

func ObjectParams(obj interface{}) []IParam {
	var objV = reflect.ValueOf(obj)
	if objV.Kind() != reflect.Ptr {
		return nil
	}
	objV = objV.Elem()
	if objV.Kind() != reflect.Struct {
		return nil
	}
	var objT = objV.Type()
	var size = objT.NumField()
	var params = make([]IParam, 0, size)
	for i := 0; i < size; i++ {
		f := objT.Field(i)
		if f.PkgPath != "" && !f.Anonymous {
			continue
		}
		tags, err := structtag.Parse(string(f.Tag))
		if err != nil {
			continue
		}
		var p = tags2Param(tags, objV.Field(i))
		if p != nil {
			params = append(params, p)
		}
	}
	return params
}

const (
	tagNameQuery      = "query"
	tagNamePath       = "path"
	tagNameHeader     = "header"
	tagNameCookie     = "cookie"
	tagOptionRequired = "required"
)

func tags2Param(tags *structtag.Tags, v reflect.Value) IParam {
	var strValue = tex.ToString(v.Interface())
	var err error
	var tag *structtag.Tag
	if tag, err = tags.Get(tagNameQuery); err == nil {
		if tag.HasOption(tagOptionRequired) || !v.IsZero() {
			return NewURLQueryParam(tag.Name, strValue)
		}
	}
	if tag, err = tags.Get(tagNamePath); err == nil {
		if tag.HasOption(tagOptionRequired) || !v.IsZero() {
			return NewURLSegmentParam(tag.Name, strValue, "")
		}
	}
	if tag, err = tags.Get(tagNameHeader); err == nil {
		if tag.HasOption(tagOptionRequired) || !v.IsZero() {
			return NewHeaderParam(tag.Name, strValue)
		}
	}
	if tag, err = tags.Get(tagNameCookie); err == nil {
		if tag.HasOption(tagOptionRequired) || !v.IsZero() {
			return NewCookieParam(&http.Cookie{
				Name:  tag.Name,
				Value: strValue,
			})
		}
	}
	return nil
}
