package base

import "math"

type Bracket struct {
	NextSequence int64
	LastSequence int64
}

func All() Bracket {
	return Bracket{
		NextSequence: 1,
		LastSequence: math.MaxInt64,
	}
}

func Range(next, last int64) Bracket {
	return Bracket{
		NextSequence: next,
		LastSequence: last,
	}
}

func From(next int64) Bracket {
	return Bracket{
		NextSequence: next,
		LastSequence: math.MaxInt64,
	}
}
