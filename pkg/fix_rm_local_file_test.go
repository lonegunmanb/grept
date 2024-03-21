package pkg

import (
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"testing"
)

type fixRmLocalFileSuite struct {
	suite.Suite
	*testBase
}

func TestFixRmLocalFileSuite(t *testing.T) {
	suite.Run(t, new(fixRmLocalFileSuite))
}

func (s *fixRmLocalFileSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *fixRmLocalFileSuite) TearDownTest() {
	s.teardown()
}

func (s *fixRmLocalFileSuite) TestRemoveFile_FileExist() {
	fs := s.fs
	t := s.T()
	tmpFile, err := afero.TempFile(fs, "", "test")
	require.NoError(t, err)
	fileName := tmpFile.Name()
	rf := &RmLocalFileFix{
		Paths: []string{fileName},
	}

	err = rf.Apply()

	require.NoError(t, err)
	exists, err := afero.Exists(fs, fileName)
	require.NoError(t, err)
	assert.False(t, exists)
}

func (s *fixRmLocalFileSuite) TestRemoveFile_FileNotExist() {
	fileName := "/path/to/not-exist-file"
	rf := &RmLocalFileFix{
		Paths: []string{fileName},
	}

	err := rf.Apply()

	s.NoError(err)
}

func TestRemoveFile_RemoveFolder(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_grept")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()
	err = os.WriteFile(filepath.Join(tmpDir, "test"), []byte("hello"), 0600)
	require.NoError(t, err)
	rf := &RmLocalFileFix{
		Paths: []string{tmpDir},
	}

	err = rf.Apply()

	require.NoError(t, err)
	exists, err := afero.DirExists(FsFactory(), tmpDir)
	require.NoError(t, err)
	assert.False(t, exists)
}
