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

type Hoszt struct {
	Irszam       string
	Varos        string
	Terulet      string
	Vegpont_mod1 string
	Vegpont_mod2 string
	Sarok        bool
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
		if am.Sarok {
			a.vegponts[am.Vegpont_mod2] = am
		}
		a.vegponts[am.Vegpont_mod1] = am
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
	a.setVegpont(data[2])
	return a
}

func (a *Hoszt) setVegpont(s string) {
	re := regexp.MustCompile(` sarok`)
        a.Sarok = re.MatchString(s)
	for _, p := range hoszt_patterns {
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

func (a *Hoszt) String() string {
	return fmt.Sprintf("%s hoszt, %s", a.Varos, a.Terulet)
}
