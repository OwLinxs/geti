package services

import (
	"fmt"
	"log"

	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
)

// NotificacaoService gera e consulta notificações in-app (sino).
type NotificacaoService struct {
	repo        repositories.NotificacaoRepository
	usuarioRepo repositories.UsuarioRepository
	cfgSvc      *ConfiguracaoService
}

func NewNotificacaoService(
	repo repositories.NotificacaoRepository,
	usuarioRepo repositories.UsuarioRepository,
	cfgSvc *ConfiguracaoService,
) *NotificacaoService {
	return &NotificacaoService{repo: repo, usuarioRepo: usuarioRepo, cfgSvc: cfgSvc}
}

// ---- Consulta (usadas pelos handlers) ----

func (s *NotificacaoService) Listar(usuarioID uint, apenasNaoLidas bool, limite int) ([]models.Notificacao, error) {
	return s.repo.ListarPorUsuario(usuarioID, apenasNaoLidas, limite)
}

func (s *NotificacaoService) ContarNaoLidas(usuarioID uint) (int64, error) {
	return s.repo.ContarNaoLidas(usuarioID)
}

func (s *NotificacaoService) MarcarLida(id, usuarioID uint) error {
	return s.repo.MarcarLida(id, usuarioID)
}

func (s *NotificacaoService) MarcarTodasLidas(usuarioID uint) error {
	return s.repo.MarcarTodasLidas(usuarioID)
}

// ---- Emissão de eventos (best-effort; nunca quebram a operação principal) ----

func (s *NotificacaoService) config() ConfigChamados {
	cfg, err := s.cfgSvc.ObterChamados()
	if err != nil {
		return ConfigChamadosPadrao()
	}
	return cfg
}

// usuariosEquipe devolve os usuários da equipe (admin/operador) ativos, exceto
// o id informado (para não notificar o próprio autor).
func (s *NotificacaoService) usuariosEquipe(excluir uint) []models.Usuario {
	usuarios, err := s.usuarioRepo.Listar()
	if err != nil {
		log.Printf("[notif] falha ao listar equipe: %v", err)
		return nil
	}
	var eq []models.Usuario
	for _, u := range usuarios {
		if u.Ativo && models.PerfilEquipe(u.Perfil) && u.ID != excluir {
			eq = append(eq, u)
		}
	}
	return eq
}

// usuario devolve um usuário pelo id (para notificar o solicitante).
func (s *NotificacaoService) usuario(id uint) []models.Usuario {
	if id == 0 {
		return nil
	}
	u, err := s.usuarioRepo.BuscarPorID(id)
	if err != nil {
		return nil
	}
	return []models.Usuario{*u}
}

// emitir grava as notificações in-app e, se o e-mail estiver ligado, dispara os
// e-mails de forma assíncrona (best-effort).
func (s *NotificacaoService) emitir(dest []models.Usuario, titulo, msg, tipo string, recursoID uint) {
	if len(dest) == 0 {
		return
	}
	ns := make([]models.Notificacao, 0, len(dest))
	rid := recursoID
	for _, u := range dest {
		ns = append(ns, models.Notificacao{
			UsuarioID: u.ID,
			Titulo:    titulo,
			Mensagem:  msg,
			Tipo:      tipo,
			RecursoID: &rid,
		})
	}
	if err := s.repo.CriarVarias(ns); err != nil {
		log.Printf("[notif] falha ao gravar notificações: %v", err)
	}

	cfg := s.config()
	if !cfg.EmailAtivo {
		return
	}
	emails := make([]string, 0, len(dest))
	for _, u := range dest {
		if u.Email != "" {
			emails = append(emails, u.Email)
		}
	}
	if len(emails) == 0 {
		return
	}
	corpo := titulo
	if msg != "" {
		corpo = titulo + "\n\n" + msg
	}
	// Envio assíncrono: não bloqueia a operação principal.
	go func(emails []string, assunto, corpo string) {
		if err := enviarEmailSMTP(cfg, emails, assunto, corpo); err != nil {
			log.Printf("[notif] falha ao enviar e-mail: %v", err)
		}
	}(emails, titulo, corpo)
}

// NotificarNovoChamado avisa a equipe quando um chamado é aberto.
func (s *NotificacaoService) NotificarNovoChamado(os *models.OrdemServico) {
	if os == nil || !s.config().NotificarEquipeNovoChamado {
		return
	}
	var autor uint
	if os.AbertoPorID != 0 {
		autor = os.AbertoPorID
	}
	titulo := fmt.Sprintf("Novo chamado %s", os.Numero)
	msg := os.EquipamentoSnapshot
	if os.Assunto != "" {
		msg = os.Assunto
	}
	s.emitir(s.usuariosEquipe(autor), titulo, msg, models.NotifChamadoNovo, os.ID)
}

// NotificarMensagem avisa o outro lado quando há nova mensagem no chamado.
func (s *NotificacaoService) NotificarMensagem(os *models.OrdemServico, autorTipo models.AutorTipo, interna bool) {
	if os == nil || interna {
		return
	}
	if autorTipo == models.AutorServidor {
		// Solicitante escreveu → avisa a equipe.
		s.emitir(s.usuariosEquipe(os.AbertoPorID),
			fmt.Sprintf("Nova mensagem no chamado %s", os.Numero),
			"O solicitante respondeu.", models.NotifMensagem, os.ID)
		return
	}
	// Equipe escreveu → avisa o solicitante.
	if os.AbertoPorID != 0 && s.config().NotificarSolicitanteResposta {
		s.emitir(s.usuario(os.AbertoPorID),
			fmt.Sprintf("Resposta no seu chamado %s", os.Numero),
			"A equipe de T.I. respondeu.", models.NotifMensagem, os.ID)
	}
}

// NotificarSLAEstourado avisa a equipe quando um chamado fura o prazo de SLA.
func (s *NotificacaoService) NotificarSLAEstourado(os *models.OrdemServico, tipo string) {
	if os == nil {
		return
	}
	oque := "resolução"
	if tipo == "resposta" {
		oque = "resposta"
	}
	s.emitir(s.usuariosEquipe(0),
		fmt.Sprintf("SLA estourado: chamado %s", os.Numero),
		fmt.Sprintf("O prazo de %s foi ultrapassado.", oque),
		models.NotifStatus, os.ID)
}

// NotificarReaberto avisa a equipe quando o solicitante reabre um chamado.
func (s *NotificacaoService) NotificarReaberto(os *models.OrdemServico, motivo string) {
	if os == nil {
		return
	}
	msg := "O solicitante reabriu o chamado."
	if motivo != "" {
		msg = "Reaberto: " + motivo
	}
	s.emitir(s.usuariosEquipe(0),
		fmt.Sprintf("Chamado %s reaberto", os.Numero),
		msg, models.NotifStatus, os.ID)
}

// NotificarStatus avisa o solicitante quando o status do chamado muda.
func (s *NotificacaoService) NotificarStatus(os *models.OrdemServico) {
	if os == nil || os.AbertoPorID == 0 || !s.config().NotificarSolicitanteStatus {
		return
	}
	s.emitir(s.usuario(os.AbertoPorID),
		fmt.Sprintf("Chamado %s: status atualizado", os.Numero),
		fmt.Sprintf("Novo status: %s", os.Status), models.NotifStatus, os.ID)
}
