package pkg

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type httpDataSuite struct {
	suite.Suite
	*testBase
}

func TestHttpDataSuite(t *testing.T) {
	suite.Run(t, new(httpDataSuite))
}

func (s *httpDataSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *httpDataSuite) TearDownTest() {
	s.teardown()
}

func (s *httpDataSuite) TestHttpDatasource_Load() {
	t := s.T()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock response
		w.Header().Set("Content-Type", "text/plain")
		_, _ = fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	// Create a HttpDatasource instance
	h := &HttpDatasource{
		BaseBlock: &BaseBlock{
			c:    &GreptConfig{BaseConfig: NewBasicConfig(".", context.TODO())},
			name: "test",
		},
		Url:    ts.URL,
		Method: "GET",
	}

	err := h.ExecuteDuringPlan()

	// Assert no error from ExecuteDuringPlan function
	require.NoError(t, err)
	// Assert the response body
	assert.Equal(t, "Hello, client\n", h.ResponseBody)
	assert.Equal(t, 200, h.StatusCode)
	assert.Equal(t, ts.URL, h.Url)
	assert.True(t, len(h.ResponseHeaders) > 1)
	ct, ok := h.ResponseHeaders["Content-Type"]
	assert.True(t, ok)
	assert.Equal(t, "text/plain", ct)
}

func (s *httpDataSuite) TestHttpDatasource_InvalidMethod() {
	content := `
	data "http" foo {
		url = "http://foo"
		method = "foo"
	}
	`

	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{content})

	config, err := BuildGreptConfig("/", "", nil)
	s.NoError(err)
	_, err = RunGreptPlan(config)
	s.NotNil(err)
	s.Contains(err.Error(), "valid")

}
