package pkg

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/net/http/httpproxy"
)

var _ Data = &HttpDatasource{}

type HttpDatasource struct {
	*BaseBlock
	*BaseData
	Url             string            `hcl:"url"`
	Method          string            `hcl:"method,optional" default:"GET"`
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
	retryClient.Logger = log.New(os.Stderr, fmt.Sprintf("data.http.%s:", h.name), log.LstdFlags)
	retryClient.HTTPClient.Transport = clonedTr
	retryClient.RetryMax = h.RetryMax
	request, err := retryablehttp.NewRequestWithContext(h.Context(), h.Method, h.Url, strings.NewReader(h.RequestBody))
	if err != nil {
		return fmt.Errorf("error creating request data.http.%s, %s", h.name, err.Error())
	}
	for k, v := range h.RequestHeaders {
		request.Header.Set(k, v)
	}
	response, err := retryClient.Do(request)

	if err != nil {
		return fmt.Errorf("error making request data.http.%s, detail: %s", h.name, err.Error())
	}
	defer func() { _ = response.Body.Close() }()
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error reading response body data.http.%s, detail: %s", h.name, err.Error())
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
		"url":              ToCtyValue(h.Url),
		"method":           ToCtyValue(h.Method),
		"request_body":     ToCtyValue(h.RequestBody),
		"response_body":    ToCtyValue(h.ResponseBody),
		"status_code":      ToCtyValue(int64(h.StatusCode)),
		"request_headers":  ToCtyValue(h.RequestHeaders),
		"response_headers": ToCtyValue(h.ResponseHeaders),
	}
}

var validHttpMethods = hashset.New("GET", "HEAD", "POST", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE", "PATCH")

func (h *HttpDatasource) Validate() error {
	var err error
	if !validHttpMethods.Contains(h.Method) {
		err = multierror.Append(err, fmt.Errorf(`"method"" must be one of "GET", "HEAD", "POST", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE", "PATCH"`))
	}
	return err
}
