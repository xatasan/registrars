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
	Count     uint64
	Sum       uint64
	AritMean  uint64
	HarmMean  uint64
	CHarmMean uint64
	TurncMean uint64
	WinsMean  uint64
	Midrange  uint64
	Midhinge  uint64
	Trimean   uint64
	Median    uint64
	Mode      uint64
}

type sizes []uint64

func (d sizes) sum() (r uint64) {
	for _, v := range d {
		r += v
	}
	return
}

func (d sizes) lehmerMean(p float64) uint64 {
	if len(d) == 0 {
		return 0
	}
	var a1, a2 float64
	for _, v := range d {
		a1 += math.Pow(float64(v), p)
		a2 += math.Pow(float64(v), p-1)
	}
	return uint64(a1 / a2)
}

func (d sizes) arithMean() uint64 {
	return d.lehmerMean(1)
}

func (d sizes) harmMean() uint64 {
	return d.lehmerMean(0)
}

func (d sizes) contraHarmMean() uint64 {
	return d.lehmerMean(2)
}

func (d sizes) turncMean() uint64 {
	l := len(d) / 4
	return d[l : len(d)-l].arithMean()
}

func (d sizes) winsMean() uint64 {
	e := make(sizes, len(d))
	copy(e, d)
	l := len(d) / 4
	for i := 0; i < l; i++ {
		e[i] = e[l]
	}
	for i := len(e) - l - 1; i < len(e); i++ {
		e[i] = e[len(d)-l-1]
	}
	return e.arithMean()
}

func (d sizes) midrange() uint64 {
	if len(d) == 0 {
		return 0
	}
	return (d[0] + d[len(d)-1]) / 2
}

func (d sizes) midhinge() uint64 {
	if len(d) == 0 {
		return 0
	}

	median := d.median()
	mpos := 0
	for i, v := range d {
		if median > v {
			mpos = i
			break
		}
	}
	return ((d[:mpos]).median() + (d[1+mpos:]).median()) / 2
}

func (d sizes) trimean() uint64 {
	if len(d) == 0 {
		return 0
	}

	median := d.median()
	mpos := 0
	for i, v := range d {
		if median > v {
			mpos = i
			break
		}
	}
	return ((d[:mpos]).median() +
		2*median +
		(d[mpos:]).median()) / 4
}

func (d sizes) median() (r uint64) { // assume sorted
	switch len(d) {
	case 0:
	case 1:
		r = d[0]
	default:
		if len(d)%2 == 0 {
			mid := len(d) / 2
			r = (d[mid-1] + d[mid]) / 2
		} else {
			r = d[len(d)/2]
		}
	}
	return
}

func (d sizes) mode() uint64 {
	if len(d) == 0 {
		return 0
	}
	var p, c uint64             // previous, counter
	var pp, pc uint64 = d[0], 1 // prev. previous, prev. counter
	for _, v := range d {
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
	return pp
}

func calcStats(dir string) (Statistics, error) {
	var s sizes
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return Statistics{}, err
	}
	for _, fn := range files {
		fs, err := os.Stat(path.Join(dir, fn.Name()))
		if err == nil {
			s = append(s, uint64(fs.Size()))
		}
	}
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	return Statistics{
		Count:     uint64(len(s)),
		Sum:       s.sum(),
		AritMean:  s.arithMean(),
		HarmMean:  s.harmMean(),
		CHarmMean: s.contraHarmMean(),
		TurncMean: s.turncMean(),
		WinsMean:  s.winsMean(),
		Midrange:  s.midrange(),
		Midhinge:  s.midhinge(),
		Trimean:   s.trimean(),
		Median:    s.median(),
		Mode:      s.mode(),
	}, nil
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
func byteSize(bytes uint64) string {
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	exp := math.Floor(math.Log(float64(bytes)) / math.Log(unit))
	return fmt.Sprintf("%.2f %cB",
		float64(bytes)/(math.Pow(unit, exp)),
		"KMGTPE"[int(exp)-1])

}
