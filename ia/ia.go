package ia

import (
	"robot/carte"
	"robot/protocole"
)

const (
	Portee = 5 // dist max d'un tir
	ObjetProche = 3 // dist max pour considérer un objet comme proche
	SansLimite  = -1 // pas de cout max pour aller chercher un objet
	PatienceMax = 3  // nombre de tours max a attendre l'ennemi

	PVMax = 100
	GainSoin = 30
	MunSuffisantes = 10
)

type Cerveau struct {
	carte        carte.Carte
	ennemiAvant  Pos
	dejaJoue     bool
	toursAttente int
}

func NouveauCerveau(c carte.Carte) *Cerveau {
	return &Cerveau{carte: c}
}

func Decider(cerveau *Cerveau, etat protocole.Etat) (string, string) {
	c := cerveau.carte
	moi := Pos{etat.Moi.X, etat.Moi.Y}
	ennemi := Pos{etat.Ennemi.X, etat.Ennemi.Y}

	ennemiABouge := cerveau.dejaJoue && ennemi != cerveau.ennemiAvant
	cerveau.ennemiAvant = ennemi
	cerveau.dejaJoue = true

	dejaAttendu := cerveau.toursAttente
	cerveau.toursAttente = 0

	directionTir, peutTirer := directionDeTir(c, moi, ennemi)
	aDesMunitions := etat.Moi.MUN != 0
	if peutTirer && aDesMunitions {
		return "TIRE " + directionTir, "ennemi aligné"
	}

	coutTotal, premierPas := Distances(c, moi, ennemi, etat.Moi.VIE)

	if ennemiABouge && aDesMunitions {
		directionAnticipee, touche := tirAnticipe(c, moi, ennemi, coutTotal)
		if touche {
			return "TIRE " + directionAnticipee, "tir anticipé sur son prochain pas"
		}
	}

	if etat.Moi.VIE <= SeuilPVBas {
		direction := versObjet(etat.Objets, "SOIN", moi, coutTotal, premierPas, SansLimite)
		if direction != "" {
			return "AVANCE " + direction, "PV bas, va au soin"
		}
	}

	if !aDesMunitions {
		direction := versObjet(etat.Objets, "MUNITION", moi, coutTotal, premierPas, SansLimite)
		if direction != "" {
			return "AVANCE " + direction, "plus de munitions, va recharger"
		}
	}

	direction := versObjet(objetsUtiles(etat), "", moi, coutTotal, premierPas, ObjetProche)
	if direction != "" {
		return "AVANCE " + direction, "ramasse un objet proche"
	}

	direction = versCaseDeTir(c, moi, ennemi, coutTotal, premierPas)
	if direction != "" {
		prochaineCase := voisin(moi, direction)
		dangereuse := exposee(c, prochaineCase, ennemi, coutTotal)
		if dangereuse && ennemiABouge && dejaAttendu < PatienceMax {
			cerveau.toursAttente = dejaAttendu + 1
			return "ATTENDS", "attend qu'il entre dans ma ligne de tir"
		}
		return "AVANCE " + direction, "approche d'une case de tir"
	}

	return "ATTENDS", "rien d'utile à faire"
}

func casesProbablesEnnemi(ennemi Pos, coutTotal map[Pos]int) []Pos {
	coutMin := -1
	for _, direction := range directions {
		coutCase, atteignable := coutTotal[voisin(ennemi, direction)]
		if atteignable && (coutMin == -1 || coutCase < coutMin) {
			coutMin = coutCase
		}
	}
	if coutMin <= 0 {
		return nil
	}

	var cases []Pos
	for _, direction := range directions {
		caseVoisine := voisin(ennemi, direction)
		coutCase, atteignable := coutTotal[caseVoisine]
		if atteignable && coutCase == coutMin {
			cases = append(cases, caseVoisine)
		}
	}
	return cases
}

func tirAnticipe(c carte.Carte, moi, ennemi Pos, coutTotal map[Pos]int) (string, bool) {
	for _, caseFuture := range casesProbablesEnnemi(ennemi, coutTotal) {
		directionTir, touche := directionDeTir(c, moi, caseFuture)
		if touche {
			return directionTir, true
		}
	}
	return "", false
}

func exposee(c carte.Carte, caseTestee, ennemi Pos, coutTotal map[Pos]int) bool {
	positionsEnnemi := append([]Pos{ennemi}, casesProbablesEnnemi(ennemi, coutTotal)...)
	for _, positionEnnemi := range positionsEnnemi {
		_, touche := directionDeTir(c, positionEnnemi, caseTestee)
		if touche {
			return true
		}
	}
	return false
}

func objetsUtiles(etat protocole.Etat) []protocole.Objet {
	var utiles []protocole.Objet

	for _, objet := range etat.Objets {
		switch objet.Type {
		case "SOIN":
			if etat.Moi.VIE > PVMax-GainSoin {
				continue
			}
		case "MUNITION":
			if etat.Moi.MUN == protocole.MunIllimitees || etat.Moi.MUN >= MunSuffisantes {
				continue
			}
		}
		utiles = append(utiles, objet)
	}
	return utiles
}

func versObjet(objets []protocole.Objet, typeVoulu string, moi Pos,
	coutTotal map[Pos]int, premierPas map[Pos]string, coutMax int) string {

	meilleurCout := -1
	meilleureDirection := ""

	for _, objet := range objets {
		if typeVoulu != "" && objet.Type != typeVoulu {
			continue
		}
		caseObjet := Pos{objet.X, objet.Y}
		coutObjet, atteignable := coutTotal[caseObjet]
		if !atteignable || caseObjet == moi {
			continue
		}
		if coutMax != SansLimite && coutObjet > coutMax {
			continue // trop loin
		}
		if meilleurCout == -1 || coutObjet < meilleurCout {
			meilleurCout = coutObjet
			meilleureDirection = premierPas[caseObjet]
		}
	}
	return meilleureDirection
}

func versCaseDeTir(c carte.Carte, moi, ennemi Pos,
	coutTotal map[Pos]int, premierPas map[Pos]string) string {

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
	return meilleureDirection
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