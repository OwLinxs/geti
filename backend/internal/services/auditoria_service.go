package services

import (
	"log"

	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
)

// AuditoriaService registra e consulta a trilha de auditoria. O registro é
// "best-effort": uma falha ao gravar auditoria NUNCA deve derrubar a operação
// principal do usuário — apenas logamos o erro.
type AuditoriaService struct {
	repo repositories.AuditoriaRepository
}

func NewAuditoriaService(repo repositories.AuditoriaRepository) *AuditoriaService {
	return &AuditoriaService{repo: repo}
}

// Registrar grava um evento de auditoria. Erros são logados, não propagados.
func (s *AuditoriaService) Registrar(reg *models.RegistroAuditoria) {
	if err := s.repo.Registrar(reg); err != nil {
		log.Printf("auditoria: falha ao registrar evento (%s %s): %v", reg.Acao, reg.Recurso, err)
	}
}

// Listar devolve os eventos filtrados e paginados (mais recentes primeiro).
func (s *AuditoriaService) Listar(f repositories.FiltroAuditoria) ([]models.RegistroAuditoria, int64, error) {
	return s.repo.Listar(f)
}
