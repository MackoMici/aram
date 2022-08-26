package main

import (
	"fmt"

	"github.com/MackoMici/aram/internal"
)

func main() {
	asz := internal.NewAramSzunets("./aramszunet.txt")
	am := internal.NewActiveModems("./activemodemlist.csv")

	for _, a := range asz.List {

		if v := am.Vegpont(a.Vegpont()); v != nil {
			fmt.Printf("Áramszünet miatt ellenőrizni: %#s => %#s\n", a, v)
		}
	}
}
