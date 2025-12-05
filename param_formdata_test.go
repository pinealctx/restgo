package restgo

import (
	"bytes"
	"strings"
	"testing"
)

// TestNewFormBody 测试 NewFormBody 函数
func TestNewFormBody(t *testing.T) {
	fields := map[string]string{
		"name":  "John Doe",
		"email": "john@example.com",
	}
	files := map[string][]byte{
		"avatar": []byte("fake image data"),
	}

	body, err := NewFormBody(fields, files)
	if err != nil {
		t.Fatalf("NewFormBody failed: %v", err)
	}

	if body == nil {
		t.Fatal("body should not be nil")
	}

	if body.ContentType == "" {
		t.Fatal("ContentType should not be empty")
	}

	if !strings.Contains(body.ContentType, "multipart/form-data") {
		t.Fatalf("ContentType should contain 'multipart/form-data', got: %s", body.ContentType)
	}

	// 读取 body 内容
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(body.Value)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}

	bodyStr := buf.String()
	if !strings.Contains(bodyStr, "name") ||
		!strings.Contains(bodyStr, "John Doe") {
		t.Fatal("body should contain form fields")
	}

	if !strings.Contains(bodyStr, "avatar") {
		t.Fatal("body should contain file field")
	}

	t.Log("TestNewFormBody passed")
}

// TestFormDataBuilder 测试 FormDataBuilder
func TestFormDataBuilder(t *testing.T) {
	builder := NewFormDataBuilder().
		AddField("username", "alice").
		AddField("password", "secret123").
		AddFile("profile", []byte("profile image data"))

	body, err := NewFormBodyBuilder(builder)
	if err != nil {
		t.Fatalf("NewFormBodyBuilder failed: %v", err)
	}

	if body == nil {
		t.Fatal("body should not be nil")
	}

	if !strings.Contains(body.ContentType, "multipart/form-data") {
		t.Fatalf("ContentType should contain 'multipart/form-data'")
	}

	t.Log("TestFormDataBuilder passed")
}

// TestRequestSetFormBody 测试 Request.SetFormBody 方法
func TestRequestSetFormBody(t *testing.T) {
	req := NewRequest("POST", "/upload")

	fields := map[string]string{
		"title": "My Document",
	}
	files := map[string][]byte{
		"document": []byte("PDF content here"),
	}

	req.SetFormBody(fields, files)

	if req.Body == nil {
		t.Fatal("Body should not be nil after SetFormBody")
	}

	if !strings.Contains(req.Body.ContentType, "multipart/form-data") {
		t.Fatalf("Body ContentType should be multipart/form-data")
	}

	t.Log("TestRequestSetFormBody passed")
}
