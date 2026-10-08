package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	
	"robot/carte"
	"robot/ia"
)

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
	
	entree := bufio.NewScanner(os.Stdin)
	for entree.Scan() {
		if entree.Text() == "FIN" {
			fmt.Println(ia.Decider())
		}
	}
}