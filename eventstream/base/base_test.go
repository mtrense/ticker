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
		Expect(sel.NextSequence).To(Equal(int64(1)))
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
})
