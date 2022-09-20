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

type AramSzunet struct {
	ID          string
	Datum       string
	Idoszak     string
	Varos       string
	VarosLink   string
	Terulet     string
	Megjegyzes  string
	Forras      string
	Bekerules   string
	Terulet_mod string
	TeljesTel   bool
}

type AramSzunets struct {
	List []*AramSzunet
	file string
}

var aram_patterns []*regexp.Regexp
var aram_replace  []*config.Replacements

func NewAramSzunets(file string, conf *config.Config) *AramSzunets {
	am := &AramSzunets{
		file: file,
	}

	for _, p  := range conf.AramszunetPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			log.Println("Invalid pattern ", p, err)
		}
		aram_patterns = append(aram_patterns, re)
	}
	for _, p := range conf.AramszunetReplacements {
		aram_replace = append(aram_replace, p)
	}
	am.Load()
	return am
}

func (a *AramSzunets) Load() {
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
			fmt.Println("ERROR: ", err.Error())
			break
		}
		a.List = append(a.List, NewAramSzunet(rStr))
	}
	log.Println("Áramszünet darabszám: ", len(a.List))
}

func (a *AramSzunet) Vegpont() string {
	return fmt.Sprintf("%s %s", a.Varos, a.Terulet_mod)
}

func NewAramSzunet(data []string) *AramSzunet {
	a := &AramSzunet{
		ID:         data[0],
		Datum:      data[1],
		Idoszak:    data[2],
		Varos:      data[3],
		VarosLink:  data[4],
		Terulet:    data[5],
		Megjegyzes: data[6],
		Forras:     data[7],
		Bekerules:  data[8],
	}
	a.setTerulet(data[5])
	return a
}

func (a *AramSzunet) setTerulet(s string) {
	re := regexp.MustCompile(`Teljes település`)
	a.TeljesTel = re.MatchString(s)
	for _, p := range aram_patterns {
		a.Terulet_mod = teruletMod(p.ReplaceAllString(s, ""))
	}
}

func teruletMod(s string) string {
	if len(aram_replace) > 0 {
		for _, r := range aram_replace {
			s = r.Replace(s)
		}
	}
	return s
}

func (a *AramSzunet) String() string {
	return fmt.Sprintf("%s %s - %s; %s %s", a.ID, a.Datum, a.Idoszak, a.Varos, a.Terulet)
}
