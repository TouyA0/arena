package carte

import "os"

type Carte struct {
	Nom string
	Largeur int
	Hauteur int
	Grille []string
}
func Charger(chemin string) (Carte, error) {
	_, err := os.ReadFile(chemin)
	if err != nil {
		
	}
	// TODO : en-tête, grille, règles
	return Carte{}, nil
}
