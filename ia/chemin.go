package ia

import "robot/carte"

func Distances(c carte.Carte, depart, ennemi Pos, pv int) (map[Pos]int, map[Pos]string) {
	coutTotal := map[Pos]int{depart: 0}
	premierPas := map[Pos]string{}
	casesAVisiter := []Pos{depart}
 
	for len(casesAVisiter) > 0 {
		caseActuelle := casesAVisiter[0]
		casesAVisiter = casesAVisiter[1:] // retire la case qu'on traite
 
		for _, direction := range directions {
			caseVoisine := voisin(caseActuelle, direction)
			if caseVoisine == ennemi {
				continue
			}
			coutVoisine := cout(c, caseVoisine, pv)
			if coutVoisine == Interdit {
				continue
			}
 
			nouveauCout := coutTotal[caseActuelle] + coutVoisine
			ancienCout, dejaNotee := coutTotal[caseVoisine]
			if dejaNotee && ancienCout <= nouveauCout {
				continue // on connaît déjà un chemin aussi bon
			}
 
			coutTotal[caseVoisine] = nouveauCout
			if caseActuelle == depart {
				premierPas[caseVoisine] = direction // on vient du robot
			} else {
				premierPas[caseVoisine] = premierPas[caseActuelle] // on hérite
			}
			casesAVisiter = append(casesAVisiter, caseVoisine)
		}
	}
	return coutTotal, premierPas
}
