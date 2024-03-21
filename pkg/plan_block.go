package pkg

type PlanBlock interface {
	Block
	ExecuteDuringPlan() error
}
