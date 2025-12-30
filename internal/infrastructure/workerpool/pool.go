package workerpool

type Pool interface {
	Submit(func()) error
	Release()
}
