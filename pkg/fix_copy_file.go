package pkg

import (
	"fmt"
	"github.com/Azure/grept/golden"
	"github.com/spf13/afero"
)

var _ Fix = &CopyFileFix{}

type CopyFileFix struct {
	*golden.BaseBlock
	*BaseFix
	Src  string `json:"src" hcl:"src"`
	Dest string `json:"dest" hcl:"dest"`
}

func (c *CopyFileFix) Type() string {
	return "copy_file"
}

func (c *CopyFileFix) Apply() error {
	fs := FsFactory()
	file, err := fs.Open(c.Src)
	if err != nil {
		return fmt.Errorf("error on reading src %s %+v", c.Src, err)
	}
	defer func() {
		_ = file.Close()
	}()
	err = afero.WriteReader(fs, c.Dest, file)
	if err != nil {
		return fmt.Errorf("error on writing dest %s %+v", c.Dest, err)
	}
	return nil
}
