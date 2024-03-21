package pkg

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	iofs "io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type localFileFixSuite struct {
	suite.Suite
	*testBase
}

func TestLocalFileFix(t *testing.T) {
	suite.Run(t, new(localFileFixSuite))
}

func (s *localFileFixSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *localFileFixSuite) SetupSubTest() {
	s.SetupTest()
}

func (s *localFileFixSuite) TearDownTest() {
	s.teardown()
}

func (s *localFileFixSuite) TearDownSubTest() {
	s.TearDownTest()
}

func (s *localFileFixSuite) TestLocalFile_ApplyFix_CreateNewFile() {
	fs := s.fs
	t := s.T()
	mode := iofs.FileMode(0644)
	fix := &LocalFileFix{
		Paths:   []string{"/file1.txt"},
		Content: "Hello, world!",
		Mode:    &mode,
	}

	err := fix.Apply()
	assert.NoError(t, err)

	// ExecuteDuringPlan that the file was created with the correct content
	content, err := afero.ReadFile(fs, fix.Paths[0])
	assert.NoError(t, err)
	assert.Equal(t, fix.Content, string(content))
}

func (s *localFileFixSuite) TestLocalFile_ApplyFix_OverwriteExistingFile() {
	fs := s.fs
	t := s.T()
	path := "/file1.txt"
	mode := iofs.FileMode(0644)
	fix := &LocalFileFix{
		BaseBlock: &BaseBlock{},
		Paths:     []string{path},
		Content:   "Hello, world!",
		Mode:      &mode,
	}

	// Create the file first
	err := fix.Apply()
	assert.NoError(t, err)

	// Now overwrite it
	fix.Content = "New content"
	err = fix.Apply()
	assert.NoError(t, err)

	// ExecuteDuringPlan that the file was overwritten with the correct content
	content, err := afero.ReadFile(fs, path)
	assert.NoError(t, err)
	assert.Equal(t, fix.Content, string(content))
}

//nolint:all
func TestLocalFile_ApplyFix_FileInSubFolder(t *testing.T) {
	path := filepath.Join(os.TempDir(), uuid.NewString())
	defer func() {
		_ = os.RemoveAll(path)
	}()
	filePath := filepath.Join(path, "a", "b", "tmp")
	//must use 644 here
	mode := iofs.FileMode(644)
	fix := &LocalFileFix{
		BaseBlock: &BaseBlock{},
		Paths:     []string{filePath},
		Content:   "Hello, world!",
		Mode:      &mode,
	}

	fs := FsFactory()
	// Create the file first
	err := fix.Apply()
	require.NoError(t, err)
	content, err := afero.ReadFile(fs, filePath)
	assert.Equal(t, "Hello, world!", string(content))

	// Now overwrite it
	fix.Content = "New content"
	err = fix.Apply()
	require.NoError(t, err)

	// ExecuteDuringPlan that the file was overwritten with the correct content
	content, err = afero.ReadFile(fs, filePath)
	require.NoError(t, err)
	assert.Equal(t, fix.Content, string(content))
}

func (s *localFileFixSuite) TestLocalFile_ApplyFix_FileMode() {
	var pString = func(s string) *string {
		return &s
	}
	cases := []struct {
		desc                 string
		mode                 *string
		wanted               iofs.FileMode
		expectedErrorMessage *string
	}{
		{
			desc:   "no assignment",
			mode:   nil,
			wanted: iofs.FileMode(0644),
		},
		{
			desc:   "customized mode",
			mode:   pString("0777"),
			wanted: iofs.ModePerm,
		},
		{
			desc:   "explicit null",
			mode:   pString("null"),
			wanted: iofs.FileMode(0644),
		},
		{
			desc:                 "invalid mode1",
			mode:                 pString("0778"),
			expectedErrorMessage: pString("file_mode"),
		},
	}

	for i := 0; i < len(cases); i++ {
		sc := cases[i]
		s.Run(sc.desc, func() {
			var assignment = ""
			if sc.mode != nil {
				assignment = fmt.Sprintf("mode = %s", *sc.mode)
			}
			config := fmt.Sprintf(`
	rule "must_be_true" "sample" {
		condition = false
	}
	
	fix "local_file" "sample" {
		rule_ids = [rule.must_be_true.sample.id]
		paths = ["/file1.txt"]
		content = "Hello world!"
		%s
	}
`, assignment)
			s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{config})
			path := "/file1.txt"
			// Create the file first
			c, err := BuildGreptConfig("", "", nil)
			s.NoError(err)
			p, err := RunGreptPlan(c)
			if sc.expectedErrorMessage != nil {
				s.Contains(err.Error(), *sc.expectedErrorMessage)
				return
			}
			s.NoError(err)
			err = p.Apply()
			s.NoError(err)

			finfo, err := s.fs.Stat(path)
			s.NoError(err)
			s.Equal(sc.wanted, finfo.Mode())
		})
	}
}
