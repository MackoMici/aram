package internal

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
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
}

type ActiveModems struct {
	List     []*ActiveModem
	file     string
	vegponts map[string]*ActiveModem
}

func NewActiveModems(file string) *ActiveModems {
	am := &ActiveModems{
		file:     file,
		vegponts: make(map[string]*ActiveModem),
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
			fmt.Println("ERROR: ", err.Error())
			break
		}
		am := a.parseStruct(rStr)
		a.List = append(a.List, am)
		a.vegponts[am.Vegpont] = am
	}
	fmt.Println("Count ActiveModem ", len(a.List))
//	fmt.Println(a.vegponts)
}

func (a *ActiveModems) parseStruct(data []string) *ActiveModem {
	return &ActiveModem{
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
		Vegpont:      a.Modify(data[10]),
		ID:           data[11],
		Ugyfel:       data[12],
		ElozoAllapot: data[13],
		OfflineDatum: data[14],
	}
}

func (a *ActiveModems) Vegpont(vegpont string) *ActiveModem {
	if v, ok := a.vegponts[vegpont]; ok {
		return v
	}
	return nil
}

func (a *ActiveModems) Modify(s string) string {
	pattern := regexp.MustCompile(`(?: utca| tér| lakótelep| körút| út| szállás| hegy|[0-9]+\.).+`)
	s = pattern.ReplaceAllString(s, "")
        pattern = regexp.MustCompile(`\s+`)
	s = pattern.ReplaceAllString(s, " ")
        return s
}
