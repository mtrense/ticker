package base

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Base", func() {
	It("Creates a catch-all Selector", func() {
		sel := Select()
		Expect(sel.Type).To(Equal(""))
		Expect(sel.Aggregate).To(Equal([]string{}))
		//Expect(sel.NextSequence).To(Equal(int64(1)))
		//Expect(sel.LastSequence).To(Equal(int64(math.MaxUint64)))
	})
	It("A catch-all Selector matches", func() {
		sel := Select()
		ev1 := &Event{
			Sequence:  3,
			Aggregate: []string{"test"},
			Type:      "create",
		}
		ev2 := &Event{
			Sequence:  6,
			Aggregate: []string{"test2"},
			Type:      "update",
		}
		ev3 := &Event{
			Sequence:  10,
			Aggregate: []string{"test1", "test2"},
			Type:      "delete",
		}
		Expect(sel.Matches(ev1)).To(BeTrue())
		Expect(sel.Matches(ev2)).To(BeTrue())
		Expect(sel.Matches(ev3)).To(BeTrue())
	})
	It("Selector matches the right events", func() {
		ev1 := &Event{
			Sequence:  3,
			Aggregate: []string{"test"},
			Type:      "create",
		}
		ev2 := &Event{
			Sequence:  4,
			Aggregate: []string{"test", "test3"},
			Type:      "create",
		}
		ev3 := &Event{
			Sequence:  6,
			Aggregate: []string{"test2"},
			Type:      "update",
		}
		ev4 := &Event{
			Sequence:  10,
			Aggregate: []string{"test1", "test2"},
			Type:      "delete",
		}
		selAgg1 := Select(SelectAggregate("test"))
		Expect(selAgg1.Matches(ev1)).To(BeTrue())
		Expect(selAgg1.Matches(ev2)).To(BeTrue())
		Expect(selAgg1.Matches(ev3)).To(BeFalse())
		Expect(selAgg1.Matches(ev4)).To(BeFalse())
		selAgg2 := Select(SelectAggregate("test1", "test2"))
		Expect(selAgg2.Matches(ev1)).To(BeFalse())
		Expect(selAgg2.Matches(ev2)).To(BeFalse())
		Expect(selAgg2.Matches(ev3)).To(BeFalse())
		Expect(selAgg2.Matches(ev4)).To(BeTrue())
		selAgg3 := Select(SelectAggregate("test", "test3"))
		Expect(selAgg3.Matches(ev1)).To(BeFalse())
		Expect(selAgg3.Matches(ev2)).To(BeTrue())
		Expect(selAgg3.Matches(ev3)).To(BeFalse())
		Expect(selAgg3.Matches(ev4)).To(BeFalse())
		selType1 := Select(SelectType("create"))
		Expect(selType1.Matches(ev1)).To(BeTrue())
		Expect(selType1.Matches(ev2)).To(BeTrue())
		Expect(selType1.Matches(ev3)).To(BeFalse())
		Expect(selType1.Matches(ev4)).To(BeFalse())
		selAll1 := Select(SelectType("create"), SelectAggregate("test"))
		Expect(selAll1.Matches(ev1)).To(BeTrue())
		Expect(selAll1.Matches(ev2)).To(BeTrue())
		Expect(selAll1.Matches(ev3)).To(BeFalse())
		Expect(selAll1.Matches(ev4)).To(BeFalse())
		selAll2 := Select(SelectType("delete"), SelectAggregate("test"))
		Expect(selAll2.Matches(ev1)).To(BeFalse())
		Expect(selAll2.Matches(ev2)).To(BeFalse())
		Expect(selAll2.Matches(ev3)).To(BeFalse())
		Expect(selAll2.Matches(ev4)).To(BeFalse())
	})
})
