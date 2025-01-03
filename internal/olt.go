package internal

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"

	"github.com/MackoMici/aram/config"
)

type Olt struct {
	Nev         string
	Varos       string
	Terulet     string
	Vegpont_mod string
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
		a.vegponts[am.Vegpont_mod] = am
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
	a.setVegpont()
	return a
}

func (a *Olt) setVegpont() {
	for _, p := range olt_patterns {
		a.Vegpont_mod = fmt.Sprintf("%s %s", a.Varos, p.ReplaceAllString(a.Terulet, ""))
		break
	}
}

func (a *Olt) String() string {
	return fmt.Sprintf("%s, %s", a.Nev, a.Terulet)
}
