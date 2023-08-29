package manager

type Options struct {
	// Get未找到数据时，返回nil
	NotFoundReturnNil bool
}

func GetOptions(opfs []OptionFunc) *Options {
	ops := &Options{}
	for _, opf := range opfs {
		opf(ops)
	}
	return ops
}

type OptionFunc func(o *Options)

func NotFoundReturnNil(o *Options) {
	o.NotFoundReturnNil = true
}
