package pkg

import (
	"fmt"

	"github.com/spf13/afero"
)

var _ Rule = &DirExistRule{}

type DirExistRule struct {
	*BaseBlock
	*BaseRule
	Dir         string `hcl:"dir"`
	FailOnExist bool   `hcl:"fail_on_exist,optional"`
}

func (d *DirExistRule) Type() string {
	return "dir_exist"
}

func (d *DirExistRule) ExecuteDuringPlan() error {
	fs := FsFactory()
	exists, err := afero.Exists(fs, d.Dir)
	if err != nil {
		return err
	}
	if d.FailOnExist && exists {
		d.setCheckError(fmt.Errorf("directory exists: %s", d.Dir))
		return nil
	}
	if !d.FailOnExist && !exists {
		d.setCheckError(fmt.Errorf("directory does not exist: %s", d.Dir))
		return nil
	}
	return nil
}
