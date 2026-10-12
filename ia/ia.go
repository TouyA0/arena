package ia

import (
	"robot/carte"
	"robot/protocole"
)

const (
	Portee         = 5  // dist max d'un tir
	ObjetProche    = 3  // dist max pour considérer un objet comme proche
	SansLimite     = -1 // pas de cout max pour aller chercher un objet
	PatienceMax    = 3  // nombre de tours max a attendre l'ennemi
	DetourBouclier = 10

	PVMax          = 100
	GainSoin       = 30
	MunSuffisantes = 10

	RatesMax          = 2  // tirs anticipés ratés d'affilée avant de faire une pause
	PauseAnticipation = 10 // nombre de tours sans tir anticipé après trop de ratés

	Vision       = 4  // distance à laquelle on voit l'ennemi et les objets (v3)
	OubliEnnemi  = 6  // tours sans voir l'ennemi avant de lancer le radar
	DelaiRadar   = 10 // tours minimum entre deux radars
	FraicheurMax = 2
)

type Cerveau struct {
	carte        carte.Carte
	dejaJoue     bool
	ennemiConnu  Pos
	vieEnnemi    int
	vuAuTour     int
	objetsConnus []protocole.Objet
	dernierRadar int
	toursAttente int
	plusProche   int
	caseVisee    Pos
	aVise        bool
	rates        int
	pause        int
}

func NouveauCerveau(c carte.Carte) *Cerveau {
	return &Cerveau{carte: c, plusProche: -1, vieEnnemi: PVMax, vuAuTour: -1, dernierRadar: -DelaiRadar}
}

func Decider(cerveau *Cerveau, etat protocole.Etat) (string, string) {
	c := cerveau.carte
	moi := Pos{etat.Moi.X, etat.Moi.Y}

	if !cerveau.dejaJoue {
		cerveau.ennemiConnu = departEnnemi(c, moi)
		cerveau.dejaJoue = true
	}
	ennemiAvant := cerveau.ennemiConnu
	vuAuTourAvant := cerveau.vuAuTour
	memoriser(cerveau, etat, moi)
	ennemi := cerveau.ennemiConnu
	objets := cerveau.objetsConnus
	vuAuTourPrecedent := vuAuTourAvant == etat.Tour-1
	ilEstImmobile := etat.EnnemiVu && vuAuTourPrecedent && ennemi == ennemiAvant
	ilBouge := !ilEstImmobile
	infoFraiche := etat.Tour-cerveau.vuAuTour <= FraicheurMax

	if !etat.EnnemiVu {
		coutAvant, _ := Distances(c, moi, ennemi, etat.Moi.VIE)
		ennemi = positionPrevue(ennemi, moi, coutAvant)
		cerveau.ennemiConnu = ennemi
	}

	dejaAttendu := cerveau.toursAttente
	cerveau.toursAttente = 0
	plusProcheAvant := cerveau.plusProche
	cerveau.plusProche = -1

	if cerveau.aVise {
		if ennemi == cerveau.caseVisee {
			cerveau.rates = 0
		} else {
			cerveau.rates++
		}
		cerveau.aVise = false
	}
	if cerveau.rates >= RatesMax {
		cerveau.pause = PauseAnticipation
		cerveau.rates = 0
	}
	peutAnticiper := cerveau.pause == 0
	if cerveau.pause > 0 {
		cerveau.pause--
	}

	directionTir, peutTirer := directionDeTir(c, moi, ennemi)
	aDesMunitions := etat.Moi.MUN != 0
	if etat.EnnemiVu && peutTirer && aDesMunitions {
		return "TIRE " + directionTir, "ennemi aligné"
	}
	if !etat.EnnemiVu && infoFraiche && peutTirer && aDesMunitions {
		cerveau.caseVisee = ennemi
		cerveau.aVise = true
		return "TIRE " + directionTir, "tir sur sa position prévue"
	}

	if radarUtile(cerveau, etat, moi) {
		cerveau.dernierRadar = etat.Tour
		return "RADAR", "ennemi perdu de vue"
	}

	coutTotal, premierPas := Distances(c, moi, ennemi, etat.Moi.VIE)
	coutEnnemi, _ := Distances(c, ennemi, moi, cerveau.vieEnnemi)

	distance := distanceEnnemi(ennemi, coutTotal)
	seRapproche := distance != -1 && (plusProcheAvant == -1 || distance < plusProcheAvant)

	choix, raison := "", ""
	if etat.Moi.BOUCLIER == 0 {
		boucliers := bouclierDeMonCote(objets, coutTotal, coutEnnemi)
		choix = versObjet(boucliers, "BOUCLIER", moi, coutTotal, premierPas, DetourBouclier)
		raison = "va chercher le bouclier de son côté"
	}

	if choix == "" && ilBouge && infoFraiche && aDesMunitions && peutAnticiper {
		directionAnticipee, caseVisee, touche := tirAnticipe(c, moi, ennemi, coutTotal)
		if touche {
			cerveau.caseVisee = caseVisee
			cerveau.aVise = true
			return "TIRE " + directionAnticipee, "tir anticipé sur son prochain pas"
		}
	}

	if choix == "" && etat.Moi.VIE <= SeuilPVBas {
		choix = versObjet(objets, "SOIN", moi, coutTotal, premierPas, SansLimite)
		raison = "PV bas, va au soin"
	}
	if choix == "" && !aDesMunitions {
		choix = versObjet(objets, "MUNITION", moi, coutTotal, premierPas, SansLimite)
		raison = "plus de munitions, va recharger"
	}
	if choix == "" && etat.Moi.BOUCLIER == 0 {
		choix = versObjet(objets, "BOUCLIER", moi, coutTotal, premierPas, DetourBouclier)
		raison = "va chercher un bouclier"
	}
	if choix == "" {
		choix = versObjet(objetsUtiles(objets, etat.Moi), "", moi, coutTotal, premierPas, ObjetProche)
		raison = "ramasse un objet proche"
	}
	if choix == "" {
		choix = versCaseDeTir(c, moi, ennemi, coutTotal, premierPas)
		raison = "approche d'une case de tir"
	}
	if choix == "" {
		return "ATTENDS", "rien d'utile à faire"
	}

	dangereuse := exposee(c, voisin(moi, choix), ennemi, coutTotal)
	if dangereuse && ilBouge && (seRapproche || dejaAttendu < PatienceMax) {
		cerveau.toursAttente = dejaAttendu
		cerveau.plusProche = plusProcheAvant
		if seRapproche {
			cerveau.plusProche = distance
		} else {
			cerveau.toursAttente = dejaAttendu + 1
		}
		if !etat.EnnemiVu {
			cerveau.dernierRadar = etat.Tour
			return "RADAR", "attend en le surveillant au radar"
		}
		return "ATTENDS", "attend qu'il entre dans ma ligne de tir"
	}
	return "AVANCE " + choix, raison
}

func memoriser(cerveau *Cerveau, etat protocole.Etat, moi Pos) {
	if etat.EnnemiVu {
		cerveau.ennemiConnu = Pos{etat.Ennemi.X, etat.Ennemi.Y}
		cerveau.vieEnnemi = etat.Ennemi.VIE
		cerveau.vuAuTour = etat.Tour
	}

	var gardes []protocole.Objet
	if !vueComplete(cerveau, etat, moi) {
		for _, objet := range cerveau.objetsConnus {
			if ecart(moi, Pos{objet.X, objet.Y}) > Vision {
				gardes = append(gardes, objet)
			}
		}
	}
	cerveau.objetsConnus = append(gardes, etat.Objets...)
}

func vueComplete(cerveau *Cerveau, etat protocole.Etat, moi Pos) bool {
	if cerveau.dernierRadar == etat.Tour-1 {
		return true
	}
	if etat.EnnemiVu && ecart(moi, Pos{etat.Ennemi.X, etat.Ennemi.Y}) > Vision {
		return true
	}
	for _, objet := range etat.Objets {
		if ecart(moi, Pos{objet.X, objet.Y}) > Vision {
			return true
		}
	}
	return false
}

func radarUtile(cerveau *Cerveau, etat protocole.Etat, moi Pos) bool {
	if etat.EnnemiVu || etat.Tour-cerveau.dernierRadar < DelaiRadar {
		return false
	}
	pasVuDepuisLongtemps := etat.Tour-cerveau.vuAuTour >= OubliEnnemi
	plusLaOuJeCroyais := ecart(moi, cerveau.ennemiConnu) <= Vision
	return pasVuDepuisLongtemps || plusLaOuJeCroyais
}

func positionPrevue(ennemi, moi Pos, coutTotal map[Pos]int) Pos {
	for _, caseFuture := range casesProbablesEnnemi(ennemi, coutTotal) {
		if ecart(caseFuture, moi) > Vision {
			return caseFuture
		}
	}
	return ennemi
}

func departEnnemi(c carte.Carte, moi Pos) Pos {
	depart1 := Pos{c.Depart1.X, c.Depart1.Y}
	if moi == depart1 {
		return Pos{c.Depart2.X, c.Depart2.Y}
	}
	return depart1
}

func ecart(a, b Pos) int {
	dx := a.X - b.X
	if dx < 0 {
		dx = -dx
	}
	dy := a.Y - b.Y
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

func distanceEnnemi(ennemi Pos, coutTotal map[Pos]int) int {
	coutMin := -1
	for _, direction := range directions {
		coutCase, atteignable := coutTotal[voisin(ennemi, direction)]
		if atteignable && (coutMin == -1 || coutCase < coutMin) {
			coutMin = coutCase
		}
	}
	return coutMin
}

func casesProbablesEnnemi(ennemi Pos, coutTotal map[Pos]int) []Pos {
	coutMin := distanceEnnemi(ennemi, coutTotal)
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

func tirAnticipe(c carte.Carte, moi, ennemi Pos, coutTotal map[Pos]int) (string, Pos, bool) {
	for _, caseFuture := range casesProbablesEnnemi(ennemi, coutTotal) {
		directionTir, touche := directionDeTir(c, moi, caseFuture)
		if touche {
			return directionTir, caseFuture, true
		}
	}
	return "", Pos{}, false
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

func bouclierDeMonCote(objets []protocole.Objet, coutTotal, coutEnnemi map[Pos]int) []protocole.Objet {
	var boucliers []protocole.Objet

	for _, objet := range objets {
		if objet.Type != "BOUCLIER" {
			continue
		}
		caseObjet := Pos{objet.X, objet.Y}
		monCout, jePeux := coutTotal[caseObjet]
		sonCout, ilPeut := coutEnnemi[caseObjet]
		if jePeux && (!ilPeut || monCout < sonCout) {
			boucliers = append(boucliers, objet)
		}
	}
	return boucliers
}

func objetsUtiles(objets []protocole.Objet, robot protocole.Robot) []protocole.Objet {
	var utiles []protocole.Objet

	for _, objet := range objets {
		switch objet.Type {
		case "SOIN":
			if robot.VIE > PVMax-GainSoin {
				continue
			}
		case "MUNITION":
			if robot.MUN == protocole.MunIllimitees || robot.MUN >= MunSuffisantes {
				continue
			}
		case "BOUCLIER", "TRESOR":
		default:
			continue
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