package services

import (
	"fmt"
	"net/smtp"
	"strings"
)

// enviarEmailSMTP envia um e-mail simples (texto) via SMTP com autenticação e
// STARTTLS (padrão em porta 587). Best-effort: o chamador trata/loga o erro.
func enviarEmailSMTP(cfg ConfigChamados, para []string, assunto, corpo string) error {
	if cfg.SMTPHost == "" || cfg.SMTPPorta == 0 {
		return fmt.Errorf("SMTP não configurado")
	}
	destinatarios := make([]string, 0, len(para))
	for _, p := range para {
		if p = strings.TrimSpace(p); p != "" {
			destinatarios = append(destinatarios, p)
		}
	}
	if len(destinatarios) == 0 {
		return nil
	}

	remetente := cfg.SMTPRemetente
	if remetente == "" {
		remetente = cfg.SMTPUsuario
	}

	cabecalho := map[string]string{
		"From":         remetente,
		"To":           strings.Join(destinatarios, ", "),
		"Subject":      assunto,
		"MIME-Version": "1.0",
		"Content-Type": "text/plain; charset=\"UTF-8\"",
	}
	var sb strings.Builder
	for k, v := range cabecalho {
		fmt.Fprintf(&sb, "%s: %s\r\n", k, v)
	}
	sb.WriteString("\r\n")
	sb.WriteString(corpo)

	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPorta)
	var auth smtp.Auth
	if cfg.SMTPUsuario != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUsuario, cfg.SMTPSenha, cfg.SMTPHost)
	}
	return smtp.SendMail(addr, auth, remetente, destinatarios, []byte(sb.String()))
}
