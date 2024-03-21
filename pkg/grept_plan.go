package pkg

import (
	"fmt"
	"github.com/Azure/grept/golden"
	"github.com/ahmetb/go-linq/v3"
	"github.com/emirpasic/gods/queues/linkedlistqueue"
	"github.com/hashicorp/go-multierror"
	"strings"
	"sync"
)

func RunGreptPlan(c golden.Config) (*GreptPlan, error) {
	err := c.RunDag(plan)
	if err != nil {
		return nil, err
	}

	plan := newPlan()
	for _, rb := range golden.Blocks[Rule](c) {
		checkErr := rb.CheckError()
		if checkErr == nil {
			continue
		}
		plan.addRule(&FailedRule{
			Rule:       rb,
			CheckError: checkErr,
		})
		for _, fb := range golden.Blocks[Fix](c) {
			if linq.From(fb.GetRuleIds()).Contains(rb.Id()) {
				plan.addFix(fb)
			}
		}
	}

	return plan, nil
}

var _ golden.Plan = &GreptPlan{}

type GreptPlan struct {
	FailedRules []*FailedRule
	Fixes       map[string]Fix
	mu          sync.Mutex
}

func newPlan() *GreptPlan {
	return &GreptPlan{
		Fixes: make(map[string]Fix),
	}
}

func (p *GreptPlan) String() string {
	sb := strings.Builder{}
	for _, r := range p.FailedRules {
		sb.WriteString(r.String())
		sb.WriteString("\n---\n")
	}
	for _, f := range p.Fixes {
		sb.WriteString(fmt.Sprintf("%s would be apply:\n %s\n", f.Address(), golden.BlockToString(f)))
		sb.WriteString("\n---\n")
	}

	return sb.String()
}

func (p *GreptPlan) Apply() error {
	var err error
	for _, fix := range p.Fixes {
		if err = golden.Decode(fix); err != nil {
			err = multierror.Append(err, fmt.Errorf("rule.%s.%s(%s) decode error: %+v", fix.Type(), fix.Name(), fix.HclBlock().Range().String(), err))
		}
		if err != nil {
			return err
		}
	}

	for _, fix := range p.Fixes {
		if applyErr := fix.Apply(); applyErr != nil {
			err = multierror.Append(err, applyErr)
		}
	}
	if err != nil {
		return fmt.Errorf("errors applying fixes: %+v", err)
	}
	return nil
}

func (p *GreptPlan) addRule(fr *FailedRule) {
	p.mu.Lock()
	p.FailedRules = append(p.FailedRules, fr)
	p.mu.Unlock()
}

func (p *GreptPlan) addFix(f Fix) {
	p.mu.Lock()
	p.Fixes[f.Id()] = f
	p.mu.Unlock()
}

func plan(c golden.Config, dag *golden.Dag, q *linkedlistqueue.Queue, b golden.Block) error {
	self, _ := dag.GetVertex(b.Address())
	return golden.RunPlan(self.(golden.Block))
}

type FailedRule struct {
	Rule
	CheckError error
}

func (fr *FailedRule) String() string {
	address := fr.Address()
	return fmt.Sprintf("%s check return failure: %s", address, fr.CheckError.Error())
}
