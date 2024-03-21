package pkg

type ApplyBlock interface {
	Block
	Apply() error
}
