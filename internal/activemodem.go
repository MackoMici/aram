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

type ActiveModem struct {
	DS           string
	US           string
	MAC          string
	Sszam        string
	IP           string
	Allapot      string
	CmtsRx       string
	Cmts         string
	SNR          string
	Node         string
	Vegpont      string
	ID           string
	Ugyfel       string
	ElozoAllapot string
	OfflineDatum string
	Vegpont_mod  string
	Varos        string
	Terulet      string
}

type ActiveModems struct {
	List     []*ActiveModem
	file     string
	vegponts map[string]*ActiveModem
}

var patterns []*regexp.Regexp

func NewActiveModems(file string) *ActiveModems {
	am := &ActiveModems{
		file:     file,
		vegponts: make(map[string]*ActiveModem),
	}
	conf := config.NewConfig("./aram.yaml")
	for _, p  := range conf.VegpontPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			log.Println("Invalid pattern ", p, err)
		}
		patterns = append (patterns, re)
	}
	am.Load()
	return am
}

func (a *ActiveModems) Load() {
	// open file
	f, err := os.Open(a.file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fcsv := csv.NewReader(f)
	fcsv.Comma = ';'
	for {
		rStr, err := fcsv.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("ERROR: ", err.Error())
			break
		}
		am := NewActiveModem(rStr)
		a.List = append(a.List, am)
		a.vegponts[am.Vegpont_mod] = am
	}
	log.Println("Activemodem darabszám: ", len(a.List))
}

func (a *ActiveModems) Vegpont(vegpont string) *ActiveModem {
	if v, ok := a.vegponts[vegpont]; ok {
		return v
	}
	return nil
}

func NewActiveModem(data []string) *ActiveModem {
	a := &ActiveModem{
		DS:           data[0],
		US:           data[1],
		MAC:          data[2],
		Sszam:        data[3],
		IP:           data[4],
		Allapot:      data[5],
		CmtsRx:       data[6],
		Cmts:         data[7],
		SNR:          data[8],
		Node:         data[9],
		Vegpont:      data[10],
		ID:           data[11],
		Ugyfel:       data[12],
		ElozoAllapot: data[13],
		OfflineDatum: data[14],
	}
	a.setVegpont(data[10])
	return a
}

func (a *ActiveModem) setVegpont(s string) {
	for _, p := range patterns {
		if namedGroups := a.matchWithGroup(p, s); len(namedGroups) > 0 {
			//			fmt.Printf("Processing: %s\n", s)
			a.Varos = a.varos(namedGroups)
			a.Terulet = a.terulet(namedGroups)
			a.Vegpont_mod = fmt.Sprintf("%s %s", a.Varos, a.Terulet)
//			fmt.Printf("Processing: %s => %s, %s\n", s, a.Varos, a.Terulet)
			//			maradek := a.Maradek(namedGroups)
			break
		}
	}
}

func (a *ActiveModem) matchWithGroup(r *regexp.Regexp, s string) map[string]string {
	namedGroups := make(map[string]string)
	if match := r.FindStringSubmatch(s); len(match) > 0 {
		for i, name := range r.SubexpNames() {
			if i != 0 && name != "" {
				namedGroups[name] = match[i]
			}
		}
	}
	return namedGroups
}

func (a *ActiveModem) varos(namedgroups map[string]string) string {
	s, ok := namedgroups["Varos"]
	if !ok {
		return ""
	}
	return s
}

func (a *ActiveModem) terulet(namedgroups map[string]string) string {
	s, ok := namedgroups["Terulet"]
	if !ok {
		return ""
	}
	return s
}

func (a *ActiveModem) String() string {
	return fmt.Sprintf("%s node: %s", a.Node, a.Vegpont)
}
