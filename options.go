package workerpool

type Options struct {
	workers        int
	jobBufferSize  int
	disableResults bool
}

type Option func(options *Options)

func SetWorkers(workers int) Option {
	return func(options *Options) {
		options.workers = workers
	}
}

func SetJobsBuffer(size int) Option {
	return func(options *Options) {
		options.jobBufferSize = size
	}
}

func DisableResults() Option {
	return func(options *Options) {
		options.disableResults = true
	}
}
