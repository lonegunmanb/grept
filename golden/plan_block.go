package golden

type PlanBlock interface {
	Block
	ExecuteDuringPlan() error
}
