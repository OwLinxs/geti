package services

import "errors"

// Erros de domínio retornados pelos serviços. Os handlers mapeiam estes
// erros para os códigos HTTP apropriados, mantendo mensagens em português.
var (
	ErrNaoEncontrado     = errors.New("registro não encontrado")
	ErrValidacao         = errors.New("dados inválidos")
	ErrConflito          = errors.New("conflito de dados")
	ErrEstoqueInsuficiente = errors.New("estoque insuficiente para a saída solicitada")
	ErrItemBaixado       = errors.New("item já recebeu baixa patrimonial e não permite novas movimentações")
	ErrItemComHistorico  = errors.New("não é possível excluir: o item possui histórico de movimentações ou termos. Use a baixa patrimonial para retirá-lo do estoque")
	ErrCredenciaisInvalidas = errors.New("e-mail ou senha inválidos")
	ErrNaoAutorizado     = errors.New("acesso não autorizado")
	ErrRegraNegocio      = errors.New("operação não permitida pelas regras de negócio")
)

// ErroValidacao agrega mensagens de validação por campo, em português.
type ErroValidacao struct {
	Campos map[string]string
}

func (e *ErroValidacao) Error() string {
	return ErrValidacao.Error()
}

func (e *ErroValidacao) Is(target error) bool {
	return target == ErrValidacao
}

// NovoErroValidacao cria um agregador de erros de validação.
func NovoErroValidacao() *ErroValidacao {
	return &ErroValidacao{Campos: map[string]string{}}
}

func (e *ErroValidacao) Add(campo, mensagem string) {
	e.Campos[campo] = mensagem
}

func (e *ErroValidacao) TemErros() bool {
	return len(e.Campos) > 0
}
