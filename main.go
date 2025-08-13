package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/MackoMici/aram/config"
	"github.com/MackoMici/aram/internal"
	"github.com/MackoMici/aram/logging"
)

func main() {
	// Parancssori kapcsoló: -debug
	debugMode := flag.Bool("debug", false, "Debug mód engedélyezése")
	flag.Parse()

	logging.Init(*debugMode)
	defer func() {
		if err := logging.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Logfájl bezárás hiba: %v\n", err)
		}
	}()
	logging.Logger.Info("Program elindult", "verzió", "v2.0.1")
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

		if a.TeljesTel {
			_, err := fmt.Fprintln(f, "A teljes település megáll:", a)
			if err != nil {
				logging.Fatal("Teljes település", "hiba", err)
			}
		} else {
			var (
				y *internal.Mux
				w *internal.Olt
				x *internal.Hoszt
				z *internal.Fejallomas
				v *internal.Node
			)
			for _, num := range a.Hazszamok {
				y = mux.Find(a.Varos, a.Terulet_mod, num)
				if y != nil {
					_, err := fmt.Fprintln(f, "MUX áramszünet miatt ellenőrizni:", a, "=>", y)
					if err != nil {
						logging.Fatal("Mux ellenőrzés", "hiba", err)
					}
				}
				v = node.Find(a.Varos, a.Terulet_mod, num)
				if v != nil {
					_, err := fmt.Fprintln(f, "Áramszünet miatt ellenőrizni:", a, "=>", v)
					if err != nil {
						logging.Fatal("Node ellenőrzés", "hiba", err)
					}
				}
				z = fej.Find(a.Varos, a.Terulet_mod, num)
				if z != nil {
					_, err := fmt.Fprintln(f, "Fejállomás áramszünet miatt ellenőrizni:", a, "=>", z)
					if err != nil {
						logging.Fatal("Fejállomás ellenőrzés", "hiba", err)
					}
				}
				x = hoszt.Find(a.Varos, a.Terulet_mod, num)
				if x != nil {
					_, err := fmt.Fprintln(f, "Hoszt áramszünet miatt ellenőrizni:", a, "=>", x)
					if err != nil {
						logging.Fatal("Hoszt", "hiba", err)
					}
				}
				w = olt.Find(a.Varos, a.Terulet_mod, num)
				if w != nil {
					_, err := fmt.Fprintln(f, "OLT áramszünet miatt ellenőrizni:", a, "=>", w)
					if err != nil {
						logging.Fatal("OLT ellenőrzés", "hiba", err)
					}
				}
			}
		}
	}
}
