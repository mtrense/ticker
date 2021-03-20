// +build memory

package memory

import (
	"context"
	"strconv"

	es "github.com/mtrense/ticker/eventstream/base"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MemoryEventStream", func() {
	It("returns 0 as last sequence for an empty Stream", func() {
		w := es.NewWrapper(New())
		Expect(w.Stream().LastSequence()).To(Equal(int64(0)))
	})

	It("returns the number of emitted Events as last sequence", func() {
		w := es.NewWrapper(New())
		w.Emit()
		Expect(w.Stream().LastSequence()).To(Equal(int64(1)))
		w.Emit()
		Expect(w.Stream().LastSequence()).To(Equal(int64(2)))
	})

	It("streams the EventStream correctly", func() {
		w := es.NewWrapper(New())
		w.Emit()
		w.Emit()
		w.Emit()
		w.Emit()
		Expect(w.Stream().LastSequence()).To(Equal(int64(4)))
		var events []*es.Event
		ctx := context.Background()
		w.Stream().Stream(ctx, es.Select(), es.Range(2, 3), func(e *es.Event) {
			events = append(events, e)
		})
		Expect(len(events)).To(Equal(2))
		Expect(events[0].Sequence).To(Equal(int64(2)))
		Expect(events[1].Sequence).To(Equal(int64(3)))
	})

	It("Subscribe adds a subscription", func() {
		w := es.NewWrapper(New())
		Expect(len(w.Stream().Subscriptions())).To(Equal(0))
		ctx := context.Background()
		_, _ = w.Stream().Subscribe(ctx, "test", es.Select(), func(e *es.Event) {})
		Expect(len(w.Stream().Subscriptions())).To(Equal(1))
	})

	It("Subscription is live when returned", func() {
		s := New()
		w := es.NewWrapper(s)
		Expect(len(w.Stream().Subscriptions())).To(Equal(0))
		ctx := context.Background()
		sub, _ := w.Stream().Subscribe(ctx, "test", es.Select(), func(e *es.Event) {})
		Expect(len(w.Stream().Subscriptions())).To(Equal(1))
		Expect(sub.(*Subscription).live).To(BeTrue())
	})

	It("Subscribe adds an subscription and cancelling the context eventually disables it", func() {
		w := es.NewWrapper(New())
		Expect(len(w.Stream().Subscriptions())).To(Equal(0))
		ctx, cancel := context.WithCancel(context.Background())
		sub, _ := w.Stream().Subscribe(ctx, "test", es.Select(), func(e *es.Event) {})
		Expect(len(w.Stream().Subscriptions())).To(Equal(1))
		cancel()
		Eventually(func() bool { return sub.Active() }).Should(BeFalse())
	})

	It("stores the right sequence for sequentially stored Events", func() {
		w := es.NewWrapper(New())
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Expect(w.Stream().LastSequence()).To(Equal(int64(2)))
		ev1, _ := w.Stream().Get(int64(1))
		Expect(ev1.Sequence).To(Equal(int64(1)))
		ev2, _ := w.Stream().Get(int64(2))
		Expect(ev2.Sequence).To(Equal(int64(2)))
	})

	It("Subscribe gets newly emitted Events", func() {
		w := es.NewWrapper(New())
		var counter int
		ctx := context.Background()
		_, _ = w.Stream().Subscribe(ctx, "test", es.Select(), func(e *es.Event) {
			counter++
		})
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Expect(w.Stream().LastSequence()).To(Equal(int64(2)))
		Eventually(func() int { return counter }).Should(Equal(2))
	})

	It("Subscribe gets selected newly emitted Events", func() {
		w := es.NewWrapper(New())
		var counter int
		ctx := context.Background()
		_, _ = w.Stream().Subscribe(ctx, "test", es.Select(es.SelectAggregate("test", "1")), func(e *es.Event) {
			counter++
		})
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Expect(w.Stream().LastSequence()).To(Equal(int64(2)))
		Eventually(func() int { return counter }).Should(Equal(1))
	})

	It("Subscribe gets historical Events", func() {
		w := es.NewWrapper(New())
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Expect(w.Stream().LastSequence()).To(Equal(int64(2)))
		var counter int
		ctx := context.Background()
		_, _ = w.Stream().Subscribe(ctx, "test", es.Select(), func(e *es.Event) {
			counter++
		})
		Eventually(func() int { return counter }).Should(Equal(2))
	})

	It("Subscribe gets selected historical Events", func() {
		w := es.NewWrapper(New())
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Expect(w.Stream().LastSequence()).To(Equal(int64(2)))
		var counter int
		ctx := context.Background()
		_, _ = w.Stream().Subscribe(ctx, "test", es.Select(es.SelectAggregate("test", "1")), func(e *es.Event) {
			counter++
		})
		Eventually(func() int { return counter }).Should(Equal(1))
	})

	It("Subscribe gets historical and newly emitted Events", func() {
		w := es.NewWrapper(New())
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Expect(w.Stream().LastSequence()).To(Equal(int64(2)))
		var counter int
		ctx := context.Background()
		_, _ = w.Stream().Subscribe(ctx, "test", es.Select(), func(e *es.Event) {
			counter++
		})
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Eventually(func() int { return counter }).Should(Equal(4))
	})

	It("multiple Subscribers get all relevant events", func() {
		w := es.NewWrapper(New())
		w.Emit(w.Type("first"), w.Agg("test", "1"))
		w.Emit(w.Type("first"), w.Agg("test", "2"))
		Expect(w.Stream().LastSequence()).To(Equal(int64(2)))
		var counter1 int
		ctx := context.Background()
		_, _ = w.Stream().Subscribe(ctx, "test1", es.Select(), func(e *es.Event) {
			counter1++
			//fmt.Printf("%s", e.Type)
		})
		var counter2 int
		_, _ = w.Stream().Subscribe(ctx, "test2", es.Select(), func(e *es.Event) {
			counter2++
			//fmt.Println(e.Type)
		})
		w.Emit(w.Type("second"), w.Agg("test", "1"))
		w.Emit(w.Type("second"), w.Agg("test", "2"))
		Eventually(func() int { return counter1 }).Should(Equal(4))
		Eventually(func() int { return counter2 }).Should(Equal(4))
	})

	It("handles a large amount of fast Events", func() {
		s := New()
		s.defaultBufferSize = 10
		w := es.NewWrapper(s)
		for i := 0; i < 50; i++ {
			agg := i % 8
			w.Emit(w.Agg("test", strconv.Itoa(agg)))
		}
		Expect(w.Stream().LastSequence()).To(Equal(int64(50)))
		var counter int
		ctx := context.Background()
		_, _ = w.Stream().Subscribe(ctx, "test", es.Select(), func(e *es.Event) {
			counter++
		})
		for i := 0; i < 50; i++ {
			agg := i % 8
			w.Emit(w.Agg("test", strconv.Itoa(agg)))
		}
		Expect(w.Stream().LastSequence()).To(Equal(int64(100)))
		Eventually(func() int { return counter }).Should(Equal(100))
	})

	It("Subscription properly handles selections", func() {
		s := New()
		s.defaultBufferSize = 10
		w := es.NewWrapper(s)
		for i := 0; i < 20; i++ {
			agg := i % 8
			w.Emit(w.Agg("test", strconv.Itoa(agg)))
		}
		var counter int
		ctx := context.Background()
		_, _ = w.Stream().Subscribe(ctx, "test", es.Select(es.SelectAggregate("test", "1")), func(e *es.Event) {
			counter++
		})
		Expect(w.Stream().LastSequence()).To(Equal(int64(20)))
		Eventually(func() int { return counter }).Should(Equal(3))
	})

})
