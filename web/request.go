package web

import (
	"encoding/json"
	"github.com/chuccp/http2smtp/login"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type HandlerFunc func(req *Request) (any, error)

type Request struct {
	context    *gin.Context
	digestAuth *login.DigestAuth
}

func NewRequest(context *gin.Context, digestAuth *login.DigestAuth) *Request {
	return &Request{context: context, digestAuth: digestAuth}
}
func (r *Request) GetDigestAuth() *login.DigestAuth {
	return r.digestAuth
}
func (r *Request) ShouldBindBodyWithJSON(obj any) error {
	return r.context.ShouldBindBodyWith(obj, binding.JSON)
}

func (r *Request) GetContext() *gin.Context {
	return r.context
}
func (r *Request) FormValue(key string) string {
	return r.context.Request.FormValue(key)
}
func (r *Request) FormIntValue(key string) int {
	v := r.FormValue(key)
	i, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return i
}
func (r *Request) FormInt64Value(key string) int64 {
	v := r.FormValue(key)
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0
	}
	return i
}
func (r *Request) GetRemoteAddress() string {
	address := r.context.Request.RemoteAddr
	index := strings.Index(address, "_")
	if index > 0 {
		return address[:index]
	}
	return address
}
func (r *Request) GetPage() *Page {
	var page Page
	page.PageNo = r.FormIntValue("pageNo")
	page.PageSize = r.FormIntValue("pageSize")
	return &page
}
func (r *Request) GetSearchKey() string {
	return r.FormValue("searchKey")
}
func (r *Request) GetRawRequest() *http.Request {
	return r.context.Request
}
func (r *Request) GetHeader(key string) string {
	return r.context.Request.Header.Get(key)
}
func (r *Request) GetResponseWriter() http.ResponseWriter {
	return r.context.Writer
}

func (r *Request) BodyJson(v any) ([]byte, error) {
	body, err := io.ReadAll(r.context.Request.Body)
	if err != nil {
		return body, err
	}
	err = json.Unmarshal(body, v)
	if err != nil {
		return body, err
	}
	return body, nil
}

func (r *Request) FormFile(name string) (*multipart.FileHeader, error) {
	return r.context.FormFile(name)
}

func (r *Request) MultipartForm() (*multipart.Form, error) {
	return r.context.MultipartForm()
}
func (r *Request) IsForm() bool {
	return strings.Contains(r.context.GetHeader("Content-Type"), "application/x-www-form-urlencoded")
}
func (r *Request) IsMultipartForm() bool {
	return strings.Contains(r.context.GetHeader("Content-Type"), "multipart/form-data")
}

func (r *Request) Param(key string) string {
	return r.context.Param(key)
}
func (r *Request) Header(key, value string) {
	r.context.Header(key, value)
}
func (r *Request) Status(code int) {
	r.context.Status(code)
}
func (r *Request) String(code int, format string, values ...any) {
	r.context.String(code, format, values...)
}

func ToGinHandlerFuncs(handlers []HandlerFunc, digestAuth *login.DigestAuth) []gin.HandlerFunc {
	var handlerFunc = make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handlerFunc[i] = ToGinHandlerFunc(handler, digestAuth)
	}
	return handlerFunc
}
func ToGinHandlerFunc(handler HandlerFunc, digestAuth *login.DigestAuth) gin.HandlerFunc {
	handlerFunc := func(context *gin.Context) {
		value, err := handler(NewRequest(context, digestAuth))
		if err != nil {
			context.Status(500)
			context.Writer.Write([]byte(err.Error()))
			context.Abort()
		} else {
			if value != nil {
				switch t := value.(type) {
				case string:
					context.Writer.Write([]byte(t))
				case *File:
					if len(t.GetFilename()) == 0 {
						_, filename := path.Split(t.Path)
						t.FileName = filename
					}
					context.FileAttachment(t.GetPath(), t.GetFilename())
				default:
					context.AbortWithStatusJSON(200, t)
				}
			}
		}
	}
	return handlerFunc
}

// SaveUploadedFile 将上传的文件保存到指定路径
func SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	// 打开上传的临时文件
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer func() {
		// 确保临时文件关闭，并捕获可能的错误
		closeErr := src.Close()
		if err == nil {
			err = closeErr
		}
	}()

	// 创建目标目录
	if err = os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}

	// 创建目标文件
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		// 确保目标文件关闭，并捕获可能的错误
		closeErr := out.Close()
		if err == nil {
			err = closeErr
		}
	}()

	// 复制文件内容
	if _, err = io.Copy(out, src); err != nil {
		return err
	}

	// 强制将数据刷新到磁盘，确保数据写入完成
	if err = out.Sync(); err != nil {
		return err
	}

	return nil
}
