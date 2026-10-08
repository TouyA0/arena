package tests

import (
	"testing"
	"robot/protocole"
)

func TestBlocValide(t *testing.T) {
	lignes := []string{"TOUR 12", "MOI 3 4 VIE 80", "ENNEMI 7 4 VIE 40"}

	e, err := protocole.Lire(lignes)
	if err != nil {
		t.Errorf("erreur inattendue : %v", err)
	}
	if e.Tour != 12 || e.Moi.X != 3 || e.Ennemi.VIE != 40 {
		t.Errorf("mauvaises valeurs : %+v", e)
	}
}

func TestBlocsAcceptes(t *testing.T) {
	blocs := [][]string{
		{"", "TOUR 12", "MOI 3 4 VIE 80", "ENNEMI 7 4 VIE 40"},
		{"TOUR   12", "MOI 3  4 VIE 80", "ENNEMI 7 4 VIE 40"},
		{"TOUR\t12", "MOI 3 4 VIE 80", "ENNEMI 7 4 VIE 40"},
		{"TOUR 12\r", "MOI 3 4 VIE 80\r", "ENNEMI 7 4 VIE 40\r"},
		{"TOUR 12", "BONUS 5", "MOI 3 4 VIE 80", "ENNEMI 7 4 VIE 40"},
		{"TOUR 12", "MOI 3 4 VIE 80 XX", "ENNEMI 7 4 VIE 40"},
	}

	for _, bloc := range blocs {
		if _, err := protocole.Lire(bloc); err != nil {
			t.Errorf("bloc %q refusé : %v", bloc, err)
		}
	}
}

func TestBlocsRefuses(t *testing.T) {
	blocs := [][]string{
		{"MOI 3 4 VIE 80", "ENNEMI 7 4 VIE 40"},
		{"TOUR 12", "TOUR 13", "MOI 3 4 VIE 80", "ENNEMI 7 4 VIE 40"},
		{"TOUR", "MOI 3 4 VIE 80", "ENNEMI 7 4 VIE 40"},
		{"TOUR abc", "MOI 3 4 VIE 80", "ENNEMI 7 4 VIE 40"},
		{"TOUR 12", "MOI 3 4", "ENNEMI 7 4 VIE 40"},
		{"TOUR 12", "MOI 3 abc VIE 80", "ENNEMI 7 4 VIE 40"},
		{"TOUR 12", "MOI 3 4 PV 80", "ENNEMI 7 4 VIE 40"},
	}

	for _, bloc := range blocs {
		if _, err := protocole.Lire(bloc); err == nil {
			t.Errorf("bloc %q accepté alors qu'il est invalide", bloc)
		}
	}
}