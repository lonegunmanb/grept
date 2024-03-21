package pkg

import (
	"github.com/Azure/grept/golden"
	"github.com/hashicorp/go-multierror"
)

var _ Fix = &RmLocalFileFix{}

type RmLocalFileFix struct {
	*golden.BaseBlock
	*BaseFix
	Paths []string `hcl:"paths" json:"paths"`
}

func (r *RmLocalFileFix) Type() string {
	return "rm_local_file"
}

func (r *RmLocalFileFix) Apply() error {
	fs := FsFactory()
	var err error
	for _, path := range r.Paths {
		removeErr := fs.RemoveAll(path)
		if removeErr != nil {
			err = multierror.Append(err, removeErr)
		}
	}
	return err
}
