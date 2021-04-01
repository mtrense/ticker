// +build memory

package memory

func DefaultBufferSize(size int) Option {
	return func(s *EventStream) {
		s.defaultBufferSize = size
	}
}
