package services

import (
	"fmt"
	"strings"

	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
)

// CampoCategoriaService gerencia os campos customizados (atributos) de cada
// categoria de item.
type CampoCategoriaService struct {
	repo    repositories.CampoCategoriaRepository
	catRepo repositories.CategoriaRepository
}

func NewCampoCategoriaService(
	repo repositories.CampoCategoriaRepository,
	catRepo repositories.CategoriaRepository,
) *CampoCategoriaService {
	return &CampoCategoriaService{repo: repo, catRepo: catRepo}
}

// EntradaCampo reúne os dados de criação/edição de um campo customizado.
type EntradaCampo struct {
	CategoriaID uint
	Nome        string
	Tipo        models.TipoCampo
	Opcoes      string
	Obrigatorio bool
	Ordem       int
}

func (s *CampoCategoriaService) ListarPorCategoria(categoriaID uint) ([]models.CampoCategoria, error) {
	return s.repo.ListarPorCategoria(categoriaID)
}

func (s *CampoCategoriaService) validar(in EntradaCampo) error {
	ev := NovoErroValidacao()
	if strings.TrimSpace(in.Nome) == "" {
		ev.Add("nome", "Informe o nome do campo.")
	}
	if in.Tipo == "" {
		in.Tipo = models.CampoTexto
	}
	if !models.TipoCampoValido(in.Tipo) {
		ev.Add("tipo", "Tipo de campo inválido.")
	}
	if in.Tipo == models.CampoLista && strings.TrimSpace(in.Opcoes) == "" {
		ev.Add("opcoes", "Informe ao menos uma opção para o campo do tipo lista.")
	}
	if ev.TemErros() {
		return ev
	}
	return nil
}

func (s *CampoCategoriaService) Criar(in EntradaCampo) (*models.CampoCategoria, error) {
	if _, err := s.catRepo.BuscarPorID(in.CategoriaID); err != nil {
		return nil, traduzErroRepo(err)
	}
	if in.Tipo == "" {
		in.Tipo = models.CampoTexto
	}
	if err := s.validar(in); err != nil {
		return nil, err
	}
	chave, err := s.chaveUnica(in.CategoriaID, in.Nome, 0)
	if err != nil {
		return nil, err
	}
	c := &models.CampoCategoria{
		CategoriaID: in.CategoriaID,
		Nome:        strings.TrimSpace(in.Nome),
		Chave:       chave,
		Tipo:        in.Tipo,
		Opcoes:      strings.TrimSpace(in.Opcoes),
		Obrigatorio: in.Obrigatorio,
		Ordem:       in.Ordem,
	}
	if err := s.repo.Criar(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CampoCategoriaService) Atualizar(id uint, in EntradaCampo) (*models.CampoCategoria, error) {
	c, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	if in.Tipo == "" {
		in.Tipo = models.CampoTexto
	}
	if err := s.validar(in); err != nil {
		return nil, err
	}
	// A chave permanece estável (não quebra os atributos já gravados).
	c.Nome = strings.TrimSpace(in.Nome)
	c.Tipo = in.Tipo
	c.Opcoes = strings.TrimSpace(in.Opcoes)
	c.Obrigatorio = in.Obrigatorio
	c.Ordem = in.Ordem
	if err := s.repo.Atualizar(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CampoCategoriaService) Remover(id uint) error {
	if _, err := s.repo.BuscarPorID(id); err != nil {
		return traduzErroRepo(err)
	}
	return s.repo.Remover(id)
}

// chaveUnica gera uma chave slug a partir do nome, garantindo unicidade dentro
// da categoria (acrescenta sufixo numérico em caso de colisão).
func (s *CampoCategoriaService) chaveUnica(categoriaID uint, nome string, ignorarID uint) (string, error) {
	base := gerarChaveCampo(nome)
	if base == "" {
		base = "campo"
	}
	existentes, err := s.repo.ListarPorCategoria(categoriaID)
	if err != nil {
		return "", err
	}
	usadas := map[string]bool{}
	for _, c := range existentes {
		if c.ID == ignorarID {
			continue
		}
		usadas[c.Chave] = true
	}
	chave := base
	for i := 2; usadas[chave]; i++ {
		chave = fmt.Sprintf("%s_%d", base, i)
	}
	return chave, nil
}

// gerarChaveCampo transforma "Memória RAM" em "memoria_ram".
func gerarChaveCampo(nome string) string {
	s := normalizarChave(nome)
	var b strings.Builder
	anteriorUnderscore := false
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			anteriorUnderscore = false
		default:
			if !anteriorUnderscore && b.Len() > 0 {
				b.WriteRune('_')
				anteriorUnderscore = true
			}
		}
	}
	return strings.Trim(b.String(), "_")
}
