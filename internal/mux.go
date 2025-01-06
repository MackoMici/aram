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

type Mux struct {
	Nev          string
	Varos        string
	Terulet      string
	Vegpont_mod1 string
	Vegpont_mod2 string
	Sarok        bool
}

type Muxs struct {
	List     []*Mux
	file     string
	vegponts map[string]*Mux
}

var mux_patterns []*regexp.Regexp

func NewMuxs(file string, conf *config.Config) *Muxs {
	am := &Muxs{
		file:     file,
		vegponts: make(map[string]*Mux),
	}
	for _, p := range conf.TeruletPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			log.Println("Invalid pattern ", p, err)
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
		am := NewMux(rStr)
		a.List = append(a.List, am)
		if am.Sarok {
			a.vegponts[am.Vegpont_mod2] = am
		}
		a.vegponts[am.Vegpont_mod1] = am
	}
	log.Println("Mux darabszám: ", len(a.List))
}

func (a *Muxs) Vegpont(vegpont string) *Mux {
	if v, ok := a.vegponts[vegpont]; ok {
		return v
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

func (a *Mux) String() string {
	return fmt.Sprintf("%s => %s", a.Nev, a.Terulet)
}
