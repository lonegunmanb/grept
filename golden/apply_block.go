package golden

type ApplyBlock interface {
	Block
	Apply() error
}
