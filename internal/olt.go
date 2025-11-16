//Felolvassa a olt.txt file-t és szétbontja majd csinál egy indexet belőle varos-utca-házszám szintjén, hogy később tudja ellenőrizni, hogy az áramszünet listában megtalálható a cím

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

type Olt struct {
	Nev          string
	Varos        string
	Terulet      string
	Hazszam      int
	Vegpont_mod1 string
	Vegpont_mod2 string
	Sarok        bool
}

type Olts struct {
	List  []*Olt
	file  string
	index map[string]map[string]map[int]*Olt
}

var olt_patterns []*regexp.Regexp

func NewOlts(file string, conf *config.Config) *Olts {
	am := &Olts{
		file:  file,
		index: make(map[string]map[string]map[int]*Olt),
	}
	for _, p := range conf.TeruletPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			logging.Logger.Error("Érvénytelen pattern ", p, err)
		}
		olt_patterns = append(olt_patterns, re)
	}
	am.Load()
	return am
}

func (a *Olts) Load() {
	// open file
	f, err := os.Open(a.file)
	if err != nil {
		logging.Fatal("OLT fájl megnyitás", "hiba", err)
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
			logging.Logger.Error("OLT", "Load", err.Error())
			break
		}
		am := NewOlt(rStr)
		a.List = append(a.List, am)
	}
	logging.Logger.Info("OLT", "darab", len(a.List))
	a.BuildIndex()
}

func (a *Olts) BuildIndex() {
	a.index = make(map[string]map[string]map[int]*Olt)
	for _, olt := range a.List {
		city, street, num := olt.Varos, olt.Vegpont_mod1, olt.Hazszam
		// város tábla init
		if a.index[city] == nil {
			a.index[city] = make(map[string]map[int]*Olt)
		}
		// utca tábla init
		if a.index[city][street] == nil {
			a.index[city][street] = make(map[int]*Olt)
		}
		// indexelt házszám → Olt
		a.index[city][street][num] = olt
		if olt.Sarok {
			street2 := olt.Vegpont_mod2
			a.index[city][street2] = make(map[int]*Olt)
			a.index[city][street2][num] = olt
		}
	}
	logging.Logger.Debug("OLT index", "lista", a.index)
}

func (a *Olts) Find(city, street string, number int) *Olt {
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

func NewOlt(data []string) *Olt {
	a := &Olt{
		Nev:     data[0],
		Varos:   data[1],
		Terulet: data[2],
	}
	a.setVegpont(data[2])
	return a
}

func (a *Olt) setVegpont(s string) {
	re := regexp.MustCompile(` sarok`)
	a.Sarok = re.MatchString(s)
	for _, p := range olt_patterns {
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
					logging.Logger.Error("OLT", numRaw, err.Error())
				}
				a.Hazszam = i
			}
		}
	}
}
