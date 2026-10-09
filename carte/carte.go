package carte

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Position struct {
	X, Y int
}

type Carte struct {
	Nom        string
	Difficulte int
	Tours      int
	Largeur    int
	Hauteur    int
	Grille     []string
	Depart1    Position
	Depart2    Position
}

func Charger(chemin string) (Carte, error) {
	nom := filepath.Base(chemin)

	data, err := os.ReadFile(chemin)
	if err != nil {
		return Carte{}, fmt.Errorf("%s : impossible de lire le fichier", nom)
	}
	if strings.TrimSpace(string(data)) == "" {
		return Carte{}, fmt.Errorf("%s : fichier vide", nom)
	}

	// enleve \r des fichiers windows
	lignes := strings.Split(string(data), "\n")
	for i := range lignes {
		lignes[i] = strings.TrimRight(lignes[i], "\r")
	}

	var c Carte

	debutGrille, err := lireEntete(nom, lignes, &c)
	if err != nil {
		return Carte{}, err
	}

	err = lireGrille(nom, lignes, debutGrille, &c)
	if err != nil {
		return Carte{}, err
	}

	err = verifierCases(nom, debutGrille, &c)
	if err != nil {
		return Carte{}, err
	}

	return c, nil
}

func lireEntete(nom string, lignes []string, c *Carte) (int, error) {
	vu := map[string]bool{}
	i := 0

	for i < len(lignes) {
		ligne := lignes[i]
		num_ligne := i + 1

		if strings.HasPrefix(ligne, "#") {
			break
		}
		champs := strings.Fields(ligne)
		if len(champs) == 0 {
			i++
			continue
		}

		cle := champs[0]
		if cle != "NOM" && cle != "DIFFICULTE" && cle != "TOURS" && cle != "TAILLE" {
			return 0, fmt.Errorf("%s ligne %d : clé inconnue %q", nom, num_ligne, cle)
		}
		if vu[cle] {
			return 0, fmt.Errorf("%s ligne %d : %s défini deux fois", nom, num_ligne, cle)
		}
		vu[cle] = true

		switch cle {
		case "NOM":
			c.Nom = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(ligne), "NOM"))
			if c.Nom == "" {
				return 0, fmt.Errorf("%s ligne %d : NOM attend une valeur", nom, num_ligne)
			}
		
		case "DIFFICULTE":
			d, err := entier(champs, 1)
			if err != nil || len(champs) != 2 || d < 1 || d > 3 {
				return 0, fmt.Errorf("%s ligne %d : DIFFICULTE attend un entier entre 1 et 3", nom, num_ligne)
			}
			c.Difficulte = d

		case "TOURS":
			t, err := entier(champs, 1)
			if err != nil || len(champs) != 2 || t < 1 {
				return 0, fmt.Errorf("%s ligne %d : TOURS attend un entier positif", nom, num_ligne)
			}
			c.Tours = t

		case "TAILLE":
			l, err1 := entier(champs, 1)
			h, err2 := entier(champs, 2)
			if err1 != nil || err2 != nil || len(champs) != 3 || l < 3 || h < 3 {
				return 0, fmt.Errorf("%s ligne %d : TAILLE attend deux entiers", nom, num_ligne)
			}
			c.Largeur, c.Hauteur = l, h
		}
		i++
	}
	for _, cle := range []string{"NOM", "DIFFICULTE", "TOURS", "TAILLE"} {
		if !vu[cle] {
			return 0, fmt.Errorf("%s : en-tête incomplet : %s manquant", nom, cle)
		}
	}
	return i, nil
}

func entier(champs []string, i int) (int, error) {
	if i >= len(champs) {
		return 0, fmt.Errorf("champ manquant")
	}
	return strconv.Atoi(champs[i])
}

func lireGrille(nom string, lignes []string, debut int, c *Carte) error {
	grille := lignes[debut:]

	for len(grille) > 0 && strings.TrimSpace(grille[len(grille)-1]) == "" {
		grille = grille[:len(grille)-1]
	}

	for i, ligne := range grille {
		nbCases := len([]rune(ligne))
		if nbCases != c.Largeur {
			return fmt.Errorf("%s ligne %d : %d cases, %d attendues", nom, debut+i+1, nbCases, c.Largeur)
		}
	}
	if len(grille) != c.Hauteur {
		return fmt.Errorf("%s : %d lignes de grille, %d attendues", nom, len(grille), c.Hauteur)
	}

	c.Grille = grille
	return nil
}

func verifierCases(nom string, debut int, c *Carte) error {
	vu1, vu2 := false, false
 
	for y, ligne := range c.Grille {
		for x, s := range []rune(ligne) {
			numLigne := debut + y + 1
			numColonne := x + 1
 
			if !strings.ContainsRune("#.~^+*!$12", s) {
				return fmt.Errorf("%s ligne %d colonne %d : symbole inconnu %q", nom, numLigne, numColonne, string(s))
			}
 
			surBord := x == 0 || y == 0 || x == c.Largeur-1 || y == c.Hauteur-1
			if surBord && s != '#' {
				return fmt.Errorf("%s ligne %d colonne %d : la bordure doit être un mur", nom, numLigne, numColonne)
			}
 
			if s == '1' {
				if vu1 {
					return fmt.Errorf("%s ligne %d colonne %d : départ du joueur 1 déjà défini", nom, numLigne, numColonne)
				}
				vu1 = true
				c.Depart1 = Position{X: x, Y: y}
			}
			if s == '2' {
				if vu2 {
					return fmt.Errorf("%s ligne %d colonne %d : départ du joueur 2 déjà défini", nom, numLigne, numColonne)
				}
				vu2 = true
				c.Depart2 = Position{X: x, Y: y}
			}
		}
	}
 
	if !vu1 {
		return fmt.Errorf("%s : aucun départ pour le joueur 1", nom)
	}
	if !vu2 {
		return fmt.Errorf("%s : aucun départ pour le joueur 2", nom)
	}
	return nil
}