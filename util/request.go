package util

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type RetryTransport struct {
	Transport http.RoundTripper
	Retries   int
}

func (r *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 0; i <= r.Retries; i++ {
		resp, err = r.Transport.RoundTrip(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return resp, err
}

type netBreak struct {
	isBreak   bool
	breakTime time.Time
	lock      sync.Mutex
	err       error
	num       int32
}

func (nb *netBreak) beBreak(err error) {
	nb.lock.Lock()
	defer nb.lock.Unlock()
	nb.isBreak = true
	nb.breakTime = time.Now()
	nb.err = err
}
func (nb *netBreak) noBreak() {
	if nb.isBreak {
		nb.lock.Lock()
		defer nb.lock.Unlock()
		nb.isBreak = false
	}
}
func (nb *netBreak) hasBreak() (error, bool) {
	nb.lock.Lock()
	defer nb.lock.Unlock()
	ti := time.Now().Add(time.Second * -5)
	if ti.After(nb.breakTime) {
		return nil, false
	}
	return nb.err, nb.isBreak
}

type Request struct {
	client   *http.Client
	netBreak *netBreak
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
	return &Request{client: &ct, netBreak: &netBreak{isBreak: false}}
}

func (r *Request) CallBreak(link string, jsonData []byte) ([]byte, error) {
	err, b := r.netBreak.hasBreak()
	if b {
		return nil, err
	}
	if r.netBreak.isBreak {
		if !atomic.CompareAndSwapInt32(&r.netBreak.num, 0, 1) {
			return nil, r.netBreak.err
		}
	}
	call, err := r.Call(link, jsonData)
	atomic.StoreInt32(&r.netBreak.num, 0)
	if err != nil {
		if strings.Contains(err.Error(), "No connection could be made because the target machine actively refused it") {
			r.netBreak.beBreak(err)
		}
		return nil, err
	}
	return call, nil
}
func (r *Request) Call(link string, jsonData []byte) ([]byte, error) {
	var buff = new(bytes.Buffer)
	buff.Write(jsonData)
	resp, err := r.client.Post(link, "application/json", buff)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}
	r.netBreak.noBreak()
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return all, nil
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

func (r *Request) CallApi(link string, header map[string]string, method string, body []byte) ([]byte, error) {
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
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return all, nil

}

func (r *Request) Get(link string) ([]byte, error) {
	resp, err := r.client.Get(link)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	r.netBreak.noBreak()
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return all, nil
}
func (r *Request) JustCall(link string, jsonData []byte) error {
	var buff = new(bytes.Buffer)
	buff.Write(jsonData)
	resp, err := r.client.Post(link, "application/json", buff)
	if err == nil {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {

			}
		}(resp.Body)
		r.netBreak.noBreak()
	}
	return err
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
