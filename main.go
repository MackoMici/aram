package main

import (
	"fmt"
	"log"
	"os"

	"github.com/MackoMici/aram/internal"
	"github.com/MackoMici/aram/config"
)

func main() {
	conf := config.NewConfig("./aram.yaml")
	asz := internal.NewAramSzunets("./aramszunet.txt", conf)
	am := internal.NewActiveModems("./activemodemlist.csv", conf)

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
		} else if v := am.Vegpont(a.Vegpont()); v != nil {
			_, err := fmt.Fprintln(f, "Áramszünet miatt ellenőrizni:", a, "=>", v)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			_, err := fmt.Fprintln(f, "Nem találtam egyezést:", a.ID, a.Datum, "-", a.Varos, a.Terulet_mod)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
