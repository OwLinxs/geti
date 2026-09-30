package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pmfb/sige-ti/internal/config"
)

// ErrJanelaFechada indica que a janela de 24h do WhatsApp está fechada: o envio
// de texto livre foi recusado e é preciso usar um template aprovado.
var ErrJanelaFechada = errors.New("janela de 24h fechada")

// ZapGovClient fala com o gateway de WhatsApp (ZapGov). Autenticação por API key
// no header X-API-Key. Endpoints: POST /v1/mensagens e POST /v1/relay/encerrar.
type ZapGovClient struct {
	cfg  *config.Config
	http *http.Client
}

func NewZapGovClient(cfg *config.Config) *ZapGovClient {
	return &ZapGovClient{cfg: cfg, http: &http.Client{Timeout: 20 * time.Second}}
}

// Configurado informa se há credenciais para falar com o ZapGov.
func (z *ZapGovClient) Configurado() bool {
	return z.cfg.ZapGovBaseURL != "" && z.cfg.ZapGovAPIKey != ""
}

func (z *ZapGovClient) post(caminho string, payload any) (*http.Response, error) {
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, z.cfg.ZapGovBaseURL+caminho, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", z.cfg.ZapGovAPIKey)
	return z.http.Do(req)
}

// enviar despacha o payload para /v1/mensagens e devolve o wa_message_id.
// Trata 422 {"erro":"janela_fechada"} como ErrJanelaFechada.
func (z *ZapGovClient) enviar(payload map[string]any) (string, error) {
	resp, err := z.post("/v1/mensagens", payload)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusUnprocessableEntity {
		var e struct {
			Erro string `json:"erro"`
		}
		_ = json.Unmarshal(raw, &e)
		if e.Erro == "janela_fechada" {
			return "", ErrJanelaFechada
		}
		return "", fmt.Errorf("envio recusado: %s", string(raw))
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("envio falhou (%d): %s", resp.StatusCode, string(raw))
	}
	var r struct {
		WaMessageID string `json:"wa_message_id"`
	}
	_ = json.Unmarshal(raw, &r)
	return r.WaMessageID, nil
}

// EnviarTexto manda texto livre (dentro da janela de 24h).
func (z *ZapGovClient) EnviarTexto(telefone, texto string) (string, error) {
	return z.enviar(map[string]any{
		"telefone": telefone,
		"tipo":     "texto",
		"texto":    texto,
	})
}

// EnviarMidia manda um arquivo (a URL precisa ser acessível pelo ZapGov).
func (z *ZapGovClient) EnviarMidia(telefone, midiaURL, mime string) (string, error) {
	return z.enviar(map[string]any{
		"telefone":   telefone,
		"tipo":       "midia",
		"midia_url":  midiaURL,
		"midia_mime": mime,
	})
}

// EnviarTemplate manda um template aprovado (usado fora da janela de 24h).
func (z *ZapGovClient) EnviarTemplate(telefone, template string, params []string) (string, error) {
	return z.enviar(map[string]any{
		"telefone": telefone,
		"tipo":     "template",
		"template": template,
		"params":   params,
	})
}

// RelayEncerrar tira a conversa do modo relay: a próxima mensagem da pessoa
// volta a cair no menu do bot. Chamado quando o chamado é fechado no SIGE.
func (z *ZapGovClient) RelayEncerrar(telefone string) error {
	resp, err := z.post("/v1/relay/encerrar", map[string]any{"telefone": telefone})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("relay/encerrar falhou (%d): %s", resp.StatusCode, string(raw))
	}
	return nil
}
