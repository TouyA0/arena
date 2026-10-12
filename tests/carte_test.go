package tests

import (
	"path/filepath"
	"testing"

	"robot/carte"
)

// bonnes cartes
func TestCartesValides(t *testing.T) {
	fichiers, _ := filepath.Glob("cartes/*.map")
	if len(fichiers) == 0 {
		t.Fatal("aucune carte trouvée")
	}
	for _, f := range fichiers {
		_, err := carte.Charger(f)
		if err != nil {
			t.Errorf("%s refusée : %v", f, err)
		}
	}
}

// cartes avec erreurs
func TestCartesInvalides(t *testing.T) {
	cas := []struct {
		fichier string
		message string
	}{
		{"02_symbole_inconnu.map", `02_symbole_inconnu.map ligne 8 colonne 5 : symbole inconnu "?"`},
		{"05_bordure_ouverte.map", "05_bordure_ouverte.map ligne 8 colonne 1 : la bordure doit être un mur"},
		{"07_depart_double.map", "07_depart_double.map ligne 8 colonne 4 : départ du joueur 1 déjà défini"},
		{"08_cle_inconnue.map", `08_cle_inconnue.map ligne 2 : clé inconnue "VITESSE"`},
	}
	for _, c := range cas {
		_, err := carte.Charger("cartes/invalides/" + c.fichier)
		if err == nil {
			t.Errorf("%s acceptée alors qu'elle est invalide", c.fichier)
		} else if err.Error() != c.message {
			t.Errorf("%s :\nreçu    : %s\nattendu : %s", c.fichier, err.Error(), c.message)
		}
	}
}
