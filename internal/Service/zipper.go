package Service

type ZipperInterface interface {
}

type Zipper struct {
}

func NewZipper() *Zipper {
	return &Zipper{}
}
