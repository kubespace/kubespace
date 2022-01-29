package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"k8s.io/klog"
	"net/http"
	"net/url"
	"path"
)

type HttpClient struct {
	client  *http.Client
	baseUrl string
	ctx     context.Context
	headers http.Header
}

func NewHttpClient(baseUrl string) (*HttpClient, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := HttpClient{client: &http.Client{Transport: tr}}
	u, err := url.Parse(baseUrl)
	if err != nil {
		klog.Error("http request url parse error: httpUrl=%s. error=%v", baseUrl, err)
		return nil, err
	}
	httpClient.baseUrl = u.String()
	httpClient.headers = http.Header{}
	httpClient.headers.Set("Content-Type", "application/json")
	httpClient.headers.Set("Accept", "application/json")
	httpClient.ctx = context.Background()
	return &httpClient, nil
}

func (c *HttpClient) Get(getPath string, params map[string]string) ([]byte, error) {
	getUrl := c.addParamsToPath(getPath, params)
	return c.doRequest(getUrl, "GET", nil)
}

func (c *HttpClient) Delete(deletePath string, params map[string]string) ([]byte, error) {
	getUrl := c.addParamsToPath(deletePath, params)
	return c.doRequest(getUrl, "DELETE", nil)
}

func (c *HttpClient) Post(postPath string, params map[string]string, body []byte) ([]byte, error) {
	postUrl := c.addParamsToPath(postPath, params)
	return c.doRequest(postUrl, "POST", bytes.NewBuffer(body))
}

func (c *HttpClient) doRequest(url string, method string, buf io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		klog.Errorf("get http request error: error=%v, url=%s, method=%s", err, url, method)
		return nil, err
	}
	req.Header = c.headers
	req = req.WithContext(c.ctx)
	resp, err := c.client.Do(req)
	if err != nil {
		klog.Errorf("send http req error: error=%s", err)
		return nil, err
	} else {
		data, errReadBody := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if errReadBody != nil {
			klog.Error("read received http resp body error: error=", err)
			return nil, err
		}
		klog.Infof("doRequest get response: url=%s, method=%s, status=%s", url, method, resp.StatusCode)
		if resp.StatusCode != http.StatusOK {
			klog.Errorf("receive http code not 200: httpcode=%d", resp.StatusCode)
			return data, fmt.Errorf("status code %v", resp.StatusCode)
		} else {
			return data, nil
		}
	}
}

func (c *HttpClient) addParamsToPath(oriPath string, params map[string]string) string {
	oriUrl, _ := url.Parse(c.baseUrl)
	oriUrl.Path = path.Join(oriUrl.Path, oriPath)
	if params != nil && len(params) != 0 {
		urlParam := url.Values{}
		for k, v := range params {
			urlParam.Set(k, v)
		}
		oriUrl.RawQuery = urlParam.Encode()
	}
	return oriUrl.String()
}
