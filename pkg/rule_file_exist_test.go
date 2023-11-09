package pkg

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type fileExistRuleSuite struct {
	suite.Suite
	*testBase
}

func TestFileExistRuleSuite(t *testing.T) {
	suite.Run(t, new(fileExistRuleSuite))
}

func (s *fileExistRuleSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *fileExistRuleSuite) TearDownTest() {
	s.teardown()
}

func (s *fileExistRuleSuite) TestFileExistRule_Check() {
	s.dummyFsWithFiles([]string{"./file1.txt", "./file2.txt", "./file3.txt", "./pkg/sub/subfile1.txt"}, []string{"content", "content", "content", "content"})

	tests := []struct {
		name      string
		rule      *FileExistRule
		wantError bool
	}{
		{
			name: "file exists",
			rule: &FileExistRule{
				BaseBlock: &BaseBlock{
					name: "test",
					id:   "test",
				},
				Glob: "./file1.txt",
			},
			wantError: false,
		},
		{
			name: "file does not exist",
			rule: &FileExistRule{
				BaseBlock: &BaseBlock{
					name: "test",
					id:   "test",
				},
				Glob: "./nofile.txt",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			checkError, runtimeErr := tt.rule.Check()
			s.NoError(runtimeErr)
			if tt.wantError {
				s.NotNil(checkError)
			} else {
				s.NoError(checkError)
			}
		})
	}
}