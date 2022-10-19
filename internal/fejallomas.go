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

type Fejallomas struct {
	Nev         string
	Irszam      string
	Varos       string
	Terulet     string
	Vegpont_mod string
}

type Fejallomasok struct {
	List     []*Fejallomas
	file     string
	vegponts map[string]*Fejallomas
}

var fej_patterns []*regexp.Regexp

func NewFejallomasok(file string, conf *config.Config) *Fejallomasok {
	am := &Fejallomasok{
		file:     file,
		vegponts: make(map[string]*Fejallomas),
	}
	for _, p := range conf.TeruletPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			log.Println("Invalid pattern ", p, err)
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
		am := NewFejallomas(rStr)
		a.List = append(a.List, am)
		a.vegponts[am.Vegpont_mod] = am
	}
	log.Println("Fejallomás darabszám: ", len(a.List))
}

func (a *Fejallomasok) Vegpont(vegpont string) *Fejallomas {
	if v, ok := a.vegponts[vegpont]; ok {
		return v
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
	a.setVegpont()
	return a
}

func (a *Fejallomas) setVegpont() {
	for _, p := range fej_patterns {
		a.Vegpont_mod = fmt.Sprintf("%s %s", a.Varos, p.ReplaceAllString(a.Terulet, ""))
		break
	}
}

func (a *Fejallomas) String() string {
	return fmt.Sprintf("%s => %s", a.Nev, a.Terulet)
}
