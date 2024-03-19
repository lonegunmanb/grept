package pkg

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type gitIgnoreSuite struct {
	suite.Suite
	*testBase
}

func (s *gitIgnoreSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *gitIgnoreSuite) TearDownTest() {
	s.teardown()
}

func TestGitIgnoreData(t *testing.T) {
	suite.Run(t, new(gitIgnoreSuite))
}

func (s *gitIgnoreSuite) TestGitIgnore_Load() {
	t := s.T()
	ignoreContent := "# This is a comment\n*.log\n"
	s.dummyFsWithFiles([]string{".gitignore"}, []string{ignoreContent})

	// create GitIgnoreDatasource instance and load .gitignore content
	gitIgnore := &GitIgnoreDatasource{
		BaseBlock: &BaseBlock{
			c: NewGreptConfig(),
		},
	}

	err := gitIgnore.ExecuteDuringPlan()

	require.NoError(t, err)

	// only non-comment lines should be in Records
	s.Len(gitIgnore.Records, 1)
	s.Equal("*.log", gitIgnore.Records[0])
}

func (s *gitIgnoreSuite) TestGitIgnore_NoGitIgnoreFile() {
	t := s.T()

	// create GitIgnoreDatasource instance and load .gitignore content
	gitIgnore := &GitIgnoreDatasource{
		BaseBlock: &BaseBlock{
			c: NewGreptConfig(),
		},
	}

	err := gitIgnore.ExecuteDuringPlan()
	require.NoError(t, err)
	assert.Len(t, gitIgnore.Records, 0)
}

func (s *gitIgnoreSuite) TestGitIgnore_TabSpaceNewLine() {
	content := "\t\n   \n \t \n\t \t\n\n\r\n"
	s.dummyFsWithFiles([]string{".gitignore"}, []string{content})

	gitIgnore := &GitIgnoreDatasource{
		BaseBlock: &BaseBlock{
			c: NewGreptConfig(),
		},
	}
	err := gitIgnore.ExecuteDuringPlan()

	require.NoError(s.T(), err)
	s.Empty(gitIgnore.Records)
}
