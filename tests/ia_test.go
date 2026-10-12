package tests

import (
	"testing"

	"robot/carte"
	"robot/ia"
	"robot/protocole"
)

func TestIA(t *testing.T) {
	cas := []struct {
		nom     string
		carte   string
		bloc    []string
		attendu string
	}{
		{"ennemi aligné à l'est", "arene",
			[]string{"TOUR 1", "MOI 1 1 VIE 100 MUN 10", "ENNEMI 5 1 VIE 100"}, "TIRE E"},
		{"ennemi aligné au sud", "arene",
			[]string{"TOUR 1", "MOI 1 1 VIE 100 MUN 10", "ENNEMI 1 4 VIE 100"}, "TIRE S"},
		{"ennemi aligné au nord", "arene",
			[]string{"TOUR 1", "MOI 9 5 VIE 100 MUN 10", "ENNEMI 9 1 VIE 100"}, "TIRE N"},
		{"ennemi collé à nous", "arene",
			[]string{"TOUR 1", "MOI 1 1 VIE 100 MUN 10", "ENNEMI 2 1 VIE 100"}, "TIRE E"},
		{"trop loin pour tirer", "arene",
			[]string{"TOUR 1", "MOI 1 1 VIE 100 MUN 10", "ENNEMI 8 1 VIE 100"}, "AVANCE E"},
		{"mur entre nous", "arene",
			[]string{"TOUR 1", "MOI 3 2 VIE 100 MUN 10", "ENNEMI 7 2 VIE 100"}, "AVANCE N"},
		{"plus de munitions", "arene",
			[]string{"TOUR 1", "MOI 1 1 VIE 100 MUN 0", "ENNEMI 5 1 VIE 100", "OBJET 1 3 MUNITION"}, "AVANCE S"},
		{"PV bas, va au soin", "arene",
			[]string{"TOUR 1", "MOI 1 1 VIE 40 MUN 5", "ENNEMI 9 5 VIE 100", "OBJET 1 3 SOIN"}, "AVANCE S"},
		{"va chercher un bouclier", "brume",
			[]string{"TOUR 1", "MOI 3 5 VIE 100 MUN 10", "ENNEMI 19 9 VIE 100", "OBJET 6 5 BOUCLIER"}, "AVANCE E"},
		{"protocole v1 sans MUN", "arene",
			[]string{"TOUR 1", "MOI 1 1 VIE 100", "ENNEMI 5 1 VIE 100"}, "TIRE E"},
		{"ennemi hors de la carte", "arene",
			[]string{"TOUR 1", "MOI 1 1 VIE 100 MUN 10", "ENNEMI 500 -3 VIE 100"}, "ATTENDS"},
	}

	for _, c := range cas {
		laCarte, err := carte.Charger("cartes/" + c.carte + ".map")
		if err != nil {
			t.Fatalf("%s : carte illisible : %v", c.nom, err)
		}
		etat, err := protocole.Lire(c.bloc)
		if err != nil {
			t.Fatalf("%s : bloc illisible : %v", c.nom, err)
		}

		action, raison := ia.Decider(ia.NouveauCerveau(laCarte), etat)
		if action != c.attendu {
			t.Errorf("%s : %s (%s), attendu %s", c.nom, action, raison, c.attendu)
		}
	}
}