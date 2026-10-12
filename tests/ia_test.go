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
		{"ennemi aligné à l'est", "cartes/arene.map",
			[]string{"TOUR 1", "MOI 1 1 VIE 100 MUN 10", "ENNEMI 5 1 VIE 100"}, "TIRE E"},
		{"ennemi aligné au sud", "cartes/arene.map",
			[]string{"TOUR 1", "MOI 1 1 VIE 100 MUN 10", "ENNEMI 1 4 VIE 100"}, "TIRE S"},
		{"ennemi aligné au nord", "cartes/arene.map",
			[]string{"TOUR 1", "MOI 9 5 VIE 100 MUN 10", "ENNEMI 9 1 VIE 100"}, "TIRE N"},
		{"brouillard, va vers son départ", "cartes/arene.map",
			[]string{"TOUR 1", "MOI 1 1 VIE 100 MUN 10 BOUCLIER 0"}, "AVANCE E"},
		{"pas vu depuis 6 tours, radar", "cartes/arene.map",
			[]string{"TOUR 6", "MOI 1 1 VIE 100 MUN 10 BOUCLIER 0"}, "RADAR"},
		{"objet inconnu ignoré", "cartes/arene.map",
			[]string{"TOUR 1", "MOI 1 1 VIE 100 MUN 10 BOUCLIER 0", "ENNEMI 9 5 VIE 100", "OBJET 2 1 BIDULE"}, "AVANCE S"},
		{"ennemi collé à nous", "cartes/arene.map",
			[]string{"TOUR 1", "MOI 1 1 VIE 100 MUN 10", "ENNEMI 2 1 VIE 100"}, "TIRE E"},
		{"première vue, la case suivante est dans sa ligne", "cartes/arene.map",
			[]string{"TOUR 1", "MOI 1 1 VIE 100 MUN 10", "ENNEMI 8 1 VIE 100"}, "ATTENDS"},
		{"mur entre nous, attend qu'il sorte", "cartes/arene.map",
			[]string{"TOUR 1", "MOI 3 2 VIE 100 MUN 10", "ENNEMI 7 2 VIE 100"}, "ATTENDS"},
		{"plus de munitions", "cartes/arene.map",
			[]string{"TOUR 1", "MOI 1 1 VIE 100 MUN 0", "ENNEMI 5 1 VIE 100", "OBJET 1 3 MUNITION"}, "AVANCE S"},
		{"PV bas, va au soin", "cartes/arene.map",
			[]string{"TOUR 1", "MOI 1 1 VIE 40 MUN 5", "ENNEMI 9 5 VIE 100", "OBJET 1 3 SOIN"}, "AVANCE S"},
		{"va chercher un bouclier", "cartes/brume.map",
			[]string{"TOUR 1", "MOI 3 5 VIE 100 MUN 10", "ENNEMI 19 9 VIE 100", "OBJET 6 5 BOUCLIER"}, "AVANCE E"},
		{"protocole v1 sans MUN", "cartes/arene.map",
			[]string{"TOUR 1", "MOI 1 1 VIE 100", "ENNEMI 5 1 VIE 100"}, "TIRE E"},
		{"tir sur le pont (notre carte)", "../cartes/pont.map",
			[]string{"TOUR 1", "MOI 4 4 VIE 100 MUN 10", "ENNEMI 8 4 VIE 100"}, "TIRE E"},
		{"contourne le piège pour aller au soin (notre carte)", "../cartes/oasis.map",
			[]string{"TOUR 1", "MOI 3 4 VIE 40 MUN 5", "ENNEMI 13 9 VIE 100", "OBJET 7 4 SOIN", "OBJET 7 6 SOIN"}, "AVANCE N"},
		{"bouclier de son côté (notre carte)", "../cartes/oasis.map",
			[]string{"TOUR 1", "MOI 1 3 VIE 100 MUN 10", "ENNEMI 13 9 VIE 100", "OBJET 2 5 BOUCLIER", "OBJET 12 5 BOUCLIER"}, "AVANCE S"},
		{"ennemi hors de la carte", "cartes/arene.map",
			[]string{"TOUR 1", "MOI 1 1 VIE 100 MUN 10", "ENNEMI 500 -3 VIE 100"}, "ATTENDS"},
	}

	for _, c := range cas {
		laCarte, err := carte.Charger(c.carte)
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