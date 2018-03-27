package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path"
	"sort"
	"time"
)

type Statistics struct {
	data      []uint64
	Count     int
	Sum       float64
	AritMean  float64
	GeomMean  float64
	HarmMean  float64
	CHarmMean float64
	TurncMean float64
	WinsMean  float64
	Midrange  float64
	Midhinge  float64
	Trimean   float64
	Median    float64
	Mode      float64
}

func mkStat(d []uint64) Statistics {
	return Statistics{data: d}
}

func (d Statistics) count() int {
	d.Count = len(d.data)
	return d.Count
}

func (d Statistics) sum() float64 {
	if d.Sum == 0 {
		var a uint64
		for _, v := range d.data {
			a += v
		}
		d.Sum = float64(a)
	}
	return d.Sum
}

func (d Statistics) lehmerMean(p float64) float64 {
	if len(d.data) == 0 {
		return 0
	}
	var a1, a2 float64
	for _, v := range d.data {
		a1 += math.Pow(float64(v), p)
		a2 += math.Pow(float64(v), p-1)
	}
	return a1 / a2
}

func (d Statistics) arithMean() float64 {
	if d.AritMean == 0 {
		d.AritMean = d.lehmerMean(1)
	}
	return d.AritMean
}

// based off https://rosettacode.org/wiki/Nth_root#Go
func nthRoot(a float64, n int) float64 {
	var n1f, rn, x, x0 float64 = float64(n - 1), 1 / float64(n), 1, 0
	for {
		potx, t2 := 1/x, a
		for b := n - 1; b > 0; b >>= 1 {
			if b&1 == 1 {
				t2 *= potx
			}
			potx *= potx
		}
		x0, x = x, rn*(n1f*x+t2)
		if math.Abs(x-x0)*1e10 < x {
			break
		}
	}
	return x
}

func (d Statistics) geomMean() float64 {
	if d.GeomMean == 0 {
		pi := 1.0
		for _, v := range d.data {
			pi *= float64(v)
		}
		d.GeomMean = nthRoot(pi, len(d.data))
	}
	return d.GeomMean
}

func (d Statistics) harmMean() float64 {
	if d.HarmMean == 0 {
		d.HarmMean = d.lehmerMean(0)
	}
	return d.HarmMean
}

func (d Statistics) contraHarmMean() float64 {
	if d.CHarmMean == 0 {
		d.CHarmMean = d.lehmerMean(2)
	}
	return d.CHarmMean
}

func (d Statistics) turncMean() float64 {
	if d.TurncMean == 0 {
		l := len(d.data) / 4
		d.TurncMean = mkStat(d.data[l : len(d.data)-l]).arithMean()
	}
	return d.TurncMean
}

func (d Statistics) winsMean() float64 {
	if d.WinsMean == 0 {
		e := make([]uint64, len(d.data))
		copy(e, d.data)
		l := len(d.data) / 4
		for i := 0; i < l; i++ {
			e[i] = e[l]
		}
		for i := len(e) - l - 1; i < len(e); i++ {
			e[i] = e[len(d.data)-l-1]
		}
		d.WinsMean = mkStat(e).arithMean()
	}
	return d.WinsMean
}

func (d Statistics) midrange() float64 {
	if d.Midrange == 0 {
		if len(d.data) == 0 {
			return 0
		}
		d.Midrange = float64((d.data[0] + d.data[len(d.data)-1]) / 2)
	}
	return d.Midrange
}

func (d Statistics) midhinge() float64 {
	if d.Midhinge == 0 {
		if len(d.data) == 0 {
			return 0
		}

		median := d.median()
		mpos := 0
		for i, v := range d.data {
			if median > float64(v) {
				mpos = i
				break
			}
		}
		d.Midhinge = float64((mkStat(d.data[:mpos]).median() +
			mkStat(d.data[1+mpos:]).median()) / 2)
	}
	return d.Midhinge
}

func (d Statistics) trimean() float64 {
	if len(d.data) == 0 {
		return 0
	}

	median := d.median()
	mpos := 0
	for i, v := range d.data {
		if median > float64(v) {
			mpos = i
			break
		}
	}
	d.Trimean = (mkStat(d.data[:mpos]).median() +
		2*median +
		mkStat(d.data[mpos:]).median()) / 4
	return d.Trimean
}

func (d Statistics) median() float64 { // assume sorted
	if d.Median == 0 {
		var r uint64
		switch len(d.data) {
		case 0:
		case 1:
			r = d.data[0]
		default:
			if len(d.data)%2 == 0 {
				mid := len(d.data) / 2
				r = (d.data[mid-1] + d.data[mid]) / 2
			} else {
				r = d.data[len(d.data)/2]
			}
		}
		d.Median = float64(r)
	}
	return d.Median
}

func (d Statistics) mode() float64 {
	if d.Mode == 0 {
		if len(d.data) == 0 {
			return 0
		}
		var p, c uint64                  // previous, counter
		var pp, pc uint64 = d.data[0], 1 // prev. previous, prev. counter
		for _, v := range d.data {
			v = v - (v % (1 << 10)) // crop sub-MB data
			if p == v {
				c++
				if c > pc {
					pp = v
					pc = c
				}
			} else {
				c = 1
			}
			p = v
		}
		d.Mode = float64(pp)
	}
	return d.Mode
}

func (d Statistics) calculate() {
	d.count()
	d.sum()
	d.arithMean()
	d.geomMean()
	d.harmMean()
	d.contraHarmMean()
	d.turncMean()
	d.winsMean()
	d.midrange()
	d.midhinge()
	d.trimean()
	d.median()
	d.mode()
}

func calcStats(dir string) (Statistics, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return Statistics{}, err
	}
	var s Statistics
	for _, fn := range files {
		fs, err := os.Stat(path.Join(dir, fn.Name()))
		if err == nil {
			s.data = append(s.data, uint64(fs.Size()))
		}
	}
	sort.Slice(s.data, func(i, j int) bool { return s.data[i] < s.data[j] })
	s.calculate()
	return s, nil
}

func statWorker() {
	for {
		buf := bytes.NewBuffer(nil)
		hs, _ := calcStats(hdir)
		us, _ := calcStats(udir)
		t.ExecuteTemplate(buf, "index", struct {
			Stats   struct{ File, Hash Statistics }
			MaxSize uint64
		}{
			struct{ File, Hash Statistics }{hs, us},
			maxf,
		})
		index = buf.Bytes()
		time.Sleep(time.Minute * 30)
	}
}

const unit = 1024

// taken from Kagami
func byteSize(bytes float64) string {
	if bytes < unit {
		return fmt.Sprintf("%.0fB", bytes)
	}
	exp := math.Floor(math.Log(bytes) / math.Log(unit))
	return fmt.Sprintf("%.2f %cB",
		bytes/(math.Pow(unit, exp)),
		"KMGTPE"[int(exp)-1])
}
