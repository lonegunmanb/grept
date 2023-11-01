package pkg

import (
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

var _ Fix = &RenameFile{}

type RenameFile struct {
	*BaseFix
	RuleId  string `json:"rule_id" hcl:"rule_id"`
	OldName string `json:"old_name" hcl:"old_name"`
	NewName string `json:"new_name" hcl:"new_name"`
}

func (rf *RenameFile) Type() string {
	return "rename_file"
}

func (rf *RenameFile) ApplyFix() error {
	fs := FsFactory()
	return fs.Rename(rf.OldName, rf.NewName)
}

func (rf *RenameFile) Eval(b *hclsyntax.Block) error {
	err := rf.BaseFix.Parse(b)
	if err != nil {
		return err
	}
	diag := gohcl.DecodeBody(b.Body, rf.EvalContext(), rf)
	if diag.HasErrors() {
		return diag
	}
	return err
}

func (rf *RenameFile) SetValues(values map[string]cty.Value) {
	values["rule_id"] = ToCtyValue(rf.RuleId)
	values["old_name"] = ToCtyValue(rf.OldName)
	values["new_name"] = ToCtyValue(rf.NewName)
}
