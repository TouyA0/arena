package protocole

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// MunIllimitees est la valeur de MUN quand l'arbitre n'envoie pas de munitions
const MunIllimitees = -1

type Robot struct {
	NOM      string
	X, Y     int
	VIE      int
	MUN      int
	BOUCLIER int
}

type Objet struct {
	X, Y int
	Type string
}

type Etat struct {
	Tour        int
	Moi, Ennemi Robot
	EnnemiVu    bool
	Objets      []Objet
}

// Lire transforme les lignes d'un bloc en Etat. TOUR et MOI sont obligatoires,
// ENNEMI et OBJET sont facultatifs, les mots-clés inconnus sont ignorés.
func Lire(lignes []string) (Etat, error) {
	var e Etat
	okTour, okMoi, okEnnemi := false, false, false

	for _, line := range lignes {
		champs := strings.Fields(line)
		if len(champs) == 0 {
			continue
		}

		switch champs[0] {
		case "TOUR":
			if okTour {
				return Etat{}, fmt.Errorf("TOUR: en double")
			}
			if len(champs) < 2 {
				return Etat{}, fmt.Errorf("TOUR: sans numéro")
			}
			nb_tour, err := strconv.Atoi(champs[1])
			if err != nil {
				return Etat{}, fmt.Errorf("TOUR: numéro invalide %q", champs[1])
			}
			e.Tour = nb_tour
			okTour = true

		case "MOI":
			if okMoi {
				return Etat{}, fmt.Errorf("MOI: en double")
			}
			robot, err := LireRobot(champs)
			if err != nil {
				return Etat{}, err
			}
			e.Moi = robot
			okMoi = true

		case "ENNEMI":
			if okEnnemi {
				return Etat{}, fmt.Errorf("ENNEMI: en double")
			}
			robot, err := LireRobot(champs)
			if err != nil {
				return Etat{}, err
			}
			e.Ennemi = robot
			e.EnnemiVu = true
			okEnnemi = true

		case "OBJET":
			objet, err := LireObjet(champs)
			if err != nil {
				return Etat{}, err
			}
			e.Objets = append(e.Objets, objet)

		default:
		}
	}
	if !okTour || !okMoi {
		return Etat{}, fmt.Errorf("incomplet: tour=%v, moi=%v", okTour, okMoi)
	}
	return e, nil
}

// LireRobot lit "MOI x y VIE v" ou "ENNEMI x y VIE v", suivi éventuellement
// de "MUN m" et "BOUCLIER b". Les champs inconnus en fin de ligne sont ignorés.
func LireRobot(champs []string) (Robot, error) {
	nom := champs[0]
	if len(champs) < 5 {
		return Robot{}, fmt.Errorf("%s: champ manquant", nom)
	}
	if champs[3] != "VIE" {
		return Robot{}, fmt.Errorf("%s: attends VIE et a reçu %q", nom, champs[3])
	}
	x_pos, err_x := strconv.Atoi(champs[1])
	y_pos, err_y := strconv.Atoi(champs[2])
	vie, err_vie := strconv.Atoi(champs[4])
	if err_x != nil || err_y != nil || err_vie != nil {
		return Robot{}, fmt.Errorf("%s: pas un nombre", nom)
	}
	robot := Robot{NOM: nom, X: x_pos, Y: y_pos, VIE: vie, MUN: MunIllimitees, BOUCLIER: 0}

	for i := 5; i < len(champs); i++ {
		cle := champs[i]
		if cle != "MUN" && cle != "BOUCLIER" {
			continue
		}
		if i+1 >= len(champs) {
			return Robot{}, fmt.Errorf("%s: %s sans valeur", nom, cle)
		}
		valeur, err := strconv.Atoi(champs[i+1])
		if err != nil {
			return Robot{}, fmt.Errorf("%s: %s pas un nombre %q", nom, cle, champs[i+1])
		}
		if cle == "MUN" {
			robot.MUN = valeur
		} else {
			robot.BOUCLIER = valeur
		}
		i++
	}
	return robot, nil
}

// LireObjet lit une ligne "OBJET x y TYPE".
func LireObjet(champs []string) (Objet, error) {
	if len(champs) < 4 {
		return Objet{}, fmt.Errorf("OBJET: champ manquant")
	}
	x_pos, err_x := strconv.Atoi(champs[1])
	y_pos, err_y := strconv.Atoi(champs[2])
	if err_x != nil || err_y != nil {
		return Objet{}, fmt.Errorf("OBJET: pas un nombre")
	}
	return Objet{X: x_pos, Y: y_pos, Type: champs[3]}, nil
}

// LireBloc lit les lignes de l'arbitre jusqu'à FIN et renvoie l'état du tour.
// Elle renvoie io.EOF quand l'entrée est fermée (fin de la partie).
func LireBloc(s *bufio.Scanner) (Etat, error) {
	var bloc []string
	for s.Scan() { // lit la ligne suivante
		champs := strings.Fields(s.Text()) // recup texte de la ligne
		if len(champs) > 0 && champs[0] == "FIN" {
			return Lire(bloc)
		}
		bloc = append(bloc, s.Text())
	}
	return Etat{}, io.EOF
}

// Repondre envoie l'action du tour à l'arbitre sur la sortie standard.
func Repondre(action string) {
	fmt.Println(action)
}

// NouveauLecteur prépare la lecture de l'arbitre, avec des lignes jusqu'à 1 Mo
// au lieu de 64 Ko par défaut.
func NouveauLecteur(entree io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(entree)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	return scanner
}
