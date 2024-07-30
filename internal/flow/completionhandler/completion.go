package completionhandler

type CompletionHandler struct {
	err      error
	parties  int
	resolved int
	done     bool
	running  bool
	resolve  func()
	reject   func(error)
}

func New(resolve func(), reject func(error)) *CompletionHandler {
	return &CompletionHandler{
		parties:  0,
		resolved: 0,
		resolve:  resolve,
		reject:   reject,
		running:  false,
		done:     false,
	}
}

func (h *CompletionHandler) Register() (func(), func(err error)) {
	h.parties = h.parties + 1
	return func() {
			h.resolved = h.resolved + 1
			h.handle()
		}, func(err error) {
			h.err = err
			h.handle()
		}
}

func (h *CompletionHandler) handle() {
	if h.running && !h.done {
		if h.err != nil {
			h.done = true
			h.reject(h.err)
		} else if h.resolved == h.parties {
			h.done = true
			h.resolve()
		}
	}
}

func (h *CompletionHandler) Execute() {
	h.running = true
	if !h.done {
		h.handle()
	}
}
