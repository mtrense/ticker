// +build memory
// +build slow

package memory

import (
	"context"
	"strconv"
	"time"

	es "github.com/mtrense/ticker/eventstream/base"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("memory/event_stream", func() {
	It("handles a large amount of Events with delays", func() {
		totalCount := 100
		s := NewMemoryEventStream(NewMemorySequenceStore())
		s.defaultBufferSize = 10
		w := es.NewWrapper(s)
		go func() {
			for i := 0; i < totalCount; i++ {
				agg := i % 8
				w.Emit(w.Agg("test", strconv.Itoa(agg)))
				time.Sleep(100 * time.Microsecond)
			}
		}()
		time.Sleep(5 * time.Millisecond)
		ctx := context.Background()
		var counter int
		_, _ = w.Stream().Subscribe(ctx, "test", es.Select(), func(e *es.Event) error {
			counter++
			return nil
		})
		Eventually(func() int64 { return w.Stream().LastSequence() }).Should(Equal(int64(totalCount)))
		Eventually(func() int { return counter }).Should(Equal(totalCount))
	})

	It("handles a large amount of Events on slow Subscribers", func() {
		totalCount := 100
		s := NewMemoryEventStream(NewMemorySequenceStore())
		s.defaultBufferSize = 10
		w := es.NewWrapper(s)
		ctx := context.Background()
		var counter int
		_, _ = w.Stream().Subscribe(ctx, "test", es.Select(), func(e *es.Event) error {
			counter++
			time.Sleep(5 * time.Millisecond)
			return nil
		})
		time.Sleep(2 * time.Millisecond)
		go func() {
			for i := 0; i < totalCount; i++ {
				agg := i % 8
				w.Emit(w.Agg("test", strconv.Itoa(agg)))
				time.Sleep(2 * time.Millisecond)
			}
		}()
		Eventually(func() int64 { return w.Stream().LastSequence() }).Should(Equal(int64(totalCount)))
		Eventually(func() int { return counter }).Should(Equal(totalCount))
	})

	It("handles a large amount of Events on multiple slow Subscribers", func() {
		totalCount := 100
		s := NewMemoryEventStream(NewMemorySequenceStore())
		s.defaultBufferSize = 10
		w := es.NewWrapper(s)
		go func() {
			for i := 0; i < totalCount; i++ {
				agg := i % 8
				w.Emit(w.Agg("test", strconv.Itoa(agg)))
			}
		}()
		time.Sleep(2 * time.Millisecond)
		ctx := context.Background()
		var counter1 int
		_, _ = w.Stream().Subscribe(ctx, "test1", es.Select(), func(e *es.Event) error {
			counter1++
			time.Sleep(2 * time.Millisecond)
			return nil
		})
		var counter2 int
		_, _ = w.Stream().Subscribe(ctx, "test2", es.Select(), func(e *es.Event) error {
			counter2++
			time.Sleep(5 * time.Millisecond)
			return nil
		})
		Eventually(func() int64 { return w.Stream().LastSequence() }).Should(Equal(int64(totalCount)))
		Eventually(func() int { return counter1 }).Should(Equal(totalCount))
		Eventually(func() int { return counter2 }).Should(Equal(totalCount))
	})

	It("handles Events on really slow Subscribers", func() {
		totalCount := 20
		s := NewMemoryEventStream(NewMemorySequenceStore())
		s.defaultBufferSize = 10
		w := es.NewWrapper(s)
		go func() {
			for i := 0; i < totalCount; i++ {
				agg := i % 8
				w.Emit(w.Agg("test", strconv.Itoa(agg)))
			}
		}()
		time.Sleep(2 * time.Millisecond)
		ctx := context.Background()
		var counter int
		_, _ = w.Stream().Subscribe(ctx, "test", es.Select(), func(e *es.Event) error {
			counter++
			time.Sleep(100 * time.Millisecond)
			return nil
		})
		Eventually(func() int64 { return w.Stream().LastSequence() }).Should(Equal(int64(totalCount)))
		Eventually(func() int { return counter }, 3*time.Second).Should(Equal(totalCount))
	})
})
