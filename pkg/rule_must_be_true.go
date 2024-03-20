package pkg

import (
	"fmt"
)

var _ Rule = &MustBeTrueRule{}

type MustBeTrueRule struct {
	*BaseBlock
	*BaseRule
	Condition    bool   `hcl:"condition"`
	ErrorMessage string `hcl:"error_message,optional"`
}

func (m *MustBeTrueRule) ExecuteDuringPlan() error {
	if !m.Condition {
		m.setCheckError(fmt.Errorf("assertion failed: %s", m.ErrorMessage))
	}
	return nil
}

func (m *MustBeTrueRule) Type() string {
	return "must_be_true"
}
