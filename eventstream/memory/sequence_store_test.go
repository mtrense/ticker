// +build memory

package memory

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("memory/sequence_store", func() {
	It("Returns zero for an unknown ClientID", func() {
		s := NewMemorySequenceStore()
		Expect(s.Get("test")).To(Equal(int64(0)))
	})
	It("Returns a value previously Stored", func() {
		s := NewMemorySequenceStore()
		s.Store("test", 6)
		Expect(s.Get("test")).To(Equal(int64(6)))
	})
})
