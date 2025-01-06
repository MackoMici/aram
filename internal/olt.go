package internal

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/MackoMici/aram/config"
)

type Olt struct {
	Nev          string
	Varos        string
	Terulet      string
	Vegpont_mod1 string
	Vegpont_mod2 string
	Sarok        bool
}

type Olts struct {
	List     []*Olt
	file     string
	vegponts map[string]*Olt
}

var olt_patterns []*regexp.Regexp

func NewOlts(file string, conf *config.Config) *Olts {
	am := &Olts{
		file:     file,
		vegponts: make(map[string]*Olt),
	}
	for _, p := range conf.TeruletPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			log.Println("Invalid pattern ", p, err)
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
		log.Fatal(err)
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
			log.Println("ERROR: ", err.Error())
			break
		}
		am := NewOlt(rStr)
		a.List = append(a.List, am)
		if am.Sarok {
			a.vegponts[am.Vegpont_mod2] = am
		}
		a.vegponts[am.Vegpont_mod1] = am
	}
	log.Println("Olt darabszám: ", len(a.List))
}

func (a *Olts) Vegpont(vegpont string) *Olt {
	if v, ok := a.vegponts[vegpont]; ok {
		return v
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
		if a.Sarok {
			r := strings.Split(s, " - ")
			a.Vegpont_mod2 = fmt.Sprintf("%s %s", a.Varos, p.ReplaceAllString(r[1], ""))
			a.Vegpont_mod1 = fmt.Sprintf("%s %s", a.Varos, p.ReplaceAllString(r[0], ""))
		} else {
			a.Vegpont_mod1 = fmt.Sprintf("%s %s", a.Varos, p.ReplaceAllString(s, ""))
		}
		break
	}
}

func (a *Olt) String() string {
	return fmt.Sprintf("%s, %s", a.Nev, a.Terulet)
}
