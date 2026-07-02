package services

import (
	"errors"
	"regexp"

	"github.com/pmfb/sige-ti/internal/repositories"
)

var reEmail = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func emailValido(email string) bool {
	return reEmail.MatchString(email)
}

// traduzErroRepo converte erros do repositório em erros de domínio do serviço.
func traduzErroRepo(err error) error {
	if errors.Is(err, repositories.ErrNaoEncontrado) {
		return ErrNaoEncontrado
	}
	return err
}
