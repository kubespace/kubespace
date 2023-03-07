package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/google/go-querystring/query"
	"io"
	"io/ioutil"
	"k8s.io/klog/v2"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
)

type HttpClient struct {
	client  *http.Client
	baseUrl *url.URL
}

func NewHttpClient(baseUrl string) (*HttpClient, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	u, err := url.Parse(baseUrl)
	if err != nil {
		klog.Errorf("http request url parse error: httpUrl=%s. error=%v", baseUrl, err)
		return nil, err
	}
	return &HttpClient{
		client:  &http.Client{Transport: tr},
		baseUrl: u,
	}, nil
}

type RequestOptions struct {
	Header  http.Header
	Context context.Context
}

func (o *RequestOptions) WithContext(ctx context.Context) {
	o.Context = ctx
}

func (o *RequestOptions) WithHeader(name, value string) {
	if o.Header == nil {
		o.Header = make(http.Header)
	}
	o.Header.Set(name, value)
}

func (o *RequestOptions) WithHeaders(headers map[string]string) {
	if o.Header == nil {
		o.Header = make(http.Header)
	}
	for k, v := range headers {
		o.Header.Set(k, v)
	}
}

func (c *HttpClient) Get(path string, query interface{}, v interface{}, options RequestOptions) (*http.Response, error) {
	req, err := c.NewRequest(http.MethodGet, path, query, options)
	if err != nil {
		return nil, err
	}
	return c.Do(req, v)
}

func (c *HttpClient) Delete(path string, body interface{}, v interface{}, options RequestOptions) (*http.Response, error) {
	req, err := c.NewRequest(http.MethodDelete, path, body, options)
	if err != nil {
		return nil, err
	}
	return c.Do(req, v)
}

func (c *HttpClient) Post(path string, body interface{}, v interface{}, options RequestOptions) (*http.Response, error) {
	req, err := c.NewRequest(http.MethodPost, path, body, options)
	if err != nil {
		return nil, err
	}
	return c.Do(req, v)
}

func (c *HttpClient) Put(path string, body interface{}, v interface{}, options RequestOptions) (*http.Response, error) {
	req, err := c.NewRequest(http.MethodPut, path, body, options)
	if err != nil {
		return nil, err
	}
	return c.Do(req, v)
}

func (c *HttpClient) NewRequest(method, reqPath string, params interface{}, options RequestOptions) (*http.Request, error) {
	u := *c.baseUrl
	unescaped, err := url.PathUnescape(reqPath)
	if err != nil {
		return nil, err
	}
	// 替换url路径
	u.RawPath = path.Join(c.baseUrl.Path, unescaped)
	u.Path = path.Join(c.baseUrl.Path, reqPath)

	headers := make(http.Header)
	// 设置默认接收类型
	headers.Set("Accept", "application/json")
	var body io.Reader
	switch {
	case method == http.MethodPost || method == http.MethodPut:
		headers.Set("Content-Type", "application/json")

		if params != nil {
			paramBytes, err := json.Marshal(params)
			if err != nil {
				return nil, err
			}
			body = bytes.NewBuffer(paramBytes)
		}
	case params != nil:
		q, err := query.Values(params)
		if err != nil {
			return nil, err
		}
		u.RawQuery = q.Encode()
	}
	req, err := http.NewRequest(method, u.String(), body)
	if options.Context != nil {
		req.WithContext(options.Context)
	}
	if err != nil {
		klog.Errorf("get http request error: error=%v, url=%s, method=%s", err, u.String(), method)
		return nil, err
	}
	// 覆盖Header
	for k, v := range options.Header {
		headers[k] = v
	}
	req.Header = headers
	return req, nil
}

func (c *HttpClient) Do(req *http.Request, v interface{}) (*http.Response, error) {
	r, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	switch r.StatusCode {
	case 200, 201, 202, 204, 304:
		if v != nil {
			if w, ok := v.(io.Writer); ok {
				_, err = io.Copy(w, r.Body)
			} else {
				err = json.NewDecoder(r.Body).Decode(v)
			}
		}
		return r, err
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	reqUrl := req.URL.Scheme + req.URL.Host + req.URL.Path
	return r, fmt.Errorf("%s %s error: status_code=%d, %s", req.Method, reqUrl, r.StatusCode, string(body))

}

func PostFile(url, filename, filepath string) ([]byte, error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile(filename, filepath)
	if err != nil {
		fmt.Println("error writing to buffer")
		return nil, err
	}

	// 打开文件句柄操作
	fh, err := os.Open(filepath)
	if err != nil {
		fmt.Println("error opening file")
		return nil, err
	}
	defer fh.Close()

	// iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return nil, err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(url, contentType, bodyBuf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp.Status)
	fmt.Println(string(respBody))
	return respBody, nil
}
