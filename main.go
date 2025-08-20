package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"time"
	"strings"
	"strconv"

	"github.com/MackoMici/aram/config"
	"github.com/MackoMici/aram/internal"
	"github.com/MackoMici/aram/logging"
)

type Kiiras struct {
	Tipus string
	Adat  string
	Datum time.Time
	Id    int
}

func main() {

	kiirasok := []Kiiras{}
	seenElements := make(map[string]bool)

	// Parancssori kapcsoló: -debug
	debugMode := flag.Bool("debug", false, "Debug mód engedélyezése")
	flag.Parse()

	logging.Init(*debugMode)
	defer func() {
		if err := logging.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Logfájl bezárás hiba: %v\n", err)
		}
	}()
	logging.Logger.Info("Program elindult", "verzió", "v2.0.3")
	logging.Logger.Debug("Debug mód aktív", "modul", "aram")
	conf  := config.NewConfig("./aram.yaml")
	asz   := internal.NewAramSzunets("./aramszunet.txt", conf)
	node  := internal.NewNodes("./nodeok.txt", conf)
	fej   := internal.NewFejallomasok("./fejallomas.txt", conf)
	hoszt := internal.NewHoszts("./hoszt.txt", conf)
	mux   := internal.NewMuxs("./mux.txt", conf)
	olt   := internal.NewOlts("./olt.txt", conf)

	f, err := os.Create("lehetseges_aramszunet.txt")

	if err != nil {
		logging.Fatal("Lehetséges áramszünet fájl", "hiba", err)
	}

	defer f.Close()

	for _, a := range asz.List {

		datum, err := time.Parse("2006-01-02", a.Datum)
		if err != nil {
			logging.Fatal("Dátum parse hiba", "hiba", err)
		}

		id, err := toInt(a.ID)
		if err != nil {
			logging.Fatal("Id parse hiba", "hiba", err)
		}

		if a.TeljesTel {
			kiirasok = append(kiirasok, Kiiras{
				Tipus: "TELJES",
				Adat:  fmt.Sprintf("A teljes település megáll: %v", a),
				Datum: datum,
				Id:    id,
			})
		} else {
			for _, num := range a.Hazszamok {
				if y := mux.Find(a.Varos, a.Terulet_mod, num); y != nil {
					key := fmt.Sprintf("MUX-%s-%s-%d", a.Varos, a.Terulet_mod, num)
					if !seenElements[key] {
						kiirasok = append(kiirasok, Kiiras{"MUX", fmt.Sprintf("MUX áramszünet miatt ellenőrizni: %v => %v", a, y), datum, id})
						seenElements[key] = true
					}
				}
				if v := node.Find(a.Varos, a.Terulet_mod, num); v != nil {
					key := fmt.Sprintf("NODE-%s-%s-%d", a.Varos, a.Terulet_mod, num)
					if !seenElements[key] {
						kiirasok = append(kiirasok, Kiiras{"NODE", fmt.Sprintf("Node áramszünet miatt ellenőrizni: %v => %v", a, v), datum, id})
						seenElements[key] = true
					}
				}
				if z := fej.Find(a.Varos, a.Terulet_mod, num); z != nil {
					key := fmt.Sprintf("FEJ-%s-%s-%d", a.Varos, a.Terulet_mod, num)
					if !seenElements[key] {
						kiirasok = append(kiirasok, Kiiras{"FEJ", fmt.Sprintf("Fejállomás áramszünet miatt ellenőrizni: %v => %v", a, z), datum, id})
						seenElements[key] = true
					}
				}
				if x := hoszt.Find(a.Varos, a.Terulet_mod, num); x != nil {
					key := fmt.Sprintf("HOSZT-%s-%s-%d", a.Varos, a.Terulet_mod, num)
					if !seenElements[key] {
						kiirasok = append(kiirasok, Kiiras{"HOSZT", fmt.Sprintf("Hoszt áramszünet miatt ellenőrizni: %v => %v", a, x), datum, id})
						seenElements[key] = true
					}
				}
				if w := olt.Find(a.Varos, a.Terulet_mod, num); w != nil {
					key := fmt.Sprintf("OLT-%s-%s-%d", a.Varos, a.Terulet_mod, num)
					if !seenElements[key] {
						kiirasok = append(kiirasok, Kiiras{"OLT", fmt.Sprintf("OLT áramszünet miatt ellenőrizni: %v => %v", a, w), datum, id})
						seenElements[key] = true
					}
				}
			}
		}
	}

	
	// dátum szerinti rendezés
	sort.SliceStable(kiirasok, func(i, j int) bool {
		return kiirasok[i].Datum.Before(kiirasok[j].Datum)
	})

	// fájlba írás a végén
	seenIds := make(map[int]bool)

	for _, k := range kiirasok {
		
		if seenIds[k.Id] {
			continue // már kiírtuk ezt az ID-t
		}
		seenIds[k.Id] = true

		_, err := fmt.Fprintln(f, k.Adat)
		if err != nil {
			logging.Fatal("Kiírás hiba", "hiba", err)
		}
	}
}

func toInt(str string) (int, error) {
	clean := strings.TrimRight(str, ".")
	return strconv.Atoi(clean)
}
