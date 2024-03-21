package pkg

import (
	"github.com/stretchr/testify/suite"
	"io/fs"
	"testing"
)

type forEachTestSuite struct {
	suite.Suite
	*testBase
}

func TestForEachTestSuite(t *testing.T) {
	suite.Run(t, new(forEachTestSuite))
}

func (s *forEachTestSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *forEachTestSuite) SetupSubTest() {
	s.SetupTest()
}

func (s *forEachTestSuite) TearDownTest() {
	s.teardown()
}

func (s *forEachTestSuite) TearDownSubTest() {
	s.TearDownTest()
}

func (s *forEachTestSuite) TestForEachBlockWithAttributeThatHasDefaultValue() {
	config := `
	rule "must_be_true" "sample" {
		condition = false
	}
	
	fix "local_file" "sample" {
		for_each = toset([1,2,3])
		rule_ids = [rule.must_be_true.sample.id]
		paths = ["/file1.txt"]
		content = "Hello world!"
	}
`
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{config})
	c, err := BuildGreptConfig("", "", nil)
	s.NoError(err)
	_, err = RunGreptPlan(c)
	s.NoError(err)
	found := false
	for _, b := range blocks(c) {
		fix, ok := b.(*LocalFileFix)
		if ok {
			found = true
		} else {
			continue
		}
		s.Equal(fs.FileMode(644), *fix.Mode)
	}
	s.True(found)
}
