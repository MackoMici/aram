//Felolvassa a fejallomas.txt file-t és szétbontja majd csinál egy indexet belőle varos-utca-házszám szintjén, hogy később tudja ellenőrizni, hogy az áramszünet listában megtalálható a cím

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

type Fejallomas struct {
	Nev          string
	Irszam       string
	Varos        string
	Terulet      string
	Hazszam      int
	Vegpont_mod1 string
	Vegpont_mod2 string
	Sarok        bool
}

type Fejallomasok struct {
	List  []*Fejallomas
	file  string
	index map[string]map[string]map[int]*Fejallomas
}

var fej_patterns []*regexp.Regexp

func NewFejallomasok(file string, conf *config.Config) *Fejallomasok {
	am := &Fejallomasok{
		file:  file,
		index: make(map[string]map[string]map[int]*Fejallomas),
	}
	for _, p := range conf.TeruletPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			logging.Logger.Error("Érvénytelen pattern ", p, err)
		}
		fej_patterns = append(fej_patterns, re)
	}
	am.Load()
	return am
}

func (a *Fejallomasok) Load() {
	// open file
	f, err := os.Open(a.file)
	if err != nil {
		logging.Fatal("Fejállomás fájl megnyitás", "hiba", err)
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
			logging.Logger.Error("Fejállomás", "Load", err.Error())
			break
		}
		am := NewFejallomas(rStr)
		a.List = append(a.List, am)
	}
	logging.Logger.Info("Fejállomás", "darab", len(a.List))
	a.BuildIndex()
}

func (a *Fejallomasok) BuildIndex() {
	a.index = make(map[string]map[string]map[int]*Fejallomas)
	for _, fejallomas := range a.List {
		city, street, num := fejallomas.Varos, fejallomas.Vegpont_mod1, fejallomas.Hazszam
		// város tábla init
		if a.index[city] == nil {
			a.index[city] = make(map[string]map[int]*Fejallomas)
		}
		// utca tábla init
		if a.index[city][street] == nil {
			a.index[city][street] = make(map[int]*Fejallomas)
		}
		// indexelt házszám → Fejallomas
		a.index[city][street][num] = fejallomas
		if fejallomas.Sarok {
			street2 := fejallomas.Vegpont_mod2
			a.index[city][street2] = make(map[int]*Fejallomas)
			a.index[city][street2][num] = fejallomas
		}
	}
	logging.Logger.Debug("Fejállomás index", "lista", a.index)
}

func (a *Fejallomasok) Find(city, street string, number int) *Fejallomas {
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

func NewFejallomas(data []string) *Fejallomas {
	a := &Fejallomas{
		Nev:     data[0],
		Irszam:  data[1],
		Varos:   data[2],
		Terulet: data[3],
	}
	a.setVegpont(data[3])
	return a
}

func (a *Fejallomas) setVegpont(s string) {
	re := regexp.MustCompile(` sarok`)
	a.Sarok = re.MatchString(s)
	for _, p := range fej_patterns {
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
					logging.Logger.Error("Fejállomás", numRaw, err.Error())
				}
				a.Hazszam = i
			}
		}
	}
}
