package kubeagent

import (
	"crypto/tls"
	"encoding/json"
	"github.com/gorilla/websocket"
	"k8s.io/klog/v2"
	"net/http"
	"net/url"
	"time"
)

const (
	TunnelConnectPath  = "/api/v1/cluster/agent/connect"
	TunnelResponsePath = "/api/v1/cluster/agent/response"
)

type TunnelCallback interface {
	OnSuccess()
}

type ConnectOnSuccess func()

type Tunnel interface {
	Run(stopCh <-chan struct{})
	Receive() <-chan []byte
	Send(interface{})
}

type tunnel struct {
	token       string
	connUrl     *url.URL
	respUrl     *url.URL
	wsConn      *websocket.Conn
	receiveChan chan []byte
	callback    TunnelCallback
	stopped     bool
}

func NewTunnel(token string, serverHost string, callback TunnelCallback) Tunnel {
	return &tunnel{
		token:       token,
		connUrl:     &url.URL{Scheme: "ws", Host: serverHost, Path: TunnelConnectPath},
		respUrl:     &url.URL{Scheme: "ws", Host: serverHost, Path: TunnelResponsePath},
		callback:    callback,
		receiveChan: make(chan []byte),
	}
}

func (t *tunnel) Run(stopCh <-chan struct{}) {
	go func() {
		select {
		case <-stopCh:
			t.stopped = true
			if t.wsConn != nil {
				t.wsConn.Close()
			}
		}
	}()
	defer t.wsConn.Close()

	t.connect()
	for {
		_, data, err := t.wsConn.ReadMessage()
		if t.stopped {
			return
		}
		if err != nil {
			klog.Error("read err:", err)
			t.wsConn.Close()
			t.connect()
			continue
		}
		t.receiveChan <- data
	}
}

func (t *tunnel) Receive() <-chan []byte {
	return t.receiveChan
}

func (t *tunnel) Send(obj interface{}) {
	respMsg, err := json.Marshal(obj)
	if err != nil {
		klog.Errorf("response %v serializer error: %s", obj, err)
		return
	}

	klog.V(1).Info("start connect to server response websocket", t.connUrl.String())
	wsHeader := http.Header{}
	wsHeader.Add("token", t.token)

	d := &websocket.Dialer{TLSClientConfig: &tls.Config{RootCAs: nil, InsecureSkipVerify: true}}
	conn, _, err := d.Dial(t.respUrl.String(), wsHeader)

	if err != nil {
		klog.Errorf("connect to server %s error: %v", t.respUrl.String(), err)
	}
	defer conn.Close()
	if err = conn.WriteMessage(websocket.TextMessage, respMsg); err != nil {
		klog.Errorf("write message error: message=%s, error=%s", respMsg, err.Error())
		return
	}
	klog.V(1).Infof("write response %s success", string(respMsg))

}

func (t *tunnel) connect() {
	err := t.connectServer()
	if err == nil {
		return
	}
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := t.connectServer()
			if err == nil {
				return
			}
		}
	}
}

func (t *tunnel) connectServer() error {
	klog.Info("start connect to server ", t.connUrl.String())
	wsHeader := http.Header{}
	wsHeader.Add("token", t.token)
	d := &websocket.Dialer{TLSClientConfig: &tls.Config{RootCAs: nil, InsecureSkipVerify: true}}
	conn, _, err := d.Dial(t.connUrl.String(), wsHeader)
	if err != nil {
		klog.Infof("connect to server %s error: %v, retry after 5 seconds\n", t.connUrl.String(), err)
		return err
	} else {
		klog.Infof("connect to server %s success\n", t.connUrl.String())
		t.wsConn = conn
		return nil
	}
}
