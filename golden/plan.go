package golden

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
)

type Plan interface {
	String() string
	Apply() error
}

func RunPlan(b Block) error {
	decodeErr := Decode(b)
	if decodeErr != nil {
		return fmt.Errorf("%s(%s) Decode error: %+v", b.Address(), b.HclBlock().Range().String(), decodeErr)
	}
	if validateErr := Validate.Struct(b); validateErr != nil {
		return fmt.Errorf("%s.%s.%s is not valid: %s", b.BlockType(), b.Type(), b.Name(), validateErr.Error())
	}
	failedChecks, preConditionCheckError := b.PreConditionCheck(b.EvalContext())
	if preConditionCheckError != nil {
		return preConditionCheckError
	}
	if len(failedChecks) > 0 {
		var err error
		for _, c := range failedChecks {
			err = multierror.Append(err, fmt.Errorf("precondition check error: %s, %s", c.ErrorMessage, c.Body.Range().String()))
		}
		return err
	}
	pa, ok := b.(PlanBlock)
	if ok {
		execErr := pa.ExecuteDuringPlan()
		if execErr != nil {
			return fmt.Errorf("%s.%s.%s(%s) exec error: %+v", b.Type(), b.Type(), b.Name(), b.HclBlock().Range().String(), execErr)
		}
	}
	return nil
}
