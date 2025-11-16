//Felolvassa a mux.txt file-t és szétbontja majd csinál egy indexet belőle varos-utca-házszám szintjén, hogy később tudja ellenőrizni, hogy az áramszünet listában megtalálható a cím

package internal

import (
	"encoding/csv"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/MackoMici/aram/config"
	"github.com/MackoMici/aram/logging"
)

type Mux struct {
	Nev          string
	Varos        string
	Terulet      string
	Hazszam      int
	Vegpont_mod1 string
	Vegpont_mod2 string
	Sarok        bool
}

type Muxs struct {
	List  []*Mux
	file  string
	index map[string]map[string]map[int]*Mux
}

var mux_patterns []*regexp.Regexp

func NewMuxs(file string, conf *config.Config) *Muxs {
	am := &Muxs{
		file:  file,
		index: make(map[string]map[string]map[int]*Mux),
	}
	for _, p := range conf.TeruletPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			logging.Logger.Error("Érvénytelen pattern ", p, err)
		}
		mux_patterns = append(mux_patterns, re)
	}
	am.Load()
	return am
}

func (a *Muxs) Load() {
	// open file
	f, err := os.Open(a.file)
	if err != nil {
		logging.Fatal("Mux fájl megnyitás", "hiba", err)
	}
	defer f.Close()
	fcsv := csv.NewReader(f)
	fcsv.Comma = '\t'
	for {
		rStr, err := fcsv.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			logging.Logger.Error("Mux", "Load", err.Error())
			break
		}
		am := NewMux(rStr)
		a.List = append(a.List, am)
	}
	logging.Logger.Info("Mux", "darab", len(a.List))
	a.BuildIndex()
}

func (a *Muxs) BuildIndex() {
	a.index = make(map[string]map[string]map[int]*Mux)

	for _, mux := range a.List {
		city, street, num := mux.Varos, mux.Vegpont_mod1, mux.Hazszam
		// város tábla init
		if a.index[city] == nil {
			a.index[city] = make(map[string]map[int]*Mux)
		}
		// utca tábla init
		if a.index[city][street] == nil {
			a.index[city][street] = make(map[int]*Mux)
		}
		// indexelt házszám → Mux
		a.index[city][street][num] = mux
		if mux.Sarok {
			street2 := mux.Vegpont_mod2
			a.index[city][street2] = make(map[int]*Mux)
			a.index[city][street2][num] = mux
		}
	}
	logging.Logger.Debug("Mux index", "lista", a.index)
}

func (a *Muxs) Find(city, street string, number int) *Mux {
	if a.index == nil {
		a.BuildIndex()
	}
	if cityMap, ok := a.index[city]; ok {
		if streetMap, ok := cityMap[street]; ok {
			if m, ok := streetMap[number]; ok {
				return m
			}
			for _, m := range streetMap {
				if m.Sarok {
					return m
				}
			}
		}
	}
	return nil
}

func NewMux(data []string) *Mux {
	a := &Mux{
		Nev:     data[0],
		Varos:   data[1],
		Terulet: data[2],
	}
	a.setVegpont(data[2])
	return a
}

func (a *Mux) setVegpont(s string) {
	re := regexp.MustCompile(` sarok`)
	a.Sarok = re.MatchString(s)
	for _, p := range mux_patterns {
		parts := p.FindStringSubmatch(s)
		streetRaw, numRaw := parts[1], parts[2]
		if a.Sarok {
			r := strings.Split(streetRaw, " - ")
			a.Vegpont_mod2 = r[1]
			a.Vegpont_mod1 = r[0]
		} else {
			a.Vegpont_mod1 = streetRaw
			if numRaw != "" {
				i, err := strconv.Atoi(numRaw)
				if err != nil {
					logging.Logger.Error("Mux", numRaw, err.Error())
				}
				a.Hazszam = i
			}
		}
	}
}
