package pkg

type Rule interface {
	PlanBlock
	CheckError() error
	// discriminator func
	Rule()
	setCheckError(error)
}

type BaseRule struct {
	checkErr error
}

func (br *BaseRule) BlockType() string {
	return "rule"
}

func (br *BaseRule) CheckError() error {
	return br.checkErr
}

func (br *BaseRule) Rule() {}

func (br *BaseRule) AddressLength() int { return 3 }

func (br *BaseRule) setCheckError(err error) {
	br.checkErr = err
}
