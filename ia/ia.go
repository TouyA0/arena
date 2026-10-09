package ia

import (
	"robot/carte"
	"robot/protocole"
)

const Portee = 5

func Decider(etat protocole.Etat, c carte.Carte) (string, string) {
	moi := Pos{etat.Moi.X, etat.Moi.Y}
	ennemi := Pos{etat.Ennemi.X, etat.Ennemi.Y}

	directionTir, peutTirer := directionDeTir(c, moi, ennemi)
	if peutTirer {
		return "TIRE " + directionTir, "ennemi aligné"
	}

	coutTotal, premierPas := Distances(c, moi, ennemi, etat.Moi.VIE)

	meilleurCout := -1
	meilleureDirection := ""

	for caseTestee, coutCase := range coutTotal {
		if caseTestee == moi {
			continue
		}

		_, caseDeTir := directionDeTir(c, caseTestee, ennemi)
		if !caseDeTir {
			continue
		}

		if meilleurCout == -1 || coutCase < meilleurCout {
			meilleurCout = coutCase
			meilleureDirection = premierPas[caseTestee]
		}
	}

	if meilleureDirection == "" {
		return "ATTENDS", "aucune case de tir atteignable"
	}
	return "AVANCE " + meilleureDirection, "approche d'une case de tir"
}

func directionDeTir(c carte.Carte, caseDepart, cible Pos) (string, bool) {
	for _, direction := range directions {
		caseDuTir := caseDepart

		for distance := 1; distance <= Portee; distance++ {
			caseDuTir = voisin(caseDuTir, direction)

			if symbole(c, caseDuTir) == '#' {
				break
			}
			if caseDuTir == cible {
				return direction, true
			}
		}
	}
	return "", false
}