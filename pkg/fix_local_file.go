package pkg

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/afero"
)

var _ Fix = &LocalFileFix{}

type LocalFileFix struct {
	*BaseBlock
	*BaseFix
	Paths   []string     `json:"paths" hcl:"paths"`
	Content string       `json:"content" hcl:"content"`
	Mode    *fs.FileMode `json:"mode" hcl:"mode,optional" default:"0644" validate:"file_mode"`
}

func (lf *LocalFileFix) Type() string {
	return "local_file"
}

func (lf *LocalFileFix) Apply() error {
	fm, err := toDecimal(*lf.Mode)
	if err != nil {
		return err
	}

	fs := FsFactory()
	for _, path := range lf.Paths {
		dir := filepath.Dir(path)
		dirExists, dirCheckErr := afero.DirExists(fs, dir)
		if dirCheckErr != nil {
			err = multierror.Append(err, dirCheckErr)
			continue
		}
		if !dirExists {
			mkDirErr := fs.MkdirAll(dir, 0755)
			if mkDirErr != nil {
				err = multierror.Append(err, mkDirErr)
				continue
			}
		}
		writeErr := afero.WriteFile(fs, path, []byte(lf.Content), fm)
		if writeErr != nil {
			err = multierror.Append(err, writeErr)
		}
	}

	return err
}

func toDecimal(octalMode fs.FileMode) (fs.FileMode, error) {
	mode, err := strconv.ParseUint(strconv.Itoa(int(octalMode)), 8, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid file mode: %w", err)
	}
	fm := fs.FileMode(mode)
	return fm, nil
}
