package services

import (
	"encoding/json"

	"github.com/pmfb/sige-ti/internal/repositories"
)

const chaveConfigChamados = "chamados"

// ConfigChamados são os parâmetros configuráveis do módulo de chamados.
type ConfigChamados struct {
	NotificarEquipeNovoChamado   bool `json:"notificar_equipe_novo_chamado"`
	NotificarSolicitanteResposta bool `json:"notificar_solicitante_resposta"`
	NotificarSolicitanteStatus   bool `json:"notificar_solicitante_status"`

	// Envio por e-mail (SMTP). EmailAtivo liga o envio; sem isso só há sino.
	EmailAtivo    bool   `json:"email_ativo"`
	SMTPHost      string `json:"smtp_host"`
	SMTPPorta     int    `json:"smtp_porta"`
	SMTPUsuario   string `json:"smtp_usuario"`
	SMTPSenha     string `json:"smtp_senha"`
	SMTPRemetente string `json:"smtp_remetente"`

	// SLA: prazos em HORAS por prioridade. 0 = sem prazo (não cobra SLA).
	// Resposta = 1ª resposta da equipe; Resolução = conclusão do chamado.
	SLARespostaAltaH    int `json:"sla_resposta_alta_h"`
	SLARespostaNormalH  int `json:"sla_resposta_normal_h"`
	SLARespostaBaixaH   int `json:"sla_resposta_baixa_h"`
	SLAResolucaoAltaH   int `json:"sla_resolucao_alta_h"`
	SLAResolucaoNormalH int `json:"sla_resolucao_normal_h"`
	SLAResolucaoBaixaH  int `json:"sla_resolucao_baixa_h"`

	// Auto-cadastro de solicitante. Ativo liga o cadastro público;
	// DominiosPermitidos (separados por vírgula) restringe os e-mails aceitos
	// (vazio = qualquer domínio).
	AutoCadastroAtivo  bool   `json:"auto_cadastro_ativo"`
	DominiosPermitidos string `json:"dominios_permitidos"`
}

// ConfigChamadosPadrao são os defaults (sino ligado, e-mail desligado).
func ConfigChamadosPadrao() ConfigChamados {
	return ConfigChamados{
		NotificarEquipeNovoChamado:   true,
		NotificarSolicitanteResposta: true,
		NotificarSolicitanteStatus:   true,
		EmailAtivo:                   false,
		SMTPPorta:                    587,
		SLARespostaAltaH:             2,
		SLARespostaNormalH:           8,
		SLARespostaBaixaH:            24,
		SLAResolucaoAltaH:            8,
		SLAResolucaoNormalH:          24,
		SLAResolucaoBaixaH:           72,
		AutoCadastroAtivo:            true,
	}
}

// HorasSLA devolve os prazos (resposta, resolução) em horas para uma prioridade.
func (c ConfigChamados) HorasSLA(prioridade string) (resposta, resolucao int) {
	switch prioridade {
	case "alta":
		return c.SLARespostaAltaH, c.SLAResolucaoAltaH
	case "baixa":
		return c.SLARespostaBaixaH, c.SLAResolucaoBaixaH
	default:
		return c.SLARespostaNormalH, c.SLAResolucaoNormalH
	}
}

type ConfiguracaoService struct {
	repo repositories.ConfiguracaoRepository
}

func NewConfiguracaoService(repo repositories.ConfiguracaoRepository) *ConfiguracaoService {
	return &ConfiguracaoService{repo: repo}
}

// ObterChamados devolve a config de chamados, começando pelos defaults e
// sobrescrevendo com o que estiver salvo (campos novos herdam o default).
func (s *ConfiguracaoService) ObterChamados() (ConfigChamados, error) {
	cfg := ConfigChamadosPadrao()
	raw, err := s.repo.Obter(chaveConfigChamados)
	if err != nil {
		if err == repositories.ErrNaoEncontrado {
			return cfg, nil
		}
		return cfg, err
	}
	// Ignora erro de unmarshal: mantém defaults se o JSON estiver corrompido.
	_ = json.Unmarshal([]byte(raw), &cfg)
	return cfg, nil
}

func (s *ConfiguracaoService) SalvarChamados(cfg ConfigChamados) (ConfigChamados, error) {
	b, err := json.Marshal(cfg)
	if err != nil {
		return cfg, err
	}
	if err := s.repo.Salvar(chaveConfigChamados, string(b)); err != nil {
		return cfg, err
	}
	return cfg, nil
}
