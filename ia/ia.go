package ia

import (
	"robot/carte"
	"robot/protocole"
)

// Réglages de la stratégie.
const (
	// Tir et déplacements
	Portee         = 5  // un tir va jusqu'à 5 cases
	ObjetProche    = 3  // on fait un détour pour un objet s'il est à 3 pas ou moins
	SansLimite     = -1 // pas de distance maximale pour aller chercher un objet
	PatienceMax    = 3  // on attend au maximum 3 tours que l'ennemi vienne à nous
	DetourBouclier = 10 // un bouclier vaut un détour jusqu'à 10 pas

	// Objets
	PVMax          = 100 // PV au début de la partie, et maximum possible
	GainSoin       = 30  // PV rendus par un soin
	MunSuffisantes = 10  // au-delà, ramasser des munitions ne sert à rien

	// Tir anticipé
	RatesMax          = 2  // après 2 tirs anticipés ratés de suite, on fait une pause
	PauseAnticipation = 10 // la pause dure 10 tours

	// Brouillard
	Vision       = 4  // on voit l'ennemi et les objets à 4 cases ou moins
	OubliEnnemi  = 6  // après 6 tours sans le voir, on lance le radar
	DelaiRadar   = 10 // au moins 10 tours entre deux radars
	FraicheurMax = 2  // on tire sur sa position devinée seulement si on l'a vu il y a 2 tours ou moins
)

// Cerveau garde la mémoire du robot d'un tour à l'autre : la carte, ce qu'on
// sait de l'ennemi et des objets, l'attente en cours et les tirs anticipés.
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

// NouveauCerveau prépare un cerveau vide pour une partie sur la carte c.
func NouveauCerveau(c carte.Carte) *Cerveau {
	return &Cerveau{carte: c, plusProche: -1, vieEnnemi: PVMax, vuAuTour: -1, dernierRadar: -DelaiRadar}
}

// Decider choisit l'action du tour à partir de l'état reçu et de la mémoire,
// et renvoie l'action à envoyer à l'arbitre et la raison écrite dans le journal.
func Decider(cerveau *Cerveau, etat protocole.Etat) (string, string) {
	c := cerveau.carte
	moi := Pos{etat.Moi.X, etat.Moi.Y}

	// 1. Mémoire : on met à jour ce qu'on sait de l'ennemi et des objets.
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

	// 2. Si on ne le voit pas, on devine qu'il a avancé d'un pas vers nous.
	if !etat.EnnemiVu {
		coutAvant, _ := Distances(c, moi, ennemi, etat.Moi.VIE)
		ennemi = positionPrevue(ennemi, moi, coutAvant)
		cerveau.ennemiConnu = ennemi
	}

	// 3. On reprend l'attente et les tirs anticipés du tour d'avant.
	dejaAttendu := cerveau.toursAttente
	cerveau.toursAttente = 0
	plusProcheAvant := cerveau.plusProche
	cerveau.plusProche = -1

	if cerveau.aVise && etat.EnnemiVu {
		if ennemi == cerveau.caseVisee {
			cerveau.rates = 0
		} else {
			cerveau.rates++
		}
	}
	cerveau.aVise = false
	if cerveau.rates >= RatesMax {
		cerveau.pause = PauseAnticipation
		cerveau.rates = 0
	}
	peutAnticiper := cerveau.pause == 0
	if cerveau.pause > 0 {
		cerveau.pause--
	}

	// 4. Tirer s'il est aligné (vu, ou à sa position devinée).
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

	// 5. Radar si on l'a perdu depuis trop longtemps.
	if radarUtile(cerveau, etat, moi) {
		cerveau.dernierRadar = etat.Tour
		return "RADAR", "ennemi perdu de vue"
	}

	// 6. Choisir où aller, par priorité : bouclier, tir anticipé, soin,
	//    munitions, objets, puis une case d'où tirer sur lui.
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

	// 7. Ne pas entrer dans sa ligne de tir : on attend qu'il entre dans la nôtre.
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

// memoriser met à jour ce qu'on sait de l'ennemi et des objets avec l'état reçu.
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

// vueComplete dit si l'arbitre nous a tout montré (v1, v2, ou juste après un radar).
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

// radarUtile dit s'il faut lancer le radar parce qu'on a perdu l'ennemi.
func radarUtile(cerveau *Cerveau, etat protocole.Etat, moi Pos) bool {
	if etat.EnnemiVu || etat.Tour-cerveau.dernierRadar < DelaiRadar {
		return false
	}
	pasVuDepuisLongtemps := etat.Tour-cerveau.vuAuTour >= OubliEnnemi
	plusLaOuJeCroyais := ecart(moi, cerveau.ennemiConnu) <= Vision
	return pasVuDepuisLongtemps || plusLaOuJeCroyais
}

// positionPrevue avance l'ennemi d'un pas vers nous, sans le mettre à moins de 5 cases.
func positionPrevue(ennemi, moi Pos, coutTotal map[Pos]int) Pos {
	for _, caseFuture := range casesProbablesEnnemi(ennemi, coutTotal) {
		if ecart(caseFuture, moi) > Vision {
			return caseFuture
		}
	}
	return ennemi
}

// departEnnemi renvoie la case de départ qui n'est pas la nôtre.
func departEnnemi(c carte.Carte, moi Pos) Pos {
	depart1 := Pos{c.Depart1.X, c.Depart1.Y}
	if moi == depart1 {
		return Pos{c.Depart2.X, c.Depart2.Y}
	}
	return depart1
}

// ecart renvoie la distance |dx| + |dy| entre deux cases, sans tenir compte des murs.
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

// distanceEnnemi renvoie le coût pour aller à côté de l'ennemi, ou -1 si impossible.
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

// casesProbablesEnnemi renvoie les cases où l'ennemi ira s'il vient vers nous.
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

// tirAnticipe cherche une direction de tir qui touche l'ennemi sur son prochain pas.
func tirAnticipe(c carte.Carte, moi, ennemi Pos, coutTotal map[Pos]int) (string, Pos, bool) {
	for _, caseFuture := range casesProbablesEnnemi(ennemi, coutTotal) {
		directionTir, touche := directionDeTir(c, moi, caseFuture)
		if touche {
			return directionTir, caseFuture, true
		}
	}
	return "", Pos{}, false
}

// exposee dit si l'ennemi peut tirer sur cette case, maintenant ou après son prochain pas.
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

// bouclierDeMonCote garde les boucliers où l'on arrive avant l'ennemi.
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

// objetsUtiles garde les objets qui nous servent vraiment (pas de soin si on a déjà presque tous nos PV).
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

// versObjet renvoie le premier pas vers l'objet le plus proche du type voulu, ou "" s'il n'y en a pas.
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

// versCaseDeTir renvoie le premier pas vers la case la plus proche d'où l'on peut tirer sur l'ennemi.
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

// directionDeTir dit si un tir depuis caseDepart touche la cible (portée 5, murs bloquants) et dans quelle direction.
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
