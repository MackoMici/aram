//Megvizsgálja hogy az adott végpont (hoszt, node, mux, olt, fejáll.) megtalálható-e az adott áramszünetes listában ha igen akkor kiírja dátum szerint rendezve, és minden találatot csak 1* fog kiírni

package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"slices"
	"sort"
	"strconv"
	"time"

	"github.com/MackoMici/aram/config"
	"github.com/MackoMici/aram/internal"
	"github.com/MackoMici/aram/logging"
)

type Kiiras struct {
	Tipus           string
	Sorszam         string
	Datum           time.Time
	Idoszak         string
	Hol             string
	Megnevezes      string
	Vegpont         string
	modemek_szama   int
	vegpontok_szama int
}

func main() {

	// Parancssori kapcsoló: -debug -check
	debugMode := flag.Bool("debug", false, "Debug mód engedélyezése")
	checkOnly := flag.Bool("check", false, "Csak ellenőrzés futtatása")
	countCheck := flag.Bool("count", false, "Node nevekhez tartozó modemek számlálása")
	flag.Parse()

	logging.Init(*debugMode)
	defer func() {
		if err := logging.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Logfájl bezárás hiba: %v\n", err)
		}
	}()

	logging.Logger.Info("Program elindult", "verzió", "v3.3.5")

	switch {
	case *checkOnly:
		runCheck()
	case *countCheck:
		runCountCheck()
	default:

		kiirasok := []Kiiras{}

		// Konfiguráció és adatok betöltése
		conf := config.NewConfig("./aram.yaml")
		asz := internal.NewAramSzunets("./aramszunet.txt", conf)
		node := internal.NewNodes("./nodeok.txt", conf)
		modem := internal.NewActiveModems("./activemodemlist.csv", conf)
		fej := internal.NewFejallomasok("./fejallomas.txt", conf)
		hoszt := internal.NewHoszts("./hoszt.txt", conf)
		mux := internal.NewMuxs("./mux.txt", conf)
		olt := internal.NewOlts("./olt.txt", conf)

		// CSV fájl létrehozása
		f, err := os.Create("lehetseges_aramszunet.csv")
		if err != nil {
			logging.Fatal("Lehetséges áramszünet fájl", "hiba", err)
		}
		defer f.Close()

		// Adatok feldolgozása
		for _, a := range asz.List {
			datum, err := time.Parse("2006-01-02", a.Datum)
			if err != nil {
				logging.Fatal("Dátum parse hiba", "hiba", err)
			}

			for _, num := range a.Hazszamok {
				if vs := node.Find(a.Varos, a.Terulet_mod, num); vs != nil {
					for _, v := range vs {
						modems := modem.FindByNode(v.Node)
						affected := internal.FilterAffectedModems(modems, asz, datum)
						kiirasok = append(kiirasok, Kiiras{
							Tipus:           "NODE",
							Sorszam:         a.ID,
							Datum:           datum,
							Idoszak:         a.Idoszak,
							Hol:             fmt.Sprintf("%s, %s", a.Varos, a.Terulet),
							Megnevezes:      v.Node,
							Vegpont:         fmt.Sprintf("%s, %s", v.Varos, v.Terulet),
							modemek_szama:   len(modems),
							vegpontok_szama: len(affected),
						})
					}
				}
				if y := mux.Find(a.Varos, a.Terulet_mod, num); y != nil {
					kiirasok = append(kiirasok, Kiiras{
						Tipus:      "MUX",
						Sorszam:    a.ID,
						Datum:      datum,
						Idoszak:    a.Idoszak,
						Hol:        fmt.Sprintf("%s, %s", a.Varos, a.Terulet),
						Megnevezes: y.Nev,
						Vegpont:    fmt.Sprintf("%s, %s", y.Varos, y.Terulet),
					})
				}
				if z := fej.Find(a.Varos, a.Terulet_mod, num); z != nil {
					kiirasok = append(kiirasok, Kiiras{
						Tipus:      "FEJÁLLOMÁS",
						Sorszam:    a.ID,
						Datum:      datum,
						Idoszak:    a.Idoszak,
						Hol:        fmt.Sprintf("%s, %s", a.Varos, a.Terulet),
						Megnevezes: z.Nev,
						Vegpont:    fmt.Sprintf("%s, %s", z.Varos, z.Terulet),
					})
				}
				if x := hoszt.Find(a.Varos, a.Terulet_mod, num); x != nil {
					kiirasok = append(kiirasok, Kiiras{
						Tipus:      "HOSZT",
						Sorszam:    a.ID,
						Datum:      datum,
						Idoszak:    a.Idoszak,
						Hol:        fmt.Sprintf("%s, %s", a.Varos, a.Terulet),
						Megnevezes: x.Varos,
						Vegpont:    fmt.Sprintf("%s, %s", x.Varos, x.Terulet),
					})
				}
				if w := olt.Find(a.Varos, a.Terulet_mod, num); w != nil {
					kiirasok = append(kiirasok, Kiiras{
						Tipus:      "OLT",
						Sorszam:    a.ID,
						Datum:      datum,
						Idoszak:    a.Idoszak,
						Hol:        fmt.Sprintf("%s, %s", a.Varos, a.Terulet),
						Megnevezes: w.Nev,
						Vegpont:    fmt.Sprintf("%s, %s", w.Varos, w.Terulet),
					})
				}
			}
		}

		// dátum szerinti rendezés
		sort.SliceStable(kiirasok, func(i, j int) bool {
			return kiirasok[i].Datum.Before(kiirasok[j].Datum)
		})

		// fájlba írás
		seenIds := make(map[Kiiras]bool)

		// UTF-8 BOM kiírása
		if _, err := f.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
			logging.Fatal("Hiba BOM írásakor", "hiba", err)
		}

		writer := csv.NewWriter(f)
		defer writer.Flush()
		writer.Comma = ';'

		header := []string{"Típus", "Áramszünet Id", "Dátum", "Időpont", "Helyszín", "Megnevezés", "Végpont", "Modemek száma", "Végpontok száma"}
		if err := writer.Write(header); err != nil {
			logging.Fatal("Hiba a fejléc írásakor:", err)
		}

		for _, k := range kiirasok {

			logging.Logger.Debug("Kiírás", "tipus", k.Tipus, "sorszam", k.Sorszam, "datum", k.Datum, "idoszak", k.Idoszak, "hol", k.Hol, "megnevezes", k.Megnevezes, "vegpont", k.Vegpont, "modemek_szama", k.modemek_szama, "vegpontok_szama", k.vegpontok_szama)

			key := Kiiras{Tipus: k.Tipus, Megnevezes: k.Megnevezes, Datum: k.Datum}
			if seenIds[key] {
				continue // már kiírtuk ezt a Tipus+Dátum kombinációt
			}
			seenIds[key] = true

			record := []string{
				k.Tipus,
				k.Sorszam,
				k.Datum.Format("2006-01-02"),
				k.Idoszak,
				k.Hol,
				k.Megnevezes,
				k.Vegpont,
				strconv.Itoa(k.modemek_szama),
				strconv.Itoa(k.vegpontok_szama),
			}

			if err := writer.Write(record); err != nil {
				logging.Fatal("Kiírás hiba", "hiba", err)
			}
		}

		logging.Logger.Info("CSV fájl létrehozva", "fájlnév", "lehetseges_aramszunet.csv")
	}
}

func runCheck() {
	conf := config.NewConfig("./aram.yaml")
	node := internal.NewNodes("./nodeok.txt", conf)
	modem := internal.NewActiveModems("./activemodemlist.csv", conf)

	// alakítsuk map-pé a gyors kereséshez
	nodeMap := make(map[string]struct{})
	for _, n := range node.List {
		nodeMap[n.Node] = struct{}{}
	}

	modemMap := make(map[string]struct{})
	for _, m := range modem.List {
		modemMap[m.Node1] = struct{}{}
		if m.Node2 != "" {
			modemMap[m.Node2] = struct{}{}
		}
		if m.Node3 != "" {
			modemMap[m.Node3] = struct{}{}
		}
	}

	logging.Logger.Info("Eltérések:")

	// csak nodeokban
	for n := range nodeMap {
		if _, ok := modemMap[n]; !ok {
			logging.Logger.Info("Csak nodeokban:", "név", n)
		}
	}

	// csak modemekben
	for m := range modemMap {
		if _, ok := nodeMap[m]; !ok {
			logging.Logger.Info("Csak activemodem-ben:", "név", m)
		}
	}
}

func runCountCheck() {
	start := time.Now()
	conf := config.NewConfig("./aram.yaml")
	node := internal.NewNodes("./nodeok.txt", conf)
	modem := internal.NewActiveModems("./activemodemlist.csv", conf)

	logging.Logger.Info("Node modem darabszámok:")

	nodeNames := make([]string, len(node.List))
	for i, n := range node.List {
		nodeNames[i] = n.Node
	}
	slices.Sort(nodeNames)
	uniqueNodes := slices.Compact(nodeNames)

	// számláló map
	counts := make(map[string]int)
	for _, m := range modem.List {
		counts[m.Node1]++
		if m.Node2 != "" {
			counts[m.Node2]++
		}
		if m.Node3 != "" {
			counts[m.Node3]++
		}
	}

	for _, n := range uniqueNodes {
		logging.Logger.Info("Modem darabszám:", n, counts[n])
	}
	logging.Logger.Info("Ellenőrzés kész", "eltelt idő", time.Since(start))
}
