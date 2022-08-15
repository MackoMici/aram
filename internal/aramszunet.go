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
	ID         string
	Datum      string
	Idoszak    string
	Varos      string
	VarosLink  string
	Terulet    string
	Megjegyzes string
	Forras     string
	Bekerules  string
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
		a.List = append(a.List, a.parseStruct(rStr))
	}
	fmt.Println("Count AramSzunet ", len(a.List))
}

func (a *AramSzunets) parseStruct(data []string) *AramSzunet {
	return &AramSzunet{
		ID:         data[0],
		Datum:      data[1],
		Idoszak:    data[2],
		Varos:      data[3],
		VarosLink:  data[4],
		Terulet:    a.Modify(data[5]),
		Megjegyzes: data[6],
		Forras:     data[7],
		Bekerules:  data[8],
	}
}

func (a *AramSzunet) Vegpont() string {
	return fmt.Sprintf("%s %s", a.Varos, a.Terulet)
}

func (a *AramSzunets) Modify(s string) string {
	pattern := regexp.MustCompile(`(?:\.| utca|\:|hrsz|dűlő| - Egész| liget| puszta).+`)
	s = pattern.ReplaceAllString(s, "")
	return s
}
