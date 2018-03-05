package main

import (
	"math/rand"
	"testing"
)

func TestByteSize(t *testing.T) {
	tests := map[uint64]string{
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
	for i := 0; i < 1000; i++ {
		var sum uint64
		var data sizes
		for j := 0; j < 100; j++ {
			var x uint64 = rand.Uint64()
			sum += x
			data = append(data, x)
		}
		r := data.sum()
		if r != sum {
			t.Errorf("expected %d but got %d", sum, r)
		}
	}
}

type test struct {
	data   sizes
	exp    uint64
	result bool
}

func avgTester(t *testing.T, tests []test, f func(sizes) uint64) {
	for _, T := range tests {
		a := f(T.data)
		if (T.exp == a) != T.result {
			var pre string
			if T.result {
				pre = "expected"
			} else {
				pre = "didn't expect"
			}
			t.Errorf("%s %d but got %d", pre, T.exp, a)
		}
	}
}

func TestArithMean(t *testing.T) {
	avgTester(t, []test{
		{sizes{1, 2, 3}, 2, true},
		{sizes{0, 0, 0, 0, 50}, 10, true},
		{sizes{3, 2, 1}, 2, true},              // invariance under exchange
		{sizes{1, 1, 1, 1, 1}, 1, true},        // value preservation
		{sizes{5, 5, 5, 5, 5, 5}, 5, true},     // first-order preservation
		{sizes{10, 20, 30, 40, 50}, 0, false},  // more than min
		{sizes{10, 20, 30, 40, 50}, 60, false}, // less than max
		{sizes{2}, 1, false},
	}, (sizes).arithMean)
}

func TestGeomMean(t *testing.T) {
	avgTester(t, []test{
		{sizes{1, 2, 4}, 2, true},
		{sizes{0, 0, 0, 0, 50}, 0, true},
		{sizes{1, 3, 9, 27, 81}, 9, true},
		{sizes{2, 4, 1}, 2, true},              // invariance under exchange
		{sizes{1, 1, 1, 1, 1}, 1, true},        // value preservation
		{sizes{5, 5, 5, 5, 5, 5}, 5, true},     // first-order preservation
		{sizes{10, 20, 30, 40, 50}, 0, false},  // more than min
		{sizes{10, 20, 30, 40, 50}, 60, false}, // less than max
		{sizes{2}, 1, false},
	}, (sizes).geomMean)
}

func TestHarmMean(t *testing.T) {
	avgTester(t, []test{
		{sizes{1, 4, 4}, 2, true},
		{sizes{2, 2, 8, 8, 8, 8}, 4, true},
		{sizes{4, 4, 1}, 2, true},              // invariance under exchange
		{sizes{1, 1, 1, 1, 1}, 1, true},        // value preservation
		{sizes{5, 5, 5, 5, 5, 5}, 5, true},     // first-order preservation
		{sizes{10, 20, 30, 40, 50}, 0, false},  // more than min
		{sizes{10, 20, 30, 40, 50}, 60, false}, // less than max
		{sizes{2}, 1, false},
	}, (sizes).harmMean)
}

func TestCHarmMean(t *testing.T) {
	avgTester(t, []test{
		{sizes{1, 4, 5}, 2, true},
		{sizes{2, 2, 8, 8, 8, 8}, 4, true},
		{sizes{4, 4, 1}, 2, true},              // invariance under exchange
		{sizes{1, 1, 1, 1, 1}, 1, true},        // value preservation
		{sizes{5, 5, 5, 5, 5, 5}, 5, true},     // first-order preservation
		{sizes{10, 20, 30, 40, 50}, 0, false},  // more than min
		{sizes{10, 20, 30, 40, 50}, 60, false}, // less than max
		{sizes{2}, 1, false},
	}, (sizes).harmMean)
}

func TestTurncMean(t *testing.T) {
	avgTester(t, []test{
		{sizes{1}, 1, true},
		{sizes{1, 3, 7, 10}, 5, true},
		{sizes{3, 1, 10, 7}, 5, true},          // invariance under exchange
		{sizes{1, 1, 1, 1, 1}, 1, true},        // value preservation
		{sizes{5, 5, 5, 5, 5, 5}, 5, true},     // first-order preservation
		{sizes{10, 20, 30, 40, 50}, 0, false},  // more than min
		{sizes{10, 20, 30, 40, 50}, 60, false}, // less than max
		{sizes{2}, 1, false},
	}, (sizes).turncMean)
}

func TestWinsMean(t *testing.T) {
	avgTester(t, []test{
		{sizes{1}, 1, true},
		{sizes{1, 10, 10, 100}, 10, true},
		{sizes{10, 1, 100, 10}, 10, true},      // invariance under exchange
		{sizes{1, 1, 1, 1, 1}, 1, true},        // value preservation
		{sizes{5, 5, 5, 5, 5, 5}, 5, true},     // first-order preservation
		{sizes{10, 20, 30, 40, 50}, 0, false},  // more than min
		{sizes{10, 20, 30, 40, 50}, 60, false}, // less than max
		{sizes{2}, 1, false},
	}, (sizes).winsMean)
}

func TestMidrange(t *testing.T) {
	avgTester(t, []test{
		{sizes{1}, 1, true},
		{sizes{1, 3, 3, 4, 5}, 3, true},
		{sizes{1, 10, 10, 99}, 50, true},
		{sizes{99, 1, 10, 10}, 10, true},       // invariance under exchange
		{sizes{1, 1, 1, 1, 1}, 1, true},        // value preservation
		{sizes{5, 5, 5, 5, 5, 5}, 5, true},     // first-order preservation
		{sizes{10, 20, 30, 40, 50}, 0, false},  // more than min
		{sizes{10, 20, 30, 40, 50}, 60, false}, // less than max
		{sizes{2}, 1, false},
	}, (sizes).winsMean)
}

func TestMidhinge(t *testing.T) {
	avgTester(t, []test{
		{sizes{1}, 1, true},
		{sizes{1, 3, 4, 5, 4, 5, 6}, 4, true},
		{sizes{10, 1, 100, 10}, 10, true},      // invariance under exchange
		{sizes{1, 1, 1, 1, 1}, 1, true},        // value preservation
		{sizes{5, 5, 5, 5, 5, 5}, 5, true},     // first-order preservation
		{sizes{10, 20, 30, 40, 50}, 0, false},  // more than min
		{sizes{10, 20, 30, 40, 50}, 60, false}, // less than max
		{sizes{2}, 1, false},
	}, (sizes).midhinge)
}

func TestTrimean(t *testing.T) {
	avgTester(t, []test{
		{sizes{}, 0, true},
		{sizes{1}, 1, true},
		{sizes{1, 2, 3, 4, 5, 6, 7}, 4, true},
		{sizes{1, 2, 3, 15, 1000, 2000, 3000}, 508, true},
		{sizes{1, 1, 1, 1, 1}, 1, true},        // value preservation
		{sizes{5, 5, 5, 5, 5, 5}, 5, true},     // first-order preservation
		{sizes{10, 20, 30, 40, 50}, 0, false},  // more than min
		{sizes{10, 20, 30, 40, 50}, 60, false}, // less than max
		{sizes{2}, 1, false},
	}, (sizes).median)
}

func TestMedian(t *testing.T) {
	avgTester(t, []test{
		{sizes{}, 0, true},
		{sizes{1}, 1, true},
		{sizes{0, 3, 4, 5, 10}, 4, true},
		{sizes{1, 2, 2, 6, 10, 100}, 4, true},
		{sizes{1, 1, 1, 1, 1}, 1, true},        // value preservation
		{sizes{5, 5, 5, 5, 5, 5}, 5, true},     // first-order preservation
		{sizes{10, 20, 30, 40, 50}, 0, false},  // more than min
		{sizes{10, 20, 30, 40, 50}, 60, false}, // less than max
		{sizes{2}, 1, false},
	}, (sizes).median)
}

func TestMode(t *testing.T) {
	avgTester(t, []test{
		{sizes{1}, 1, true},
		{sizes{1, 1, 1, 2, 2, 3}, 1, true},
		{sizes{1, 1, 1, 2, 3, 3, 3, 3, 4}, 3, true},
		{sizes{1, 10, 10, 100}, 10, true},
		{sizes{10, 100, 1, 10}, 10, true},      // invariance under exchange
		{sizes{1, 1, 1, 1, 1}, 1, true},        // value preservation
		{sizes{5, 5, 5, 5, 5, 5}, 5, true},     // first-order preservation
		{sizes{10, 20, 30, 40, 50}, 0, false},  // more than min
		{sizes{10, 20, 30, 40, 50}, 60, false}, // less than max
		{sizes{2}, 1, false},
	}, (sizes).mode)
}
