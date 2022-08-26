package internal

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
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
}

type AramSzunets struct {
	List []*AramSzunet
	file string
}

func NewAramSzunets(file string) *AramSzunets {
	am := &AramSzunets{
		file: file,
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
	pattern := regexp.MustCompile(`(?:\.| utca|\:|hrsz|dűlő| - Egész| liget| puszta).+`)
	a.Terulet_mod = pattern.ReplaceAllString(s, "")
}

func (a *AramSzunet) String() string {
	return fmt.Sprintf("%s %s - %s; %s", a.ID, a.Datum, a.Idoszak, a.Terulet)
}
