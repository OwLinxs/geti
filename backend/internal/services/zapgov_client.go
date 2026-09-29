package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/pmfb/sige-ti/internal/config"
)

// ErrJanelaFechada indica que a janela de 24h do WhatsApp está fechada e o envio
// de texto livre foi recusado (é preciso usar um template aprovado).
var ErrJanelaFechada = errors.New("janela de 24h fechada")

// ZapGovClient fala com a API do ZapGov (login JWT + envio de mensagens).
type ZapGovClient struct {
	cfg  *config.Config
	http *http.Client

	mu       sync.Mutex
	token    string
	expiraEm time.Time
}

func NewZapGovClient(cfg *config.Config) *ZapGovClient {
	return &ZapGovClient{cfg: cfg, http: &http.Client{Timeout: 20 * time.Second}}
}

// Configurado informa se há credenciais para enviar mensagens.
func (z *ZapGovClient) Configurado() bool {
	return z.cfg.ZapGovBaseURL != "" && z.cfg.ZapGovEmail != "" && z.cfg.ZapGovSenha != ""
}

// ResultadoEnvio traz os dados úteis da resposta de envio.
type ResultadoEnvio struct {
	WaMessageID    string `json:"wa_message_id"`
	ConversationID int64  `json:"conversation_id"`
	Status         string `json:"status"`
}

func (z *ZapGovClient) garantirToken() (string, error) {
	z.mu.Lock()
	defer z.mu.Unlock()
	if z.token != "" && time.Now().Before(z.expiraEm) {
		return z.token, nil
	}
	body, _ := json.Marshal(map[string]string{
		"email":      z.cfg.ZapGovEmail,
		"senha":      z.cfg.ZapGovSenha,
		"subdominio": z.cfg.ZapGovSubdominio,
	})
	resp, err := z.http.Post(z.cfg.ZapGovBaseURL+"/auth/login",
		"application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login ZapGov falhou (%d): %s", resp.StatusCode, string(b))
	}
	var r struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil || r.Token == "" {
		return "", fmt.Errorf("login ZapGov: token ausente")
	}
	z.token = r.Token
	z.expiraEm = time.Now().Add(11 * time.Hour) // margem sobre as 12h
	return z.token, nil
}

func (z *ZapGovClient) postAutenticado(caminho string, payload any) (*http.Response, error) {
	token, err := z.garantirToken()
	if err != nil {
		return nil, err
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, z.cfg.ZapGovBaseURL+caminho, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := z.http.Do(req)
	if err != nil {
		return nil, err
	}
	// Token pode ter expirado antes da margem: tenta uma vez de novo.
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		z.mu.Lock()
		z.token = ""
		z.mu.Unlock()
		token, err = z.garantirToken()
		if err != nil {
			return nil, err
		}
		req2, _ := http.NewRequest(http.MethodPost, z.cfg.ZapGovBaseURL+caminho, bytes.NewReader(body))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", "Bearer "+token)
		resp, err = z.http.Do(req2)
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}

// Enviar manda texto livre (dentro da janela de 24h). Se a janela estiver
// fechada, devolve ErrJanelaFechada.
func (z *ZapGovClient) Enviar(to, body, replyTo string) (*ResultadoEnvio, error) {
	payload := map[string]string{"to": to, "body": body}
	if replyTo != "" {
		payload["reply_to"] = replyTo
	}
	resp, err := z.postAutenticado("/messages/send", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnprocessableEntity {
		var e struct {
			WindowClosed bool `json:"window_closed"`
		}
		raw, _ := io.ReadAll(resp.Body)
		_ = json.Unmarshal(raw, &e)
		if e.WindowClosed {
			return nil, ErrJanelaFechada
		}
		return nil, fmt.Errorf("envio recusado: %s", string(raw))
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("envio falhou (%d): %s", resp.StatusCode, string(raw))
	}
	var r ResultadoEnvio
	_ = json.NewDecoder(resp.Body).Decode(&r)
	return &r, nil
}

// ---- Entrada (SIGE puxa as mensagens recebidas do WhatsApp) ----

// MensagemZapGov representa uma mensagem de uma conversa no ZapGov.
type MensagemZapGov struct {
	ID             string `json:"id"`
	WaMessageID    string `json:"wa_message_id"`
	ConversationID int64  `json:"conversation_id"`
	// direction/from: "in" (do contato) ou "out" (da equipe). Aceita variações.
	Direction string `json:"direction"`
	From      string `json:"from"`
	Telefone  string `json:"telefone"`
	To        string `json:"to"`
	Nome      string `json:"nome"`
	ContatoNome string `json:"contato_nome"`
	Body      string `json:"body"`
	Texto     string `json:"texto"`
	MediaURL  string `json:"media_url"`
	MidiaURL  string `json:"midia_url"`
	CriadoEm  string `json:"created_at"`
}

// Recebida indica se a mensagem veio do contato (entrada), e não da equipe.
func (m MensagemZapGov) Recebida() bool {
	d := strings.ToLower(strings.TrimSpace(m.Direction + m.From))
	// Considera entrada quando marcada como "in"/"inbound"/"contato"; mensagens
	// de saída trazem "out"/"outbound"/"agent".
	if strings.Contains(d, "out") || strings.Contains(d, "agent") || strings.Contains(d, "equipe") {
		return false
	}
	return true
}

func (m MensagemZapGov) Conteudo() string {
	if strings.TrimSpace(m.Body) != "" {
		return m.Body
	}
	return m.Texto
}

func (m MensagemZapGov) Midia() string {
	if strings.TrimSpace(m.MediaURL) != "" {
		return m.MediaURL
	}
	return m.MidiaURL
}

func (m MensagemZapGov) TelefoneContato() string {
	for _, v := range []string{m.Telefone, m.From} {
		if s := strings.TrimSpace(v); s != "" && !strings.Contains(strings.ToLower(s), "out") {
			return s
		}
	}
	return strings.TrimSpace(m.To)
}

func (m MensagemZapGov) NomeContato() string {
	if strings.TrimSpace(m.Nome) != "" {
		return m.Nome
	}
	return m.ContatoNome
}

func (z *ZapGovClient) getAutenticado(caminho string) (*http.Response, error) {
	token, err := z.garantirToken()
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest(http.MethodGet, z.cfg.ZapGovBaseURL+caminho, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := z.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		z.mu.Lock()
		z.token = ""
		z.mu.Unlock()
		token, err = z.garantirToken()
		if err != nil {
			return nil, err
		}
		req2, _ := http.NewRequest(http.MethodGet, z.cfg.ZapGovBaseURL+caminho, nil)
		req2.Header.Set("Authorization", "Bearer "+token)
		resp, err = z.http.Do(req2)
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}

// MensagensConversa busca as mensagens de uma conversa (opcionalmente só as não
// lidas), usada para trazer o que chegou do contato no WhatsApp.
func (z *ZapGovClient) MensagensConversa(convID int64, limite int, apenasNaoLidas bool) ([]MensagemZapGov, error) {
	if limite <= 0 {
		limite = 200
	}
	q := url.Values{}
	q.Set("limit", fmt.Sprintf("%d", limite))
	if apenasNaoLidas {
		q.Set("read", "0")
	}
	resp, err := z.getAutenticado(fmt.Sprintf("/conversations/%d/messages?%s", convID, q.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("conversa %d falhou (%d): %s", convID, resp.StatusCode, string(raw))
	}
	// Aceita tanto um array puro quanto {data:[...]} ou {messages:[...]}.
	raw, _ := io.ReadAll(resp.Body)
	// DEBUG temporário: payload cru da conversa para ajustar os nomes de campo.
	log.Printf("[zapgov][conv %d] %s", convID, string(raw))
	var arr []MensagemZapGov
	if err := json.Unmarshal(raw, &arr); err == nil && arr != nil {
		return arr, nil
	}
	var env struct {
		Data     []MensagemZapGov `json:"data"`
		Messages []MensagemZapGov `json:"messages"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("resposta de conversa inesperada: %s", string(raw))
	}
	if env.Messages != nil {
		return env.Messages, nil
	}
	return env.Data, nil
}

// EventoZapGov é um evento recebido pelo SSE.
type EventoZapGov struct {
	Type    string          `json:"type"`
	Event   string          `json:"event"`
	Message json.RawMessage `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// StreamEventos conecta no SSE do ZapGov e chama onEvento para cada evento até o
// contexto ser cancelado ou a conexão cair (o chamador decide reconectar).
func (z *ZapGovClient) StreamEventos(ctx context.Context, onEvento func(EventoZapGov)) error {
	token, err := z.garantirToken()
	if err != nil {
		return err
	}
	q := url.Values{}
	q.Set("token", token)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		z.cfg.ZapGovBaseURL+"/events?"+q.Encode(), nil)
	req.Header.Set("Accept", "text/event-stream")
	// Stream não usa o timeout curto do cliente padrão.
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SSE falhou (%d): %s", resp.StatusCode, string(raw))
	}

	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	var dados strings.Builder
	for sc.Scan() {
		linha := sc.Text()
		if linha == "" { // fim de um evento
			bruto := strings.TrimSpace(dados.String())
			dados.Reset()
			if bruto == "" || bruto == ":" {
				continue
			}
			var ev EventoZapGov
			if err := json.Unmarshal([]byte(bruto), &ev); err == nil {
				// Guarda o payload cru quando não veio embrulhado em message/data.
				if len(ev.Message) == 0 && len(ev.Data) == 0 {
					ev.Data = json.RawMessage(bruto)
				}
				onEvento(ev)
			}
			continue
		}
		if strings.HasPrefix(linha, ":") { // comentário/keep-alive
			continue
		}
		if strings.HasPrefix(linha, "data:") {
			dados.WriteString(strings.TrimSpace(strings.TrimPrefix(linha, "data:")))
		}
	}
	return sc.Err()
}

// EnviarTemplate manda um template aprovado (usado fora da janela de 24h).
func (z *ZapGovClient) EnviarTemplate(to, nome, template, language string, params []string) (*ResultadoEnvio, error) {
	if language == "" {
		language = "pt_BR"
	}
	payload := map[string]any{
		"to": to, "nome": nome, "template": template,
		"language": language, "params": params,
	}
	resp, err := z.postAutenticado("/messages/template", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("template falhou (%d): %s", resp.StatusCode, string(raw))
	}
	var r ResultadoEnvio
	_ = json.NewDecoder(resp.Body).Decode(&r)
	return &r, nil
}
