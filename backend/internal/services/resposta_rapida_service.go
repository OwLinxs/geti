package services

import (
	"strings"

	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
)

type RespostaRapidaService struct {
	repo repositories.RespostaRapidaRepository
}

func NewRespostaRapidaService(repo repositories.RespostaRapidaRepository) *RespostaRapidaService {
	return &RespostaRapidaService{repo: repo}
}

type EntradaRespostaRapida struct {
	Titulo   string
	Conteudo string
	Ordem    int
	Ativo    bool
}

func (s *RespostaRapidaService) validar(in EntradaRespostaRapida) error {
	ev := NovoErroValidacao()
	if strings.TrimSpace(in.Titulo) == "" {
		ev.Add("titulo", "Informe o título.")
	}
	if strings.TrimSpace(in.Conteudo) == "" {
		ev.Add("conteudo", "Informe o texto da resposta.")
	}
	if ev.TemErros() {
		return ev
	}
	return nil
}

func (s *RespostaRapidaService) Criar(in EntradaRespostaRapida) (*models.RespostaRapida, error) {
	if err := s.validar(in); err != nil {
		return nil, err
	}
	x := &models.RespostaRapida{
		Titulo:   strings.TrimSpace(in.Titulo),
		Conteudo: strings.TrimSpace(in.Conteudo),
		Ordem:    in.Ordem,
		Ativo:    in.Ativo,
	}
	if err := s.repo.Criar(x); err != nil {
		return nil, err
	}
	return x, nil
}

func (s *RespostaRapidaService) Atualizar(id uint, in EntradaRespostaRapida) (*models.RespostaRapida, error) {
	x, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	if err := s.validar(in); err != nil {
		return nil, err
	}
	x.Titulo = strings.TrimSpace(in.Titulo)
	x.Conteudo = strings.TrimSpace(in.Conteudo)
	x.Ordem = in.Ordem
	x.Ativo = in.Ativo
	if err := s.repo.Atualizar(x); err != nil {
		return nil, err
	}
	return x, nil
}

func (s *RespostaRapidaService) Listar(apenasAtivas bool) ([]models.RespostaRapida, error) {
	return s.repo.Listar(apenasAtivas)
}

func (s *RespostaRapidaService) Excluir(id uint) error {
	if _, err := s.repo.BuscarPorID(id); err != nil {
		return traduzErroRepo(err)
	}
	return s.repo.Remover(id)
}

// SemearPadrao cria respostas iniciais se não houver nenhuma.
func (s *RespostaRapidaService) SemearPadrao() error {
	n, err := s.repo.Contar()
	if err != nil || n > 0 {
		return err
	}
	padrao := []models.RespostaRapida{
		{Titulo: "Recebido", Conteudo: "Olá! Recebemos seu chamado e já estamos analisando.", Ordem: 1, Ativo: true},
		{Titulo: "Mais informações", Conteudo: "Para prosseguir, poderia enviar mais detalhes (ou uma foto do erro)?", Ordem: 2, Ativo: true},
		{Titulo: "A caminho", Conteudo: "Um técnico irá até o local em breve.", Ordem: 3, Ativo: true},
		{Titulo: "Resolvido", Conteudo: "Seu chamado foi resolvido. Se precisar, é só reabrir. 🙂", Ordem: 4, Ativo: true},
	}
	for i := range padrao {
		_ = s.repo.Criar(&padrao[i])
	}
	return nil
}
