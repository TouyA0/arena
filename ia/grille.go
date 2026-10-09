package ia

import "robot/carte"

type Pos struct {
	X, Y int
}

var directions = []string{"N", "S", "E", "O"}

func voisin(p Pos, d string) Pos {
	switch d {
	case "N":
		return Pos{p.X, p.Y - 1}
	case "S":
		return Pos{p.X, p.Y + 1}
	case "E":
		return Pos{p.X + 1, p.Y}
	case "O":
		return Pos{p.X - 1, p.Y}
	}
	return p
}

const (
	CoutSol        = 1
	CoutEau        = 3
	CoutPiegeLeger = 4  // quand bcp de vie
	CoutPiegeLourd = 10 // quand peu de vie
	SeuilPVBas     = 50
	DegatsPiege    = 15
	Interdit       = -1
)

func symbole(c carte.Carte, p Pos) byte {
	if p.Y < 0 || p.Y >= c.Hauteur || p.X < 0 || p.X >= c.Largeur {
		return '#' // si extérieur de la carte
	}
	return c.Grille[p.Y][p.X]
}

func cout(c carte.Carte, p Pos, pv int) int {
	switch symbole(c, p) {
	case '#':
		return Interdit
	case '~':
		return CoutEau
	case '^':
		if pv <= DegatsPiege {
			return Interdit // la piege va nous tuer
		}
		if pv <= SeuilPVBas {
			return CoutPiegeLourd
		}
		return CoutPiegeLeger
	default:
		return CoutSol
	}
}
