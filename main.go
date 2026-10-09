package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"robot/carte"
	"robot/ia"
	"robot/protocole"
)

func lireOptions() (string, int) {
	chemin := flag.String("carte", "", "fichier .map")
	joueur := flag.Int("joueur", 0, "1 ou 2")
	flag.Parse()
	return *chemin, *joueur
}

func main() {
	cheminCarte, joueur := lireOptions()

	c, err := carte.Charger(cheminCarte)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "joueur %d, carte %q chargée\n", joueur, c.Nom)

	cerveau := ia.NouveauCerveau(c)
	lecteur := protocole.NouveauLecteur(os.Stdin)
	for {
		etat, err := protocole.LireBloc(lecteur)
		if err == io.EOF {
			return // fin de partie
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "bloc invalide :", err)
			protocole.Repondre("ATTENDS")
			continue
		}
		action, raison := ia.Decider(cerveau, etat)
		fmt.Fprintf(os.Stderr, "T%d | %s | %s\n", etat.Tour, action, raison)
		protocole.Repondre(action)
	}
}