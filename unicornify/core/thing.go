package core

type Thing interface {
	GetTracer(wv WorldView) Tracer
}
