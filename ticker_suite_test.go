package ticker_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTicker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ticker Suite")
}
