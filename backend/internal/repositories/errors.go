package repositories

import "errors"

// ErrNaoEncontrado é retornado quando um registro não existe.
var ErrNaoEncontrado = errors.New("registro não encontrado")
