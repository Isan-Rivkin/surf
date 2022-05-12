package printer

type TuiController[L Loader] interface {
	GetLoader() Loader
}

type Printer[L Loader] struct {
	loader L
}

func NewPrinter[L Loader](l Loader) TuiController[Loader] {
	return &Printer[Loader]{
		loader: l,
	}
}

func (p *Printer[L]) GetLoader() L {
	return p.loader
}
