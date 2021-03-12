// +build memory

package memory

import (
	"strconv"

	es "github.com/mtrense/ticker/eventstream/base"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MemoryEventStream", func() {
	It("returns 0 as last sequence for an empty Stream", func() {
		w := es.NewWrapper(New())
		Expect(w.Stream().LastSequence()).To(Equal(0))
	})

	It("returns the number of emitted Events as last sequence", func() {
		w := es.NewWrapper(New())
		w.Emit()
		Expect(w.Stream().LastSequence()).To(Equal(1))
		w.Emit()
		Expect(w.Stream().LastSequence()).To(Equal(2))
	})

	It("slices the EventStream correctly", func() {
		w := es.NewWrapper(New())
		w.Emit()
		w.Emit()
		w.Emit()
		w.Emit()
		Expect(w.Stream().LastSequence()).To(Equal(4))
		var events []*es.Event
		w.Stream().Slice(2, 3, func(e *es.Event) bool {
			events = append(events, e)
			return true
		})
		Expect(len(events)).To(Equal(2))
		Expect(events[0].Sequence).To(Equal(2))
		Expect(events[1].Sequence).To(Equal(3))
	})

	It("Subscribe adds a listening channel", func() {
		s := New()
		w := es.NewWrapper(s)
		Expect(len(s.listeners)).To(Equal(0))
		unsub := w.Stream().Subscribe(func(e *es.Event) {}, es.Select())
		Expect(len(s.listeners)).To(Equal(1))
		unsub()
		//time.Sleep(5 * time.Millisecond)
		//w.Emit()
		//Expect(len(s.listeners)).To(Equal(0))
	})

	It("stores the right sequence for sequentially stored Events", func() {
		w := es.NewWrapper(New())
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Expect(w.Stream().LastSequence()).To(Equal(2))
		ev1, _ := w.Stream().Get(1)
		Expect(ev1.Sequence).To(Equal(1))
		ev2, _ := w.Stream().Get(2)
		Expect(ev2.Sequence).To(Equal(2))
	})

	It("Subscribe gets newly emitted Events", func() {
		w := es.NewWrapper(New())
		var counter int
		unsub := w.Stream().Subscribe(func(e *es.Event) {
			counter++
		}, es.Select())
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Expect(w.Stream().LastSequence()).To(Equal(2))
		Eventually(func() int { return counter }).Should(Equal(2))
		unsub()
	})

	It("Subscribe gets selected newly emitted Events", func() {
		w := es.NewWrapper(New())
		var counter int
		unsub := w.Stream().Subscribe(func(e *es.Event) {
			counter++
		}, es.Select(es.SelectAggregate("test", "1")))
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Expect(w.Stream().LastSequence()).To(Equal(2))
		Eventually(func() int { return counter }).Should(Equal(1))
		unsub()
	})

	It("Subscribe gets historical Events", func() {
		w := es.NewWrapper(New())
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Expect(w.Stream().LastSequence()).To(Equal(2))
		var counter int
		unsub := w.Stream().Subscribe(func(e *es.Event) {
			counter++
		}, es.Select())
		Eventually(func() int { return counter }).Should(Equal(2))
		unsub()
	})

	It("Subscribe gets selected historical Events", func() {
		w := es.NewWrapper(New())
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Expect(w.Stream().LastSequence()).To(Equal(2))
		var counter int
		unsub := w.Stream().Subscribe(func(e *es.Event) {
			counter++
		}, es.Select(es.SelectAggregate("test", "1")))
		Eventually(func() int { return counter }).Should(Equal(1))
		unsub()
	})

	It("Subscribe gets historical and newly emitted Events", func() {
		w := es.NewWrapper(New())
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Expect(w.Stream().LastSequence()).To(Equal(2))
		var counter int
		unsub := w.Stream().Subscribe(func(e *es.Event) {
			counter++
		}, es.Select())
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Eventually(func() int { return counter }).Should(Equal(4))
		unsub()
	})

	It("handles a large amount of Events without delays", func() {
		s := New()
		s.defaultBufferSize = 10
		w := es.NewWrapper(s)
		for i := 0; i < 50; i++ {
			agg := i % 8
			w.Emit(w.Agg("test", strconv.Itoa(agg)))
		}
		Expect(w.Stream().LastSequence()).To(Equal(50))
		var counter int
		unsub := w.Stream().Subscribe(func(e *es.Event) {
			counter++
		}, es.Select())
		for i := 0; i < 50; i++ {
			agg := i % 8
			w.Emit(w.Agg("test", strconv.Itoa(agg)))
		}
		Expect(w.Stream().LastSequence()).To(Equal(100))
		Eventually(func() int { return counter }).Should(Equal(100))
		unsub()
	})

})
