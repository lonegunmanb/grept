package pkg

import (
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"testing"
)

type gitIgnoreFixSuite struct {
	suite.Suite
	*testBase
}

func (s *gitIgnoreFixSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *gitIgnoreFixSuite) TearDownTest() {
	s.testBase.teardown()
}

func TestGitIgnoreFixSuite(t *testing.T) {
	suite.Run(t, new(gitIgnoreFixSuite))
}

func (s *gitIgnoreFixSuite) TestApplyGitIgnoreFix() {
	s.dummyFsWithFiles([]string{".gitignore"}, []string{"old_content\n"})

	gitIgnoreFix := &GitIgnoreFix{
		BaseBlock: &BaseBlock{
			c: &BaseConfig{basedir: "."},
		},
		Exist:    []string{"new_content"},
		NotExist: []string{"old_content"},
	}

	// Apply the fix
	err := gitIgnoreFix.Apply()
	s.NoError(err)

	content, err := afero.ReadFile(s.fs, ".gitignore")
	s.NoError(err)
	s.Contains(string(content), "new_content")
	s.NotContains(string(content), "old_content")
}

func (s *gitIgnoreFixSuite) TestGitIgnoreFixEnsureExist() {
	s.dummyFsWithFiles([]string{".gitignore"}, []string{"content\n"})

	gitIgnoreFix := &GitIgnoreFix{
		BaseBlock: &BaseBlock{
			c: &BaseConfig{basedir: "."},
		},
		Exist: []string{"new_content"},
	}

	// Apply the fix
	err := gitIgnoreFix.Apply()
	s.NoError(err)

	content, err := afero.ReadFile(s.fs, ".gitignore")
	s.NoError(err)
	s.Contains(string(content), "new_content")
	s.Contains(string(content), "content")
}

func (s *gitIgnoreFixSuite) TestGitIgnoreFix_NotExistWontRemoveComment() {
	s.dummyFsWithFiles([]string{".gitignore"}, []string{`
#comment
content
`})

	gitIgnoreFix := &GitIgnoreFix{
		BaseBlock: &BaseBlock{
			c: &BaseConfig{basedir: "."},
		},
		NotExist: []string{"content"},
	}

	// Apply the fix
	err := gitIgnoreFix.Apply()
	s.NoError(err)

	content, err := afero.ReadFile(s.fs, ".gitignore")
	s.NoError(err)
	s.Contains(string(content), "#comment")
	s.NotContains(string(content), "content")
}

func (s *gitIgnoreFixSuite) TestGitIgnoreFix_TrimItem() {
	s.dummyFsWithFiles([]string{".gitignore"}, []string{"\r\n#comment\r\n\t.terraform \r\n\t terraform.tfstate \r\n"})

	gitIgnoreFix := &GitIgnoreFix{
		BaseBlock: &BaseBlock{
			c: &BaseConfig{basedir: "."},
		},
		Exist:    []string{" new_content\t\n", "another_content\r\n"},
		NotExist: []string{".terraform"},
	}

	// Apply the fix
	err := gitIgnoreFix.Apply()
	s.NoError(err)

	content, err := afero.ReadFile(s.fs, ".gitignore")
	s.NoError(err)
	s.Contains(string(content), "#comment\n")
	s.Contains(string(content), "\t terraform.tfstate \n")
	s.NotContains(string(content), ".terraform")
	s.Contains(string(content), "\nnew_content\n")
	s.Contains(string(content), "\nanother_content\n")
}

func (s *gitIgnoreFixSuite) TestGitIgnoreFix_GitIgnoreFileAbsent() {
	gitIgnoreFix := &GitIgnoreFix{
		BaseBlock: &BaseBlock{
			c: &BaseConfig{basedir: "."},
		},
		Exist:    []string{"content"},
		NotExist: []string{"new_content"},
	}

	// Apply the fix
	err := gitIgnoreFix.Apply()
	s.NoError(err)

	// ExecuteDuringPlan that the .gitignore file contains the correct content
	content, err := afero.ReadFile(s.fs, ".gitignore")
	s.NoError(err)
	s.Contains(string(content), "content")
	s.NotContains(string(content), "new_content")
}
