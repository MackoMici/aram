package main

import (
	"log"

	"github.com/MackoMici/aram/internal"
)

func main() {
	asz := internal.NewAramSzunets("./aramszunet.txt")
	am := internal.NewActiveModems("./activemodemlist.csv")

	for _, a := range asz.List {
		if v := am.Vegpont(a.Vegpont()); v != nil {
			log.Printf("Végpont létezik %#v => %#v\n\n", a, v)
		}
	}
}
