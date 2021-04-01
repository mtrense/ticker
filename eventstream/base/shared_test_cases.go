package base

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func EventStreamSampleGroup(factory func() EventStream) {
	It("returns 0 as last sequence for an empty Stream", func() {
		w := NewWrapper(factory())
		Expect(w.Stream().LastSequence()).To(Equal(int64(0)))
	})

	It("returns the number of emitted Events as last sequence", func() {
		w := NewWrapper(factory())
		w.Emit()
		Expect(w.Stream().LastSequence()).To(Equal(int64(1)))
		w.Emit()
		Expect(w.Stream().LastSequence()).To(Equal(int64(2)))
	})

	It("streams the EventStream correctly", func() {
		w := NewWrapper(factory())
		w.Emit()
		w.Emit()
		w.Emit()
		w.Emit()
		Expect(w.Stream().LastSequence()).To(Equal(int64(4)))
		var events []*Event
		ctx := context.Background()
		w.Stream().Stream(ctx, Select(), Range(2, 3), func(e *Event) {
			events = append(events, e)
		})
		Expect(len(events)).To(Equal(2))
		Expect(events[0].Sequence).To(Equal(int64(2)))
		Expect(events[1].Sequence).To(Equal(int64(3)))
	})

	It("Stream respects the EventStream's LastSequence when lower Bracket is too small", func() {
		w := NewWrapper(factory())
		w.Emit()
		w.Emit()
		Expect(w.Stream().LastSequence()).To(Equal(int64(2)))
		var events []*Event
		ctx := context.Background()
		w.Stream().Stream(ctx, Select(), Range(-1, 2), func(e *Event) {
			events = append(events, e)
		})
		Expect(len(events)).To(Equal(2))
		Expect(events[0].Sequence).To(Equal(int64(1)))
		Expect(events[1].Sequence).To(Equal(int64(2)))
	})

	It("Stream respects the EventStream's LastSequence when upper Bracket is too high", func() {
		w := NewWrapper(factory())
		w.Emit()
		w.Emit()
		Expect(w.Stream().LastSequence()).To(Equal(int64(2)))
		var events []*Event
		ctx := context.Background()
		w.Stream().Stream(ctx, Select(), Range(1, 5), func(e *Event) {
			events = append(events, e)
		})
		Expect(len(events)).To(Equal(2))
		Expect(events[0].Sequence).To(Equal(int64(1)))
		Expect(events[1].Sequence).To(Equal(int64(2)))
	})

	It("Subscribe adds a subscription", func() {
		w := NewWrapper(factory())
		Expect(len(w.Stream().Subscriptions())).To(Equal(0))
		ctx := context.Background()
		_, _ = w.Stream().Subscribe(ctx, "test", Select(), func(e *Event) {})
		Expect(len(w.Stream().Subscriptions())).To(Equal(1))
	})

	It("Subscribe adds an subscription and cancelling the context eventually disables it", func() {
		w := NewWrapper(factory())
		Expect(len(w.Stream().Subscriptions())).To(Equal(0))
		ctx, cancel := context.WithCancel(context.Background())
		sub, _ := w.Stream().Subscribe(ctx, "test", Select(), func(e *Event) {})
		Expect(len(w.Stream().Subscriptions())).To(Equal(1))
		cancel()
		Eventually(func() bool { return sub.Active() }).Should(BeFalse())
	})

	It("stores the right sequence for sequentially stored Events", func() {
		w := NewWrapper(factory())
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Expect(w.Stream().LastSequence()).To(Equal(int64(2)))
		ev1, _ := w.Stream().Get(int64(1))
		Expect(ev1.Sequence).To(Equal(int64(1)))
		ev2, _ := w.Stream().Get(int64(2))
		Expect(ev2.Sequence).To(Equal(int64(2)))
	})

	It("Subscribe gets newly emitted Events", func() {
		w := NewWrapper(factory())
		var counter int
		ctx := context.Background()
		_, _ = w.Stream().Subscribe(ctx, "test", Select(), func(e *Event) {
			counter++
		})
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Expect(w.Stream().LastSequence()).To(Equal(int64(2)))
		Eventually(func() int { return counter }).Should(Equal(2))
	})

	It("Subscribe gets selected newly emitted Events", func() {
		w := NewWrapper(factory())
		var counter int
		ctx := context.Background()
		_, _ = w.Stream().Subscribe(ctx, "test", Select(SelectAggregate("test", "1")), func(e *Event) {
			counter++
		})
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Expect(w.Stream().LastSequence()).To(Equal(int64(2)))
		Eventually(func() int { return counter }).Should(Equal(1))
	})

	It("Subscribe gets historical Events", func() {
		w := NewWrapper(factory())
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Expect(w.Stream().LastSequence()).To(Equal(int64(2)))
		var counter int
		ctx := context.Background()
		_, _ = w.Stream().Subscribe(ctx, "test", Select(), func(e *Event) {
			counter++
		})
		Eventually(func() int { return counter }).Should(Equal(2))
	})

	It("Subscribe gets selected historical Events", func() {
		w := NewWrapper(factory())
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Expect(w.Stream().LastSequence()).To(Equal(int64(2)))
		var counter int
		ctx := context.Background()
		_, _ = w.Stream().Subscribe(ctx, "test", Select(SelectAggregate("test", "1")), func(e *Event) {
			counter++
		})
		Eventually(func() int { return counter }).Should(Equal(1))
	})

	It("Subscribe gets historical and newly emitted Events", func() {
		w := NewWrapper(factory())
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Expect(w.Stream().LastSequence()).To(Equal(int64(2)))
		var counter int
		ctx := context.Background()
		_, _ = w.Stream().Subscribe(ctx, "test", Select(), func(e *Event) {
			counter++
		})
		w.Emit(w.Agg("test", "1"))
		w.Emit(w.Agg("test", "2"))
		Eventually(func() int { return counter }).Should(Equal(4))
	})

	It("multiple Subscribers get all relevant events", func() {
		w := NewWrapper(factory())
		w.Emit(w.Type("first"), w.Agg("test", "1"))
		w.Emit(w.Type("first"), w.Agg("test", "2"))
		Expect(w.Stream().LastSequence()).To(Equal(int64(2)))
		var counter1 int
		ctx := context.Background()
		_, _ = w.Stream().Subscribe(ctx, "test1", Select(), func(e *Event) {
			counter1++
		})
		var counter2 int
		_, _ = w.Stream().Subscribe(ctx, "test2", Select(), func(e *Event) {
			counter2++
		})
		w.Emit(w.Type("second"), w.Agg("test", "1"))
		w.Emit(w.Type("second"), w.Agg("test", "2"))
		Eventually(func() int { return counter1 }).Should(Equal(4))
		Eventually(func() int { return counter2 }).Should(Equal(4))
	})

	It("Subscription handles outage", func() {
		w := NewWrapper(factory())
		w.Emit(w.Type("first"), w.Agg("test", "1"))
		w.Emit(w.Type("first"), w.Agg("test", "2"))
		var counter int
		ctx, cancel := context.WithCancel(context.Background())
		sub, _ := w.Stream().Subscribe(ctx, "test", Select(), func(e *Event) {
			counter++
		})
		Eventually(func() int { return counter }).Should(Equal(2))
		cancel()
		Eventually(func() bool { return sub.Active() }).Should(BeFalse())
		w.Emit(w.Type("second"), w.Agg("test", "1"))
		w.Emit(w.Type("second"), w.Agg("test", "2"))
		ctx, cancel = context.WithCancel(context.Background())
		sub, _ = w.Stream().Subscribe(ctx, "test", Select(), func(e *Event) {
			counter++
		})
		Eventually(func() int { return counter }).Should(Equal(4))
	})

}
