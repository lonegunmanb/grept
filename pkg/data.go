package pkg

import "github.com/Azure/grept/golden"

type Data interface {
	golden.PlanBlock
	// discriminator func
	Data()
}

type BaseData struct{}

func (bd *BaseData) BlockType() string {
	return "data"
}

func (bd *BaseData) Data() {}

func (bd *BaseData) AddressLength() int { return 3 }
