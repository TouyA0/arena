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

func Ouvrir(chemin string) (*Journal, error) {
    f, err := os.Create(chemin)
    if err != nil {
        return nil, fmt.Errorf("journal : %w", err)
    }
    return &Journal{f: f, w: bufio.NewWriter(f)}, nil
}

func Debut(j *Journal, nomCarte string, joueur int) {
    ecrire(j, "Partie sur %s, joueur %d", nomCarte, joueur)
}

func Tour(j *Journal, etat protocole.Etat, action, raison string) {
    ecrire(j, "T%d | (%d,%d) %d PV | %s | %s",
        etat.Tour, etat.Moi.X, etat.Moi.Y, etat.Moi.VIE, action, raison)
}

func BlocInvalide(j *Journal, err error) {
    ecrire(j, "bloc invalide | %v | ATTENDS", err)
}

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

func ecrire(j *Journal, format string, valeurs ...any) {
    if j == nil {
        return
    }
    fmt.Fprintf(j.w, format+"\n", valeurs...)
}