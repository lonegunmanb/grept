package pkg

type PlanBlock interface {
	ExecuteDuringPlan() error
}
