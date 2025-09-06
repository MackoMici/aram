//Megvizsgálja hogy az adott végpont (hoszt, node, mux, olt, fejáll.) megtalálható-e az adott áramszünetes listában ha igen akkor kiírja dátum szerint rendezve, és minden találatot csak 1* fog kiírni

package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/MackoMici/aram/config"
	"github.com/MackoMici/aram/internal"
	"github.com/MackoMici/aram/logging"
)

type Kiiras struct {
	Tipus string
	Adat  string
	Datum time.Time
}

func main() {

	kiirasok := []Kiiras{}

	// Parancssori kapcsoló: -debug
	debugMode := flag.Bool("debug", false, "Debug mód engedélyezése")
	flag.Parse()

	logging.Init(*debugMode)
	defer func() {
		if err := logging.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Logfájl bezárás hiba: %v\n", err)
		}
	}()
	logging.Logger.Info("Program elindult", "verzió", "v2.0.8")
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

		if a.TeljesTel {
			kiirasok = append(kiirasok, Kiiras{
				Tipus: "TELJES",
				Adat:  fmt.Sprintf("A teljes település megáll: %v", a),
				Datum: datum,
			})
		} else {
			for _, num := range a.Hazszamok {
				if vs := node.Find(a.Varos, a.Terulet_mod, num); vs != nil {
					for _, v := range vs {
						kiirasok = append(kiirasok, Kiiras{
							Tipus: fmt.Sprintf("NODE: %s", v.Node), // vagy v.Name, ha az a mező neve
							Adat:  fmt.Sprintf("Node áramszünet miatt ellenőrizni: %v => %v", a, v),
							Datum: datum,
						})
					}
				}
				if y := mux.Find(a.Varos, a.Terulet_mod, num); y != nil {
					kiirasok = append(kiirasok, Kiiras{
						Tipus: fmt.Sprintf("MUX: %s", y.Nev), // vagy y.Name, ha az a mező neve
						Adat:  fmt.Sprintf("Mux áramszünet miatt ellenőrizni: %v => %v", a, y),
						Datum: datum,
					})
				}
				if z := fej.Find(a.Varos, a.Terulet_mod, num); z != nil {
					kiirasok = append(kiirasok, Kiiras{
						Tipus: fmt.Sprintf("FEJ: %s", z.Nev), // vagy z.Name, ha az a mező neve
						Adat:  fmt.Sprintf("Fejállomás áramszünet miatt ellenőrizni: %v => %v", a, z),
						Datum: datum,
					})
				}
				if x := hoszt.Find(a.Varos, a.Terulet_mod, num); x != nil {
					kiirasok = append(kiirasok, Kiiras{
						Tipus: fmt.Sprintf("HOSZT: %s", x.Varos), // vagy x.Name, ha az a mező neve
						Adat:  fmt.Sprintf("Hoszt áramszünet miatt ellenőrizni: %v => %v", a, x),
						Datum: datum,
					})
				}
				if w := olt.Find(a.Varos, a.Terulet_mod, num); w != nil {
					kiirasok = append(kiirasok, Kiiras{
						Tipus: fmt.Sprintf("OLT: %s", w.Nev), // vagy w.Name, ha az a mező neve
						Adat:  fmt.Sprintf("OLT áramszünet miatt ellenőrizni: %v => %v", a, w),
						Datum: datum,
					})
				}
			}
		}
	}

	
	// dátum szerinti rendezés
	sort.SliceStable(kiirasok, func(i, j int) bool {
		return kiirasok[i].Datum.Before(kiirasok[j].Datum)
	})

	// fájlba írás a végén
	seenIds := make(map[Kiiras]bool)

	for _, k := range kiirasok {
		
		logging.Logger.Debug("Kiírás", "tipus", k.Tipus, "adat", k.Adat, "datum", k.Datum)
		key := Kiiras{Tipus: k.Tipus, Datum: k.Datum}
		if seenIds[key] {
			continue // már kiírtuk ezt a Tipus+Dátum kombinációt
		}
		seenIds[key] = true

		_, err := fmt.Fprintln(f, k.Adat)
		if err != nil {
			logging.Fatal("Kiírás hiba", "hiba", err)
		}
	}
}