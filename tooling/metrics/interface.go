package metrics

type Span interface {
	Done(err error)
}

type Obs interface {
	StartSpan(name string) (span Span)
}
