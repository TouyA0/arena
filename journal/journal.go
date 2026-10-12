package journal

import (
	"bufio"
	"fmt"
	"os"

	"robot/protocole"
)

type Journal struct {
	f *os.File
	w *bufio.Writer
}

// Ouvrir crée (ou vide) le fichier du journal et le prépare pour l'écriture.
func Ouvrir(chemin string) (*Journal, error) {
	f, err := os.Create(chemin)
	if err != nil {
		return nil, fmt.Errorf("journal : %w", err)
	}
	return &Journal{f: f, w: bufio.NewWriter(f)}, nil
}

// Debut écrit la première ligne du journal : la carte et notre numéro de joueur.
func Debut(j *Journal, nomCarte string, joueur int) {
	ecrire(j, "Partie sur %s, joueur %d", nomCarte, joueur)
}

// Tour écrit une ligne pour le tour : position, PV, action choisie et sa raison.
func Tour(j *Journal, etat protocole.Etat, action, raison string) {
	ecrire(j, "T%d | (%d,%d) %d PV | %s | %s",
		etat.Tour, etat.Moi.X, etat.Moi.Y, etat.Moi.VIE, action, raison)
}

// BlocInvalide note un bloc refusé par le protocole, et qu'on a répondu ATTENDS.
func BlocInvalide(j *Journal, err error) {
	ecrire(j, "bloc invalide | %v | ATTENDS", err)
}

// Fermer vide le tampon et ferme le fichier. Elle ne fait rien si le journal n'a pas pu être ouvert.
func Fermer(j *Journal) error {
	if j == nil {
		return nil
	}
	if err := j.w.Flush(); err != nil {
		j.f.Close()
		return err
	}
	return j.f.Close()
}

// ecrire ajoute une ligne au journal et l'écrit tout de suite dans le fichier,
// pour ne rien perdre si le robot est arrêté brutalement.
func ecrire(j *Journal, format string, valeurs ...any) {
	if j == nil {
		return
	}
	fmt.Fprintf(j.w, format+"\n", valeurs...)
	j.w.Flush()
}
