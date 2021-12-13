package restgo

import (
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"net/http"
	"net/url"
	"os"
)

// The algorithm uses at most sniffLen bytes to make its decision.
const sniffLen = 512

// ZapJSONMarshal Zap JSON 序列化
func ZapJSONMarshal(obj zapcore.ObjectMarshaler) ([]byte, error) {
	var entry = zapcore.Entry{}
	var fields []zapcore.Field
	var enc = zapcore.NewJSONEncoder(zapcore.EncoderConfig{})
	var err = obj.MarshalLogObject(enc)
	if err != nil {
		return nil, err
	}
	var buf *buffer.Buffer
	buf, err = enc.EncodeEntry(entry, fields)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DetectContentTypeAndSize 发现文件的ContentType和大小
func DetectContentTypeAndSize(filePath string) (string, int64, error) {
	var fi, err = os.Stat(filePath)
	if err != nil {
		return "", 0, err
	}
	var f *os.File
	f, err = os.Open(filePath)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	var buff = make([]byte, sniffLen)
	_, err = f.Read(buff)
	if err != nil {
		return "", 0, err
	}
	return http.DetectContentType(buff), fi.Size(), nil
}

func CloneURL(o *url.URL) *url.URL {
	if o == nil {
		return nil
	}
	return &url.URL{
		Scheme:      o.Scheme,
		Opaque:      o.Opaque,
		User:        o.User,
		Host:        o.Host,
		Path:        o.Path,
		RawPath:     o.RawPath,
		ForceQuery:  o.ForceQuery,
		RawQuery:    o.RawQuery,
		Fragment:    o.Fragment,
		RawFragment: o.RawFragment,
	}
}
