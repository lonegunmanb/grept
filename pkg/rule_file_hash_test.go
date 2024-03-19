package pkg

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

type fileHashRuleSuite struct {
	suite.Suite
	*testBase
}

func TestFileHashRuleSuite(t *testing.T) {
	suite.Run(t, new(fileHashRuleSuite))
}

func (s *fileHashRuleSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *fileHashRuleSuite) TearDownTest() {
	s.teardown()
}

func (s *fileHashRuleSuite) TestFileHashRule_Check() {
	fs := s.fs
	t := s.T()
	// Write some test files
	filePaths := []string{"./file1.txt", "./file2.txt", "./file3.txt", "./pkg/sub/subfile1.txt"}
	fileContents := []string{"hello", "world", "golang", "world"}
	for i, filePath := range filePaths {
		_ = afero.WriteFile(fs, filePath, []byte(fileContents[i]), 0644)
	}

	// Calculate the md5 hash of "world"
	h := md5.New()
	h.Write([]byte("world"))
	expectedHash := h.Sum(nil)

	tests := []struct {
		name      string
		rule      *FileHashRule
		wantError bool
	}{
		{
			name: "matching file found",
			rule: &FileHashRule{
				BaseRule:  new(BaseRule),
				Glob:      "file*.txt",
				Hash:      fmt.Sprintf("%x", expectedHash),
				Algorithm: "md5",
			},
			wantError: false,
		},
		{
			name: "no matching file found",
			rule: &FileHashRule{
				BaseRule:  new(BaseRule),
				Glob:      "file*.txt",
				Hash:      fmt.Sprintf("%x", sha1.Sum([]byte("world1"))),
				Algorithm: "sha1",
			},
			wantError: true,
		},
		{
			name: "no matching glob pattern",
			rule: &FileHashRule{
				BaseRule:  new(BaseRule),
				Glob:      "nofile*.txt",
				Hash:      fmt.Sprintf("%x", expectedHash),
				Algorithm: "md5",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt.rule.ExecuteDuringPlan()
			checkError := tt.rule.CheckError()
			if (checkError != nil) != tt.wantError {
				t.Errorf("FileHashRule.ExecuteDuringPlan() error = %v, wantError %v", checkError, tt.wantError)
			}
		})
	}
}

func (s *fileHashRuleSuite) TestFileHashRule_Validate() {
	t := s.T()
	tests := []struct {
		name      string
		rule      *FileHashRule
		wantError bool
	}{
		{
			name: "valid rule",
			rule: &FileHashRule{
				BaseRule:  new(BaseRule),
				Glob:      "/file*.txt",
				Hash:      "abc123",
				Algorithm: "md5",
			},
			wantError: false,
		},
		{
			name: "missing glob",
			rule: &FileHashRule{
				BaseRule:  new(BaseRule),
				Hash:      "abc123",
				Algorithm: "md5",
			},
			wantError: true,
		},
		{
			name: "missing hash",
			rule: &FileHashRule{
				BaseRule:  new(BaseRule),
				Glob:      "/file*.txt",
				Algorithm: "md5",
			},
			wantError: true,
		},
		{
			name: "invalid algorithm",
			rule: &FileHashRule{
				BaseRule:  new(BaseRule),
				Glob:      "/file*.txt",
				Hash:      "abc123",
				Algorithm: "invalid",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("FileHashRule.Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func (s *fileHashRuleSuite) TestFileHashRule_HashMismatchFilesShouldBeExported() {
	fs := s.fs
	t := s.T()
	filename := "/example/sub1/testfile.txt"
	_ = afero.WriteFile(fs, filename, []byte("test content"), 0644)
	rule := &FileHashRule{
		BaseBlock: &BaseBlock{
			c: NewGreptConfig(),
		},
		BaseRule: new(BaseRule),
		Glob:     "/example/*/testfile.txt",
		Hash:     "non-matching-hash", // MD5 hash that doesn't match "test content"
	}
	runtimeErr := rule.ExecuteDuringPlan()
	assert.Nil(t, runtimeErr)
	checkErr := rule.CheckError()
	assert.NotNil(t, checkErr)
	assert.Contains(t, rule.HashMismatchFiles, filepath.FromSlash(filename))
}

func (s *fileHashRuleSuite) TestFileHashRule_FailOnHashMismatch() {
	fs := s.fs
	t := s.T()

	expectedContent := "hello"
	expectedHash := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" // SHA256 of "hello"
	_ = afero.WriteFile(fs, "/testfile", []byte(expectedContent), 0644)
	_ = afero.WriteFile(fs, "/example/sub1/testfile", []byte(expectedContent), 0644)
	_ = afero.WriteFile(fs, "/example/sub2/testfile", []byte("world"), 0644)
	_ = afero.WriteFile(fs, "/example2/sub1/testfile", []byte(expectedContent), 0644)
	_ = afero.WriteFile(fs, "/example2/sub2/testfile", []byte(expectedContent), 0644)

	tests := []struct {
		name      string
		rule      *FileHashRule
		wantErr   bool
		wantPaths []string
	}{
		{
			name: "FailOnHashMismatch is false, file content matches hash",
			rule: &FileHashRule{
				BaseBlock:          &BaseBlock{},
				BaseRule:           new(BaseRule),
				Glob:               "/testfile",
				Hash:               expectedHash,
				Algorithm:          "sha256",
				FailOnHashMismatch: false,
			},
			wantErr:   false,
			wantPaths: []string{},
		},
		{
			name: "FailOnHashMismatch is false, file content does not match hash",
			rule: &FileHashRule{
				BaseBlock:          &BaseBlock{},
				BaseRule:           new(BaseRule),
				Glob:               "/testfile",
				Hash:               "incorrecthash",
				Algorithm:          "sha256",
				FailOnHashMismatch: false,
			},
			wantErr:   true,
			wantPaths: []string{"/testfile"},
		},
		{
			name: "FailOnHashMismatch is false, one file content matches hash",
			rule: &FileHashRule{
				BaseBlock:          &BaseBlock{},
				BaseRule:           new(BaseRule),
				Glob:               "/example/*/testfile",
				Hash:               expectedHash,
				Algorithm:          "sha256",
				FailOnHashMismatch: false,
			},
			wantErr:   false,
			wantPaths: []string{"/example/sub2/testfile"},
		},
		{
			name: "FailOnHashMismatch is true, file content does not match hash",
			rule: &FileHashRule{
				BaseBlock:          &BaseBlock{},
				BaseRule:           new(BaseRule),
				Glob:               "/example/*/testfile",
				Hash:               "incorrecthash",
				Algorithm:          "sha256",
				FailOnHashMismatch: true,
			},
			wantErr:   true,
			wantPaths: []string{"/example/sub1/testfile", "/example/sub2/testfile"},
		},
		{
			name: "FailOnHashMismatch is true, file content matches hash exits, but still got file hash that mismatch",
			rule: &FileHashRule{
				BaseBlock:          &BaseBlock{},
				BaseRule:           new(BaseRule),
				Glob:               "/example/*/testfile",
				Hash:               expectedHash,
				Algorithm:          "sha256",
				FailOnHashMismatch: true,
			},
			wantErr:   true,
			wantPaths: []string{"/example/sub2/testfile"},
		},
		{
			name: "FailOnHashMismatch is false, all files match",
			rule: &FileHashRule{
				BaseBlock:          &BaseBlock{},
				BaseRule:           new(BaseRule),
				Glob:               "/example2/*/testfile",
				Hash:               expectedHash,
				Algorithm:          "sha256",
				FailOnHashMismatch: false,
			},
			wantErr:   false,
			wantPaths: []string{},
		},
		{
			name: "FailOnHashMismatch is false, all files match",
			rule: &FileHashRule{
				BaseBlock:          &BaseBlock{},
				BaseRule:           new(BaseRule),
				Glob:               "/example2/*/testfile",
				Hash:               expectedHash,
				Algorithm:          "sha256",
				FailOnHashMismatch: true,
			},
			wantErr:   false,
			wantPaths: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtimeErr := tt.rule.ExecuteDuringPlan()
			if runtimeErr != nil {
				t.Errorf("FileHashRule.ExecuteDuringPlan() runtime error = %+v", runtimeErr)
			}
			checkErr := tt.rule.CheckError()
			if (checkErr != nil) != tt.wantErr {
				t.Errorf("FileHashRule.ExecuteDuringPlan() error = %+v, wantErr %+v", checkErr, tt.wantErr)
			}
			var expectedPaths []string
			for i := 0; i < len(tt.rule.HashMismatchFiles); i++ {
				expectedPaths = append(expectedPaths, filepath.FromSlash(tt.rule.HashMismatchFiles[i]))
			}
			for _, path := range tt.wantPaths {
				assert.Contains(t, expectedPaths, filepath.FromSlash(path))
			}
		})
	}
}
