package golden

import (
	"fmt"
	"github.com/Azure/grept/pkg"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type dagSuite struct {
	suite.Suite
	*golden.testBase
	server *httptest.Server
}

func TestDagSuite(t *testing.T) {
	suite.Run(t, new(dagSuite))
}

func (s *dagSuite) SetupTest() {
	s.testBase = golden.newTestBase()
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Expected content"))
	}))
}

func (s *dagSuite) TearDownTest() {
	s.teardown()
	s.server.Close()
}

func (s *dagSuite) TestDag_DagVertex() {
	content := fmt.Sprintf(`
	data "http" sample {
	    url = "%s"
    }

	rule "file_hash" sample {  
		glob = "*.txt"  
		hash = "abc123"  
		algorithm = "sha256"  
	}  
  
	fix "local_file" hello_world{  
		rule_ids = [rule.file_hash.sample.id]
		paths = ["/path/to/file.txt"]  
		content = "Hello, world!"
	}  
	`, s.server.URL)

	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{content})

	config, err := BuildGreptConfig("", "", nil)
	s.NoError(err)
	d := newDag()
	err = d.buildDag(blocks(config))
	s.NoError(err)
	s.Len(d.GetVertices(), 3)

	assertVertex(s.T(), d, "rule.file_hash.sample")
	assertVertex(s.T(), d, "data.http.sample")
	assertVertex(s.T(), d, "fix.local_file.hello_world")
}

func (s *dagSuite) TestDag_DagBlocksShouldBeConnectedWithEdgeIfThereIsReferenceBetweenTwoBlocks() {
	t := s.T()
	content := fmt.Sprintf(`
	data "http" sample {
	    url = "%s"
    }

	rule "file_hash" sample {  
		glob = "*.txt"  
		hash = sha256(data.http.sample.response_body)  
		algorithm = "sha256"  
	}  
  
	fix "local_file" hello_world{  
		rule_ids = [rule.file_hash.sample.id]
		paths = ["/path/to/file.txt"]  
		content = "Hello, world!"
	}  
	`, s.server.URL)

	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{content})

	config, err := BuildGreptConfig("", "", nil)
	require.NoError(t, err)
	dag := newDag()
	err = dag.buildDag(blocks(config))
	require.NoError(t, err)
	assert.Equal(t, 2, dag.GetSize())
	roots := dag.GetRoots()
	assert.Len(t, roots, 1)
	assertEdge(t, dag, "data.http.sample", "rule.file_hash.sample")
	assertEdge(t, dag, "rule.file_hash.sample", "fix.local_file.hello_world")
}

func (s *dagSuite) TestDag_CycleDependencyShouldCauseError() {
	content := `
	data "http" sample {
	    url = data.http.sample2.url
    }

	data "http" sample2 {
		url = data.http.sample.url
    }
	`

	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{content})

	_, err := BuildGreptConfig("", "", nil)
	s.NotNil(err)
	// The error message must contain both of two blocks' address so we're sure that it's about the loop.
	s.Contains(err.Error(), "data.http.sample")
	s.Contains(err.Error(), "data.http.sample2")
}

func assertEdge(t *testing.T, dag *Dag, src, dest string) {
	from, err := dag.GetParents(dest)
	assert.NoError(t, err)
	_, ok := from[src]
	assert.True(t, ok, "cannot find edge from %s to %s", src, dest)
}

func assertVertex(t *testing.T, dag *Dag, address string) {
	b, err := dag.GetVertex(address)
	assert.NoError(t, err)
	bb, ok := b.(Block)
	require.True(t, ok)
	split := strings.Split(address, ".")
	name := split[len(split)-1]
	assert.Equal(t, name, bb.HclBlock().Labels[1])
}
