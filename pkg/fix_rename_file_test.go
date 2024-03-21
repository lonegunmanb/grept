package pkg

import (
	"context"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type renameFileFixSuite struct {
	suite.Suite
	*testBase
}

func TestRenameFixSuite(t *testing.T) {
	suite.Run(t, new(renameFileFixSuite))
}

func (s *renameFileFixSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *renameFileFixSuite) TearDownTest() {
	s.teardown()
}

func (s *renameFileFixSuite) TestRenameFile_ApplyFix() {
	fs := s.fs
	t := s.T()
	// Create temporary file for testing
	tmpFile, err := afero.TempFile(fs, "", "test")
	assert.NoError(t, err)
	oldName := tmpFile.Name()
	newName := oldName + "_renamed"

	// Initialize RenameFileFix fix
	rf := &RenameFileFix{
		BaseBlock: &BaseBlock{
			c: &GreptConfig{BaseConfig: NewBasicConfig(".", context.TODO())},
		},
		OldName: oldName,
		NewName: newName,
	}

	// Apply the fix
	err = rf.Apply()

	// Assert there is no error and the renamed file exists
	assert.NoError(t, err)
	exists, err := afero.Exists(fs, newName)
	require.NoError(t, err)
	assert.True(t, exists)
	exists, err = afero.Exists(fs, oldName)
	require.NoError(t, err)
	assert.False(t, exists)
}
