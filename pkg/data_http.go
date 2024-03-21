package pkg

import (
	"crypto/tls"
	"fmt"
	"github.com/Azure/grept/golden"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/net/http/httpproxy"
)

var _ Data = &HttpDatasource{}

type HttpDatasource struct {
	*golden.BaseBlock
	*BaseData
	Url string `hcl:"url"`
	/*
		if !validHttpMethods.Contains(h.Method) {
				err = multierror.Append(err, fmt.Errorf(`"method"" must be one of "GET", "HEAD", "POST", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE", "PATCH"`))
			}
	*/
	Method          string            `hcl:"method,optional" default:"GET" validate:"oneof=GET HEAD POST PUT DELETE CONNECT OPTIONS TRACE PATCH"`
	RequestBody     string            `hcl:"request_body,optional"`
	RequestHeaders  map[string]string `hcl:"request_headers,optional"`
	RetryMax        int               `hcl:"retry_max,optional" default:"4"`
	ResponseBody    string            `attribute:"response_body"`
	ResponseHeaders map[string]string `attribute:"response_headers"`
	StatusCode      int               `attribute:"status_code"`
}

func (h *HttpDatasource) ExecuteDuringPlan() error {
	tr, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return fmt.Errorf("error http: can't configure http transport")
	}
	clonedTr := tr.Clone()

	// Prevent issues with tests caching the proxy configuration.
	clonedTr.Proxy = func(req *http.Request) (*url.URL, error) {
		return httpproxy.FromEnvironment().ProxyFunc()(req.URL)
	}

	if clonedTr.TLSClientConfig == nil {
		clonedTr.TLSClientConfig = &tls.Config{}
	}

	retryClient := retryablehttp.NewClient()
	retryClient.Logger = log.New(os.Stderr, fmt.Sprintf("%s:", h.Address()), log.LstdFlags)
	retryClient.HTTPClient.Transport = clonedTr
	retryClient.RetryMax = h.RetryMax
	request, err := retryablehttp.NewRequestWithContext(h.Context(), h.Method, h.Url, strings.NewReader(h.RequestBody))
	if err != nil {
		return fmt.Errorf("error creating request %s, %s", h.Address(), err.Error())
	}
	for k, v := range h.RequestHeaders {
		request.Header.Set(k, v)
	}
	response, err := retryClient.Do(request)

	if err != nil {
		return fmt.Errorf("error making request %s, detail: %s", h.Address(), err.Error())
	}
	defer func() { _ = response.Body.Close() }()
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error reading response body %s, detail: %s", h.Address(), err.Error())
	}
	h.ResponseBody = string(bytes)
	h.ResponseHeaders = make(map[string]string)
	for k, v := range response.Header {
		// Concatenate according to RFC9110 https://www.rfc-editor.org/rfc/rfc9110.html#section-5.2
		h.ResponseHeaders[k] = strings.Join(v, ", ")
	}
	h.StatusCode = response.StatusCode
	return nil
}

func (h *HttpDatasource) Type() string {
	return "http"
}

func (h *HttpDatasource) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"url":              golden.ToCtyValue(h.Url),
		"method":           golden.ToCtyValue(h.Method),
		"request_body":     golden.ToCtyValue(h.RequestBody),
		"response_body":    golden.ToCtyValue(h.ResponseBody),
		"status_code":      golden.ToCtyValue(int64(h.StatusCode)),
		"request_headers":  golden.ToCtyValue(h.RequestHeaders),
		"response_headers": golden.ToCtyValue(h.ResponseHeaders),
	}
}
