package protocole

import (
	"fmt"
	"strconv"
	"strings"
)

type Robot struct {
	NOM string
	X, Y int
	VIE int
}

type Etat struct {
	Tour int;
	Moi, Ennemi Robot
}

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
				return Etat{}, fmt.Errorf("MOI: erreur de lecture %v", err)
			}
			e.Moi = robot
			okMoi = true

		case "ENNEMI":
			if okEnnemi {
				return Etat{}, fmt.Errorf("ENNEMI: en double")
			}
			robot, err := LireRobot(champs)
			if err != nil {
				return Etat{}, fmt.Errorf("ENNEMI: erreur de lecture %v", err)
			}
			e.Ennemi = robot
			okEnnemi = true

		default:
		}
	}
	if !okTour || !okMoi || !okEnnemi {
		return Etat{}, fmt.Errorf("incomplet: tour=%v, moi=%v, ennemi=%v", okTour, okMoi, okEnnemi)
	}
	return e, nil
}

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
	return Robot{NOM: nom, X: x_pos, Y: y_pos, VIE: vie}, nil
}
