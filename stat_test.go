package main

import (
	"math"
	"math/rand"
	"testing"
)

func TestByteSize(t *testing.T) {
	tests := map[float64]string{
		0:    "0B",
		512:  "512B",
		1023: "1023B",

		1 << 10: "1.00 KB",
		1 << 12: "4.00 KB",
		1 << 20: "1.00 MB",
		1 << 23: "8.00 MB",
		1 << 30: "1.00 GB",
		1 << 40: "1.00 TB",
		1 << 41: "2.00 TB",
		1 << 50: "1.00 PB",
		1 << 57: "128.00 PB",
		1 << 60: "1.00 EB",

		(1 << 20) + (1 << 19): "1.50 MB",
		(1 << 30) + (1 << 28): "1.25 GB",
	}

	for i, e := range tests {
		o := byteSize(i)
		if o != e {
			t.Errorf("malformatted bytes: expected %s but got %s", e, o)
		}
	}
}

func TestSum(t *testing.T) {
	for i := 0; i < 100; i++ {
		var sum uint64
		var data []uint64
		for j := 0; j < 50+rand.Intn(50); j++ {
			var x uint64 = rand.Uint64()
			sum += x
			data = append(data, x)
		}
		r := mkStat(data).sum()
		if r != float64(sum) {
			t.Errorf("expected %d but got %f", sum, r)
		}
	}
}

type test struct {
	data   []uint64
	exp    float64
	result bool
}

func avgTester(t *testing.T, tests []test, f func(Statistics) float64) {
	for _, T := range tests {
		lp := len(T.data)
		a := f(mkStat(T.data))
		if (math.Abs(T.exp-a) < 1) != T.result {
			var pre string
			if T.result {
				pre = "expected"
			} else {
				pre = "didn't expect"
			}
			t.Errorf("%s %f but got %f (%v)", pre, T.exp, a, T.data)
		}
		if lp != len(T.data) {
			t.Errorf("Length of dataset (%v) changed. Was %d but now is %d",
				T.data, lp, len(T.data))
		}
	}
}

func TestArithMean(t *testing.T) {
	avgTester(t, []test{
		{[]uint64{1, 2, 3}, 2, true},
		{[]uint64{0, 0, 0, 0, 50}, 10, true},
		{[]uint64{3, 2, 1}, 2, true},              // invariance under exchange
		{[]uint64{1, 1, 1, 1, 1}, 1, true},        // value preservation
		{[]uint64{5, 5, 5, 5, 5, 5}, 5, true},     // first-order preservation
		{[]uint64{10, 20, 30, 40, 50}, 0, false},  // more than min
		{[]uint64{10, 20, 30, 40, 50}, 60, false}, // less than max
		{[]uint64{2}, 1, false},
	}, (Statistics).arithMean)
}

func TestHarmMean(t *testing.T) {
	avgTester(t, []test{
		{[]uint64{1, 4, 4}, 2, true},
		{[]uint64{2, 2, 8, 8, 8, 8}, 4, true},
		{[]uint64{4, 4, 1}, 2, true},              // invariance under exchange
		{[]uint64{1, 1, 1, 1, 1}, 1, true},        // value preservation
		{[]uint64{5, 5, 5, 5, 5, 5}, 5, true},     // first-order preservation
		{[]uint64{10, 20, 30, 40, 50}, 0, false},  // more than min
		{[]uint64{10, 20, 30, 40, 50}, 60, false}, // less than max
		{[]uint64{2}, 1, false},
	}, (Statistics).harmMean)
}

func TestCHarmMean(t *testing.T) {
	avgTester(t, []test{
		{[]uint64{1, 4, 5}, 4, true},
		{[]uint64{2, 2, 8, 8, 8, 8}, 7, true},
		{[]uint64{4, 4, 1}, 3, true},              // invariance under exchange
		{[]uint64{1, 1, 1, 1, 1}, 1, true},        // value preservation
		{[]uint64{5, 5, 5, 5, 5, 5}, 5, true},     // first-order preservation
		{[]uint64{10, 20, 30, 40, 50}, 0, false},  // more than min
		{[]uint64{10, 20, 30, 40, 50}, 60, false}, // less than max
		{[]uint64{2}, 1, false},
	}, (Statistics).contraHarmMean)
}

func TestTurncMean(t *testing.T) {
	avgTester(t, []test{
		{[]uint64{1}, 1, true},
		{[]uint64{1, 3, 7, 10}, 5, true},
		{[]uint64{1, 1, 1, 1, 1}, 1, true},        // value preservation
		{[]uint64{5, 5, 5, 5, 5, 5}, 5, true},     // first-order preservation
		{[]uint64{10, 20, 30, 40, 50}, 0, false},  // more than min
		{[]uint64{10, 20, 30, 40, 50}, 60, false}, // less than max
		{[]uint64{2}, 1, false},
	}, (Statistics).turncMean)
}

func TestWinsMean(t *testing.T) {
	avgTester(t, []test{
		{[]uint64{1}, 1, true},
		{[]uint64{1, 10, 10, 100}, 10, true},
		{[]uint64{1, 1, 1, 1, 1}, 1, true},        // value preservation
		{[]uint64{5, 5, 5, 5, 5, 5}, 5, true},     // first-order preservation
		{[]uint64{10, 20, 30, 40, 50}, 0, false},  // more than min
		{[]uint64{10, 20, 30, 40, 50}, 60, false}, // less than max
		{[]uint64{2}, 1, false},
	}, (Statistics).winsMean)
}

func TestMidrange(t *testing.T) {
	avgTester(t, []test{
		{[]uint64{1}, 1, true},
		{[]uint64{1, 3, 3, 4, 5}, 3, true},
		{[]uint64{1, 10, 10, 99}, 50, true},
		{[]uint64{1, 1, 1, 1, 1}, 1, true},        // value preservation
		{[]uint64{5, 5, 5, 5, 5, 5}, 5, true},     // first-order preservation
		{[]uint64{10, 20, 30, 40, 50}, 0, false},  // more than min
		{[]uint64{10, 20, 30, 40, 50}, 60, false}, // less than max
		{[]uint64{2}, 1, false},
	}, (Statistics).midrange)
}

func TestMidhinge(t *testing.T) {
	avgTester(t, []test{
		{[]uint64{1}, 1, true},
		{[]uint64{1, 3, 4, 5, 4, 5, 6}, 4, true},
		{[]uint64{1, 1, 1, 1, 1}, 1, true},        // value preservation
		{[]uint64{5, 5, 5, 5, 5, 5}, 5, true},     // first-order preservation
		{[]uint64{10, 20, 30, 40, 50}, 0, false},  // more than min
		{[]uint64{10, 20, 30, 40, 50}, 60, false}, // less than max
		{[]uint64{2}, 1, false},
	}, (Statistics).midhinge)
}

func TestTrimean(t *testing.T) {
	avgTester(t, []test{
		{[]uint64{}, 0, true},
		{[]uint64{1}, 1, true},
		{[]uint64{1, 2, 3, 4, 5, 6, 7}, 4, true},
		{[]uint64{1, 2, 3, 15, 1000, 2000, 3000}, 508, true},
		{[]uint64{1, 1, 1, 1, 1}, 1, true},        // value preservation
		{[]uint64{5, 5, 5, 5, 5, 5}, 5, true},     // first-order preservation
		{[]uint64{10, 20, 30, 40, 50}, 0, false},  // more than min
		{[]uint64{10, 20, 30, 40, 50}, 60, false}, // less than max
		{[]uint64{2}, 1, false},
	}, (Statistics).trimean)
}

func TestMedian(t *testing.T) {
	avgTester(t, []test{
		{[]uint64{}, 0, true},
		{[]uint64{1}, 1, true},
		{[]uint64{0, 3, 4, 5, 10}, 4, true},
		{[]uint64{1, 2, 2, 6, 10, 100}, 4, true},
		{[]uint64{5, 5, 5, 5, 5, 5}, 5, true},     // first-order preservation
		{[]uint64{10, 20, 30, 40, 50}, 0, false},  // more than min
		{[]uint64{10, 20, 30, 40, 50}, 60, false}, // less than max
		{[]uint64{2}, 1, false},
	}, (Statistics).median)
}

func TestMode(t *testing.T) {
	avgTester(t, []test{
		{[]uint64{1 << 14}, 1 << 14, true},
		{[]uint64{
			1 << 14,
			1 << 14,
			1 << 14,
			2 * 1 << 14,
			2 * 1 << 14,
			3 * 1 << 14,
		}, 1 << 14, true},
		{[]uint64{
			1 << 14,
			1 << 14,
			1 << 14,
			2 * 1 << 14,
			3 * 1 << 14,
			3 * 1 << 14,
			3 * 1 << 14,
			3 * 1 << 14,
			4 * 1 << 14,
		}, 3 * 1 << 14, true},
		{[]uint64{
			1 << 14,
			10 * 1 << 14,
			10 * 1 << 14,
			100 * 1 << 14,
		}, 10 * 1 << 14, true},
	}, (Statistics).mode)
}
