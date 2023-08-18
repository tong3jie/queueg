package queueg

type Option[T any] struct {
	ShardsMax    int64
	MaxSize      int64
	Callback     func(T)
	PanicHandler func(e any)
}

func loadOptions[T any](options ...*Option[T]) *Option[T] {
	opts := NewOption[T]()
	for _, opt := range options {
		if opt.ShardsMax != 0 {
			opts.ShardsMax = opt.ShardsMax
		} else {
			opts.ShardsMax = SHARDSMAX
		}

		if opt.Callback != nil {
			opts.Callback = opt.Callback
		}

		if opt.PanicHandler != nil {
			opts.PanicHandler = opt.PanicHandler
		} else {
			opts.PanicHandler = defaultStackTraceHandler
		}

		if opt.MaxSize != 0 {
			opts.MaxSize = opt.MaxSize
		} else {
			opts.MaxSize = SHARDSMAX
		}
	}
	return opts
}

func NewOption[T any]() *Option[T] {
	return &Option[T]{
		ShardsMax:    SHARDSMAX,
		MaxSize:      SHARDSMAX * 10,
		Callback:     nil,
		PanicHandler: defaultStackTraceHandler,
	}
}

// ShardsMax set shards max
func WithShardsMax[T any](shardsMax int64) *Option[T] {
	o := &Option[T]{
		ShardsMax: shardsMax,
	}
	return o
}

// ShardsMax set shards max
func WithSize[T any](size int64) *Option[T] {
	o := &Option[T]{
		MaxSize: size,
	}
	return o
}

func WithCallback[T any](callback func(T)) *Option[T] {
	o := &Option[T]{}
	o.Callback = callback
	return o
}

func WithPanicHandler[T any](panicHandler func(e any)) *Option[T] {
	o := &Option[T]{}
	o.PanicHandler = panicHandler
	return o
}
