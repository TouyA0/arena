package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"robot/carte"
	"robot/ia"
	"robot/journal"
	"robot/protocole"
)

func lireOptions() (string, int) {
	chemin := flag.String("carte", "", "fichier .map")
	joueur := flag.Int("joueur", 0, "1 ou 2")
	flag.Parse()
	return *chemin, *joueur
}

func lancer() int {
	cheminCarte, joueur := lireOptions()

	c, err := carte.Charger(cheminCarte)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	j, err := journal.Ouvrir("partie.log")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	defer journal.Fermer(j)
	journal.Debut(j, c.Nom, joueur)

	cerveau := ia.NouveauCerveau(c)
	lecteur := protocole.NouveauLecteur(os.Stdin)
	for {
		etat, err := protocole.LireBloc(lecteur)
		if err == io.EOF {
			return 0
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "bloc invalide :", err)
			journal.BlocInvalide(j, err)
			protocole.Repondre("ATTENDS")
			continue
		}
		action, raison := ia.Decider(cerveau, etat)
		fmt.Fprintf(os.Stderr, "T%d | %s | %s\n", etat.Tour, action, raison)
		journal.Tour(j, etat, action, raison)
		protocole.Repondre(action)
	}
}

func main() {
	os.Exit(lancer())
}
