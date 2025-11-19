# Form-Data 支持使用指南

本文档介绍如何使用新增的 form-data 功能来构建和发送 multipart/form-data 请求。

## 功能概述

`restgo` 现在支持以下与 form-data 相关的功能：

1. **`NewFormBody`** - 直接创建 multipart/form-data 格式的请求体
2. **`FormDataBuilder`** - 流式 API 来构建 form-data 内容
3. **`Request.SetFormBody`** - 在请求对象上设置 form-data 体
4. **`NewFormBodyBuilder`** - 使用 builder 创建 form-data 请求体

## 使用示例

### 方法 1：直接使用 NewFormBody

```go
package main

import (
	"context"
	"github.com/pinealctx/restgo"
)

func main() {
	client := restgo.New()
	
	// 准备表单字段和文件
	fields := map[string]string{
		"name":  "John Doe",
		"email": "john@example.com",
	}
	
	files := map[string][]byte{
		"avatar":  []byte("...image data..."),
		"resume":  []byte("...pdf data..."),
	}
	
	// 创建请求
	req := restgo.NewRequest("POST", "/upload")
	req.SetFormBody(fields, files)
	
	// 发送请求
	resp, err := client.Do(context.Background(), req)
	if err != nil {
		panic(err)
	}
	
	data, _ := resp.Data()
	println(string(data))
}
```

### 方法 2：使用 FormDataBuilder

```go
package main

import (
	"context"
	"github.com/pinealctx/restgo"
)

func main() {
	client := restgo.New()
	
	// 使用 builder 构建 form-data
	builder := restgo.NewFormDataBuilder().
		AddField("username", "alice").
		AddField("password", "secret123").
		AddFile("profile_pic", []byte("...image data..."))
	
	// 创建请求
	req := restgo.NewRequest("POST", "/user/profile")
	body, err := restgo.NewFormBodyBuilder(builder)
	if err != nil {
		panic(err)
	}
	req.AddParam(body)
	
	// 发送请求
	resp, err := client.Do(context.Background(), req)
	if err != nil {
		panic(err)
	}
	
	data, _ := resp.Data()
	println(string(data))
}
```

### 方法 3：与其他参数结合

```go
package main

import (
	"context"
	"github.com/pinealctx/restgo"
)

func main() {
	client := restgo.New()
	
	fields := map[string]string{
		"title": "My Document",
		"category": "work",
	}
	
	files := map[string][]byte{
		"document": []byte("...document data..."),
	}
	
	req := restgo.NewRequest("POST", "/documents").
		AddHeader("Authorization", "Bearer token123").
		AddURLQuery("version", "v1").
		SetFormBody(fields, files)
	
	resp, err := client.Do(context.Background(), req)
	if err != nil {
		panic(err)
	}
	
	data, _ := resp.Data()
	println(string(data))
}
```

## API 参考

### NewFormBody

```go
func NewFormBody(fields map[string]string, files map[string][]byte) (*BodyParam, error)
```

创建一个 multipart/form-data 格式的 BodyParam。

**参数：**
- `fields`: 表单字段键值对映射
- `files`: 文件字段映射（字段名 -> 文件内容）

**返回值：**
- `*BodyParam`: 构建好的请求体
- `error`: 构建过程中的错误

### FormDataBuilder

```go
type FormDataBuilder struct {
	Fields map[string]string
	Files  map[string][]byte
}

func NewFormDataBuilder() *FormDataBuilder
func (fdb *FormDataBuilder) AddField(name, value string) *FormDataBuilder
func (fdb *FormDataBuilder) AddFile(fieldName string, fileContent []byte) *FormDataBuilder
```

用于流式构建 form-data 内容的 builder 类。

**示例：**
```go
builder := NewFormDataBuilder().
	AddField("name", "value").
	AddFile("file", fileBytes)
```

### NewFormBodyBuilder

```go
func NewFormBodyBuilder(builder *FormDataBuilder) (*BodyParam, error)
```

使用 FormDataBuilder 创建 BodyParam。

### Request.SetFormBody

```go
func (r *Request) SetFormBody(fields map[string]string, files map[string][]byte) IRequest
```

在请求上直接设置 form-data 体。

**返回值：** 返回 Request 本身，支持链式调用。

## 与现有功能的区别

| 功能                           | 适用场景                 | 是否包含文件 |
| ------------------------------ | ------------------------ | ------------ |
| `AddFormItem`                  | 简单表单字段             | ❌            |
| `AddFileBytes` / `AddFilePath` | 上传单个或多个文件       | ✅            |
| `SetFormBody` (新)             | 一次性设置完整 form-data | ✅            |

## 注意事项

1. **ContentType 自动设置**: 使用 `SetFormBody` 时，ContentType 会自动设置为 `multipart/form-data; boundary=...`
2. **内存消耗**: 文件内容存储在内存中，大文件可能消耗较多内存
3. **链式调用**: `FormDataBuilder` 支持链式调用以提高代码可读性
4. **错误处理**: 所有创建函数都会返回 error，需要正确处理

## 完整工作流示例

```go
package main

import (
	"context"
	"log"
	"github.com/pinealctx/restgo"
)

func main() {
	// 初始化客户端
	client := restgo.New(restgo.WithBaseURL("https://api.example.com"))
	
	// 准备数据
	fields := map[string]string{
		"user_id":   "12345",
		"timestamp": "2024-11-19",
	}
	
	files := map[string][]byte{
		"report": []byte("report content"),
	}
	
	// 创建请求
	req := restgo.NewRequest("POST", "/reports/submit")
	req.SetFormBody(fields, files)
	req.AddHeader("X-Custom-Header", "value")
	
	// 检查是否有错误
	if req.Err != nil {
		log.Fatalf("Failed to create request: %v", req.Err)
	}
	
	// 发送请求
	ctx := context.Background()
	resp, err := client.Do(ctx, req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	
	// 处理响应
	data, err := resp.Data()
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}
	
	log.Printf("Response: %s", string(data))
}
```
