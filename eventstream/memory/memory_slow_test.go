// +build memory
// +build slow

package memory

import (
	"strconv"
	"time"

	es "github.com/mtrense/ticker/eventstream/base"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MemoryEventStream", func() {
	It("handles a large amount of Events with delays", func() {
		totalCount := 100
		s := New()
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
		var counter int
		unsub := w.Stream().Subscribe(func(e *es.Event) {
			counter++
		}, es.Select())
		Eventually(func() int { return w.Stream().LastSequence() }).Should(Equal(totalCount))
		Eventually(func() int { return counter }).Should(Equal(totalCount))
		unsub()
	})

	It("handles a large amount of Events on slow Subscribers", func() {
		totalCount := 100
		s := New()
		s.defaultBufferSize = 10
		w := es.NewWrapper(s)
		go func() {
			for i := 0; i < totalCount; i++ {
				agg := i % 8
				w.Emit(w.Agg("test", strconv.Itoa(agg)))
			}
		}()
		time.Sleep(2 * time.Millisecond)
		var counter int
		unsub := w.Stream().Subscribe(func(e *es.Event) {
			counter++
			time.Sleep(2 * time.Millisecond)
		}, es.Select())
		Eventually(func() int { return w.Stream().LastSequence() }).Should(Equal(totalCount))
		Eventually(func() int { return counter }).Should(Equal(totalCount))
		unsub()
	})

	It("handles Events on really slow Subscribers", func() {
		totalCount := 20
		s := New()
		s.defaultBufferSize = 10
		w := es.NewWrapper(s)
		go func() {
			for i := 0; i < totalCount; i++ {
				agg := i % 8
				w.Emit(w.Agg("test", strconv.Itoa(agg)))
			}
		}()
		time.Sleep(2 * time.Millisecond)
		var counter int
		unsub := w.Stream().Subscribe(func(e *es.Event) {
			counter++
			time.Sleep(100 * time.Millisecond)
		}, es.Select())
		Eventually(func() int { return w.Stream().LastSequence() }).Should(Equal(totalCount))
		Eventually(func() int { return counter }, 3*time.Second).Should(Equal(totalCount))
		unsub()
	})
})
