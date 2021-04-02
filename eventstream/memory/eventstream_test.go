// +build memory

package memory

import (
	"context"
	"strconv"

	es "github.com/mtrense/ticker/eventstream/base"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("memory/eventstream", func() {
	es.EventStreamSampleGroup(func() es.EventStream {
		return NewMemoryEventStream(NewMemorySequenceStore())
	})

	It("Subscription is live when returned", func() {
		w := es.NewWrapper(NewMemoryEventStream(NewMemorySequenceStore()))
		Expect(len(w.Stream().Subscriptions())).To(Equal(0))
		ctx := context.Background()
		sub, _ := w.Stream().Subscribe(ctx, "test", es.Select(), func(e *es.Event) error { return nil })
		Expect(len(w.Stream().Subscriptions())).To(Equal(1))
		Expect(sub.(*Subscription).live).To(BeTrue())
	})

	It("handles a large amount of fast Events", func() {
		s := NewMemoryEventStream(NewMemorySequenceStore())
		s.defaultBufferSize = 10
		w := es.NewWrapper(s)
		for i := 0; i < 50; i++ {
			agg := i % 8
			w.Emit(w.Agg("test", strconv.Itoa(agg)))
		}
		Expect(w.Stream().LastSequence()).To(Equal(int64(50)))
		var counter int
		ctx := context.Background()
		_, _ = w.Stream().Subscribe(ctx, "test", es.Select(), func(e *es.Event) error {
			counter++
			return nil
		})
		for i := 0; i < 50; i++ {
			agg := i % 8
			w.Emit(w.Agg("test", strconv.Itoa(agg)))
		}
		Expect(w.Stream().LastSequence()).To(Equal(int64(100)))
		Eventually(func() int { return counter }).Should(Equal(100))
	})

	It("Subscription properly handles selections", func() {
		s := NewMemoryEventStream(NewMemorySequenceStore())
		s.defaultBufferSize = 10
		w := es.NewWrapper(s)
		for i := 0; i < 20; i++ {
			agg := i % 8
			w.Emit(w.Agg("test", strconv.Itoa(agg)))
		}
		var counter int
		ctx := context.Background()
		_, _ = w.Stream().Subscribe(ctx, "test", es.Select(es.SelectAggregate("test", "1")), func(e *es.Event) error {
			counter++
			return nil
		})
		Expect(w.Stream().LastSequence()).To(Equal(int64(20)))
		Eventually(func() int { return counter }).Should(Equal(3))
	})

})
