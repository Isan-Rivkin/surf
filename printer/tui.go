package printer

type TuiController[L Loader, T Table] interface {
	GetLoader() Loader
	GetTable() Table
}

type Printer[L Loader, T Table] struct {
	loader L
	table  T
}

func NewPrinter[L Loader, T Table](l Loader, t Table) TuiController[Loader, Table] {
	return &Printer[Loader, Table]{
		loader: l,
		table:  t,
	}
}

func (p *Printer[L, T]) GetLoader() L {
	return p.loader
}

func (p *Printer[L, T]) GetTable() T {
	return p.table
}
