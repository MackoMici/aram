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

type Hoszt struct {
	Irszam      string
	Varos       string
	Terulet     string
	Vegpont_mod string
}

type Hoszts struct {
	List     []*Hoszt
	file     string
	vegponts map[string]*Hoszt
}

var hoszt_patterns []*regexp.Regexp

func NewHoszts(file string, conf *config.Config) *Hoszts {
	am := &Hoszts{
		file:     file,
		vegponts: make(map[string]*Hoszt),
	}
	for _, p := range conf.TeruletPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			log.Println("Invalid pattern ", p, err)
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
		am := NewHoszt(rStr)
		a.List = append(a.List, am)
		a.vegponts[am.Vegpont_mod] = am
	}
	log.Println("Hoszt darabszám: ", len(a.List))
}

func (a *Hoszts) Vegpont(vegpont string) *Hoszt {
	if v, ok := a.vegponts[vegpont]; ok {
		return v
	}
	return nil
}

func NewHoszt(data []string) *Hoszt {
	a := &Hoszt{
		Irszam:  data[0],
		Varos:   data[1],
		Terulet: data[2],
	}
	a.setVegpont()
	return a
}

func (a *Hoszt) setVegpont() {
	for _, p := range hoszt_patterns {
		a.Vegpont_mod = fmt.Sprintf("%s %s", a.Varos, p.ReplaceAllString(a.Terulet, ""))
		break
	}
}

func (a *Hoszt) String() string {
	return fmt.Sprintf("%s hoszt, %s", a.Varos, a.Terulet)
}
