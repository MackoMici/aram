package main

import (
	"fmt"
	"log"
	"os"

	"github.com/MackoMici/aram/config"
	"github.com/MackoMici/aram/internal"
)

func main() {
	conf := config.NewConfig("./aram.yaml")
	asz := internal.NewAramSzunets("./aramszunet.txt", conf)
	node := internal.NewNodes("./nodeok.txt", conf)
	fej := internal.NewFejallomasok("./fejallomas.txt", conf)
	hoszt := internal.NewHoszts("./hoszt.txt", conf)
	mux := internal.NewMuxs("./mux.txt", conf)
	olt := internal.NewOlts("./olt.txt", conf)

	f, err := os.Create("lehetseges_aramszunet.txt")

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	for _, a := range asz.List {

		if a.TeljesTel {
			_, err := fmt.Fprintln(f, "A teljes település megáll:", a)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			v := node.Vegpont(a.Vegpont())
			z := fej.Vegpont(a.Vegpont())
			x := hoszt.Vegpont(a.Vegpont())
			y := mux.Vegpont(a.Vegpont())
                        w := olt.Vegpont(a.Vegpont())
			if v != nil {
				_, err := fmt.Fprintln(f, "Áramszünet miatt ellenőrizni:", a, "=>", v)
				if err != nil {
					log.Fatal(err)
				}
			}
			if z != nil {
				_, err := fmt.Fprintln(f, "Fejállomás áramszünet miatt ellenőrizni:", a, "=>", z)
				if err != nil {
					log.Fatal(err)
				}
			}
			if x != nil {
				_, err := fmt.Fprintln(f, "Hoszt áramszünet miatt ellenőrizni:", a, "=>", x)
				if err != nil {
					log.Fatal(err)
				}
			}
			if y != nil {
				_, err := fmt.Fprintln(f, "MUX áramszünet miatt ellenőrizni:", a, "=>", y)
				if err != nil {
					log.Fatal(err)
				}
			}
			if w != nil {
				_, err := fmt.Fprintln(f, "OLT áramszünet miatt ellenőrizni:", a, "=>", w)
				if err != nil {
					log.Fatal(err)
				}
			}
			if v == nil && z == nil && x == nil && y == nil && w == nil {
				_, err := fmt.Fprintln(f, "Nem találtam node-ot az utcában:", a.ID, a.Datum, "-", a.Varos, a.Terulet)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
