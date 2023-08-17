package queueg

type Option[T any] struct {
	shardsMax    int64
	maxSize      int64
	callback     func(T)
	panicHandler func(e any)
}

func loadOptions[T any](options ...Option[T]) *Option[T] {
	opts := NewOption[T]()
	for _, opt := range options {
		if opt.shardsMax != 0 {
			opts.shardsMax = opt.shardsMax
		} else {
			opts.shardsMax = SHARDSMAX
		}

		if opt.callback != nil {
			opts.callback = opt.callback
		}

		if opt.panicHandler != nil {
			opts.panicHandler = opt.panicHandler
		} else {
			opts.panicHandler = defaultStackTraceHandler
		}

		if opt.maxSize != 0 {
			opts.maxSize = opt.maxSize
		} else {
			opts.maxSize = SHARDSMAX
		}
	}
	return opts
}

func NewOption[T any]() *Option[T] {
	return &Option[T]{}
}

// ShardsMax set shards max
func (o *Option[T]) ShardsMax(shardsMax int64) *Option[T] {
	o.shardsMax = shardsMax
	return o
}

// ShardsMax set shards max
func (o *Option[T]) Size(size int64) *Option[T] {
	o.maxSize = size
	return o
}

func (o *Option[T]) Callback(callback func(T)) *Option[T] {
	o.callback = callback
	return o
}

func (o *Option[T]) PanicHandler(panicHandler func(e any)) *Option[T] {
	o.panicHandler = panicHandler
	return o
}
