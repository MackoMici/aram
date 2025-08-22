package internal

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"strconv"

	"github.com/MackoMici/aram/config"
	"github.com/MackoMici/aram/logging"
)

type Hoszt struct {
	Irszam       string
	Varos        string
	Terulet      string
	Hazszam      int
	Vegpont_mod1 string
	Vegpont_mod2 string
	Sarok        bool
}

type Hoszts struct {
	List  []*Hoszt
	file  string
	index map[string]map[string]map[int]*Hoszt
}

var hoszt_patterns []*regexp.Regexp

func NewHoszts(file string, conf *config.Config) *Hoszts {
	am := &Hoszts{
		file:  file,
		index: make(map[string]map[string]map[int]*Hoszt),
	}
	for _, p := range conf.TeruletPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			logging.Logger.Error("Érvénytelen pattern ", p, err)
		}
		hoszt_patterns = append(hoszt_patterns, re)
	}
	am.Load()
	return am
}

func (a *Hoszts) Load() {
	// open file
	f, err := os.Open(a.file)
	if err != nil {
		logging.Fatal("Hoszt fájl megnyitás", "hiba", err)
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
			logging.Logger.Error("Hoszt", "Load", err.Error())
			break
		}
		am := NewHoszt(rStr)
		a.List = append(a.List, am)
	}
	logging.Logger.Info("Hoszt", "darab", len(a.List))
	a.BuildIndex()
}

func (a *Hoszts) BuildIndex() {
	a.index = make(map[string]map[string]map[int]*Hoszt)
	for _, hoszt := range a.List {
		city, street, num := hoszt.Varos, hoszt.Vegpont_mod1, hoszt.Hazszam
		// város tábla init
		if a.index[city] == nil {
			a.index[city] = make(map[string]map[int]*Hoszt)
		}
		// utca tábla init
		if a.index[city][street] == nil {
			a.index[city][street] = make(map[int]*Hoszt)
		}
		// indexelt házszám → Hoszt
		a.index[city][street][num] = hoszt
		if hoszt.Sarok {
			street2 := hoszt.Vegpont_mod2
			a.index[city][street2] = make(map[int]*Hoszt)
			a.index[city][street2][num] = hoszt
		}
	}
	logging.Logger.Debug("Hoszt index", "lista", a.index)
}

func (a *Hoszts) Find(city, street string, number int) *Hoszt {
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

func NewHoszt(data []string) *Hoszt {
	a := &Hoszt{
		Irszam:  data[0],
		Varos:   data[1],
		Terulet: data[2],
	}
	a.setVegpont(data[2])
	return a
}

func (a *Hoszt) setVegpont(s string) {
	re := regexp.MustCompile(` sarok`)
		a.Sarok = re.MatchString(s)
	for _, p := range hoszt_patterns {
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
					logging.Logger.Error("Hoszt", numRaw, err.Error())
				}
				a.Hazszam = i
			}
		}
		break
	}
}

func (a *Hoszt) String() string {
	return fmt.Sprintf("%s hoszt, %s", a.Varos, a.Terulet)
}
