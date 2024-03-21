package pkg

import "github.com/Azure/grept/golden"

var _ Fix = &RenameFileFix{}

type RenameFileFix struct {
	*golden.BaseBlock
	*BaseFix
	OldName string `json:"old_name" hcl:"old_name"`
	NewName string `json:"new_name" hcl:"new_name"`
}

func (rf *RenameFileFix) Type() string {
	return "rename_file"
}

func (rf *RenameFileFix) Apply() error {
	fs := FsFactory()
	return fs.Rename(rf.OldName, rf.NewName)
}
