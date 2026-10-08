package main

import (
	"flag"
	"fmt"
	"os"

	"robot/carte"
)

type Robot struct {
	NOM string
	X, Y int
	VIE int
}

func main() {
	chemin := flag.String("carte", "", "fichier .map")
	joueur := flag.Int("joueur", 0, "1 ou 2")
	flag.Parse()

	c, err := carte.Charger(*chemin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "joueur %d, carte %q chargée\n", *joueur, c.Nom)
}