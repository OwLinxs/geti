package services

import (
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
)

// PermissaoService concentra as regras de permissão por usuário (overrides
// sobre a baseline do perfil).
//
// Baseline:
//   - administrador: tem TODAS as permissões (super-admin, não editável) — evita
//     trancar o sistema (lockout).
//   - operador: parte SEM nenhuma permissão privilegiada; recebe-as uma a uma.
//   - solicitante: fora da área de equipe; não participa deste catálogo.
type PermissaoService struct {
	repo        repositories.PermissaoUsuarioRepository
	usuarioRepo repositories.UsuarioRepository
}

func NewPermissaoService(repo repositories.PermissaoUsuarioRepository, usuarioRepo repositories.UsuarioRepository) *PermissaoService {
	return &PermissaoService{repo: repo, usuarioRepo: usuarioRepo}
}

// Catalogo devolve a lista de permissões delegáveis.
func (s *PermissaoService) Catalogo() []models.Permissao {
	return models.CatalogoPermissoes()
}

// Tem informa se o usuário (com o perfil informado) possui a permissão. O
// administrador sempre tem; os demais dependem de concessão explícita.
func (s *PermissaoService) Tem(usuarioID uint, perfil models.Perfil, chave string) (bool, error) {
	if perfil == models.PerfilAdministrador {
		return true, nil
	}
	return s.repo.Existe(usuarioID, chave)
}

// EstadoPermissoes é o retorno da consulta das permissões de um usuário.
type EstadoPermissoes struct {
	UsuarioID  uint          `json:"usuario_id"`
	Perfil     models.Perfil `json:"perfil"`
	TodasFixas bool          `json:"todas_fixas"` // true p/ administrador (não editável)
	Editavel   bool          `json:"editavel"`    // true só p/ operador
	Concedidas []string      `json:"concedidas"`  // chaves efetivamente ativas
}

// DoUsuario devolve o estado de permissões de um usuário específico.
func (s *PermissaoService) DoUsuario(usuarioID uint) (*EstadoPermissoes, error) {
	u, err := s.usuarioRepo.BuscarPorID(usuarioID)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	est := &EstadoPermissoes{UsuarioID: u.ID, Perfil: u.Perfil}
	switch u.Perfil {
	case models.PerfilAdministrador:
		est.TodasFixas = true
		for _, p := range models.CatalogoPermissoes() {
			est.Concedidas = append(est.Concedidas, p.Chave)
		}
	case models.PerfilOperador:
		est.Editavel = true
		chaves, err := s.repo.Concedidas(u.ID)
		if err != nil {
			return nil, err
		}
		est.Concedidas = chaves
	default: // solicitante
		est.Concedidas = []string{}
	}
	if est.Concedidas == nil {
		est.Concedidas = []string{}
	}
	return est, nil
}

// Salvar substitui o conjunto de permissões concedidas a um usuário. Só é
// permitido para operadores (administrador já tem tudo; solicitante não tem
// área de equipe). Ignora chaves fora do catálogo.
func (s *PermissaoService) Salvar(usuarioID uint, chaves []string) (*EstadoPermissoes, error) {
	u, err := s.usuarioRepo.BuscarPorID(usuarioID)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	if u.Perfil != models.PerfilOperador {
		ev := NovoErroValidacao()
		ev.Add("perfil", "Só é possível definir permissões para usuários com perfil 'operador'. Administrador já possui acesso total.")
		return nil, ev
	}
	// Filtra/deduplica para o catálogo válido.
	vistos := map[string]bool{}
	validas := make([]string, 0, len(chaves))
	for _, ch := range chaves {
		if models.PermissaoValida(ch) && !vistos[ch] {
			vistos[ch] = true
			validas = append(validas, ch)
		}
	}
	if err := s.repo.Substituir(u.ID, validas); err != nil {
		return nil, err
	}
	return s.DoUsuario(u.ID)
}
