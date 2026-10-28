package util

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Request struct {
	client *http.Client
}
type Response struct {
	StatusCode int
	Body       []byte
}

func createResponse(StatusCode int, body []byte) *Response {
	return &Response{StatusCode: StatusCode, Body: body}
}

func NewRequest() *Request {
	ct := http.Client{Timeout: time.Second * 2, Transport: http.DefaultTransport}
	return &Request{client: &ct}
}

func (r *Request) CallApiForResponse(link string, header map[string]string, method string, body []byte) (*Response, error) {
	req, err := http.NewRequest(method, link, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return createResponse(resp.StatusCode, all), nil
}

func ValidateURL(urlStr string) error {

	// 空字符串无效
	if strings.TrimSpace(urlStr) == "" {
		return errors.New("URL cannot be empty")
	}
	// 尝试解析URL
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return errors.New("URL must start with http or https")
	}

	if u.Host == "" {
		return errors.New("URL must have a host")
	}
	return nil
}
