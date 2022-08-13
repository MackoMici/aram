package main

import (
    "encoding/csv"
    "fmt"
    "io"
    "log"
    "os"
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

func main() {
    // open file
    f1, err := os.Open("./aram/activemodemlist.csv")
    if err != nil {
	log.Fatal(err)
    }
    f2, err := os.Open("./aram/aramszunet.txt")
    if err != nil {
        log.Fatal(err)
    }

    // remember to close the file at the end of the program
    defer f1.Close()
    defer f2.Close()

    activemodemRS(f1)
    aramszunetRS(f2)
}

func aramszunetRS(f *os.File) {
	fcsv := csv.NewReader(f)
        fcsv.Comma = '\t'
	rs := make([]*AramSzunet, 0)
	for {
		rStr, err := fcsv.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("ERROR: ", err.Error())
			break
		}
		rs = append(rs, parseStructAram(rStr))
	}
	fmt.Println("Count AramSzunet ", len(rs))
}

func parseStructAram(data []string) *AramSzunet{
	return &AramSzunet{
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
}

func activemodemRS(f *os.File) {
	fcsv := csv.NewReader(f)
        fcsv.Comma = ';'
	rs := make([]*ActiveModem, 0)
	for {
		rStr, err := fcsv.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("ERROR: ", err.Error())
			break
		}
		rs = append(rs, parseStruct(rStr))
	}
	fmt.Println("Count ActiveModem ", len(rs))
//        fmt.Println(rs[0].DS)
}

func parseStruct(data []string) *ActiveModem {
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
		Vegpont:      data[10],
		ID:           data[11],
		Ugyfel:       data[12],
		ElozoAllapot: data[13],
		OfflineDatum: data[14],
	}
}
