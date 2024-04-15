package ghttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Request struct {
	Method        string
	ReadTimeOut   time.Duration
	Proxy         string
	ProxyUsername string
	ProxyPassword string
	Url           string
	Headers       map[string]string
	Params        map[string]string
	Body          interface{}
	TlsCertFile   string
	TlsKeyFile    string
}

func (r *Request) DoReq() (respBytes []byte, httpStatusCode int, errRet error) {
	r.EnsureDefaults()

	ctx, cancel := context.WithTimeout(context.Background(), r.ReadTimeOut)
	defer cancel()

	var reader io.Reader

	if r.Body != nil {
		switch r.Body.(type) {
		case []byte:
			if len(r.Body.([]byte)) > 0 {
				reader = bytes.NewReader(r.Body.([]byte))
			}
		case string:
			if r.Body != "" {
				reader = bytes.NewReader([]byte(r.Body.(string)))
			}
		}
	} else if r.Params != nil {
		form := url.Values{}
		for k, v := range r.Params {
			form.Add(k, v)
		}
		r.Headers["Content-Type"] = "application/x-www-form-urlencoded"
		reader = strings.NewReader(form.Encode())
	}

	if req, err := http.NewRequestWithContext(ctx, strings.ToUpper(r.Method), r.Url, reader); err != nil {
		errRet = err
		return
	} else {
		if r.Headers != nil {
			for k, v := range r.Headers {
				req.Header.Set(k, v)
			}
		}

		var transport = &http.Transport{
			DisableKeepAlives: true,
		}
		if r.Proxy != "" {
			_url := &url.URL{
				Scheme: "http",
				Host:   r.Proxy,
			}
			if r.ProxyUsername != "" {
				_url.User = url.UserPassword(r.ProxyUsername, r.ProxyPassword)
			}
			transport.Proxy = http.ProxyURL(_url)
		}

		if r.TlsCertFile != "" && r.TlsKeyFile != "" {
			cert, err2 := tls.LoadX509KeyPair(r.TlsCertFile, r.TlsKeyFile)
			if err2 != nil {
				log.Fatal(err2)
			}
			tlsConfig := &tls.Config{
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: true,
			}
			transport.TLSClientConfig = tlsConfig
		}

		c := &http.Client{
			Transport: transport,
			Timeout:   r.ReadTimeOut,
		}
		defer c.CloseIdleConnections()
		if resp, err2 := c.Do(req); err2 != nil {
			errRet = err2
		} else {
			defer resp.Body.Close()
			httpStatusCode = resp.StatusCode
			respBytes, _ = io.ReadAll(resp.Body)
		}
	}
	return
}

func (r *Request) EnsureDefaults() {
	if r.ReadTimeOut == 0 {
		r.ReadTimeOut = 10 * time.Minute
	}
	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}
}
