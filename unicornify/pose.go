package unicornify

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
	"math"
	"sort"
)

type tv struct {
	t float64
	v float64
}

type tvSorter struct {
	tvs []tv
}

func (s tvSorter) Len() int {
	return len(s.tvs)
}

func (s tvSorter) Less(i, j int) bool {
	return s.tvs[i].t < s.tvs[j].t
}

func (s tvSorter) Swap(i, j int) {
	s.tvs[i], s.tvs[j] = s.tvs[j], s.tvs[i]
}

func interpol(tvs ...tv) func(float64) float64 {
	sort.Sort(tvSorter{tvs})
	l := make([]tv, len(tvs)+2)
	last := tvs[len(tvs)-1]
	l[0] = tv{t: last.t - 1, v: last.v}
	l[len(tvs)+1] = tv{t: tvs[0].t + 1, v: tvs[0].v}
	copy(l[1:len(tvs)+1], tvs)

	return func(t float64) float64 {
		for t < 0 {
			t++
		}
		_, t = math.Modf(t)
		var t1, v1, t2, v2 float64
		t1 = -2
		t2 = 2
		for _, tv := range l {
			if tv.t <= t && tv.t > t1 {
				t1, v1 = tv.t, tv.v
			}
			if tv.t >= t && tv.t < t2 {
				t2, v2 = tv.t, tv.v
			}
		}
		if t1 == t2 {
			return v1
		}
		return MixFloats(v1, v2, (t-t1)/(t2-t1))
	}
}

func RotatoryGallop(u *Unicorn, phase float64) {
	// movement per phase: ca. 125
	fl, fr, bl, br := u.Legs[0], u.Legs[1], u.Legs[2], u.Legs[3]

	// approximated from http://commons.wikimedia.org/wiki/File:Horse_gif_slow.gif
	frontTop := interpol(tv{9. / 12, 74}, tv{2.5 / 12, -33})
	frontBottom := interpol(tv{2. / 12, 0}, tv{6. / 12, -107}, tv{8. / 12, -90}, tv{10. / 12, 0})
	backTop := interpol(tv{11. / 12, -53}, tv{4. / 12, 0}, tv{6. / 12, 0})
	backBottom := interpol(tv{11. / 12, 0}, tv{1.5 / 12, 90}, tv{6. / 12, 30}, tv{8. / 12, 50})

	fr.Knee.RotateAround(*fr.Hip, frontTop(phase)*DEGREE, 2)
	fr.Hoof.RotateAround(*fr.Hip, frontTop(phase)*DEGREE, 2)
	fr.Hoof.RotateAround(*fr.Knee, frontBottom(phase)*DEGREE, 2)

	fl.Knee.RotateAround(*fl.Hip, frontTop(phase-.25)*DEGREE, 2)
	fl.Hoof.RotateAround(*fl.Hip, frontTop(phase-.25)*DEGREE, 2)
	fl.Hoof.RotateAround(*fl.Knee, frontBottom(phase-.25)*DEGREE, 2)

	br.Knee.RotateAround(*br.Hip, backTop(phase)*DEGREE, 2)
	br.Hoof.RotateAround(*br.Hip, backTop(phase)*DEGREE, 2)
	br.Hoof.RotateAround(*br.Knee, backBottom(phase)*DEGREE, 2)

	bl.Knee.RotateAround(*bl.Hip, backTop(phase-.167)*DEGREE, 2)
	bl.Hoof.RotateAround(*bl.Hip, backTop(phase-.167)*DEGREE, 2)
	bl.Hoof.RotateAround(*bl.Knee, backBottom(phase-.167)*DEGREE, 2)
}

func Walk(u *Unicorn, phase float64) {

	fl, fr, bl, br := u.Legs[0], u.Legs[1], u.Legs[2], u.Legs[3]

	//approximated from http://de.wikipedia.org/w/index.php?title=Datei:Muybridge_horse_walking_animated.gif&filetimestamp=20061003154457
	frontTop := interpol(tv{6.5 / 9, 40}, tv{2.5 / 9, -35})
	frontBottom := interpol(tv{7. / 9, 0}, tv{2. / 9, 0}, tv{5. / 9, -70})
	backTop := interpol(tv{1. / 9, -35}, tv{4. / 9, 0}, tv{6. / 12, 0})
	backBottom := interpol(tv{5. / 9, 40}, tv{9. / 9, 10})

	fr.Knee.RotateAround(*fr.Hip, frontTop(phase)*DEGREE, 2)
	fr.Hoof.RotateAround(*fr.Hip, frontTop(phase)*DEGREE, 2)
	fr.Hoof.RotateAround(*fr.Knee, frontBottom(phase)*DEGREE, 2)

	fl.Knee.RotateAround(*fl.Hip, frontTop(phase-.56)*DEGREE, 2)
	fl.Hoof.RotateAround(*fl.Hip, frontTop(phase-.56)*DEGREE, 2)
	fl.Hoof.RotateAround(*fl.Knee, frontBottom(phase-.56)*DEGREE, 2)

	br.Knee.RotateAround(*br.Hip, backTop(phase)*DEGREE, 2)
	br.Hoof.RotateAround(*br.Hip, backTop(phase)*DEGREE, 2)
	br.Hoof.RotateAround(*br.Knee, backBottom(phase)*DEGREE, 2)

	bl.Knee.RotateAround(*bl.Hip, backTop(phase-.44)*DEGREE, 2)
	bl.Hoof.RotateAround(*bl.Hip, backTop(phase-.44)*DEGREE, 2)
	bl.Hoof.RotateAround(*bl.Knee, backBottom(phase-.44)*DEGREE, 2)
}

var Poses = [...]func(*Unicorn, float64){RotatoryGallop, Walk}
