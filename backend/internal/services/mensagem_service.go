package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pmfb/sige-ti/internal/config"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
)

// tamanhoMaxAnexo limita o tamanho de um anexo (10 MB).
const tamanhoMaxAnexo = 10 << 20

// MensagemService gerencia a conversa (mensagens) vinculada a um chamado/OS.
type MensagemService struct {
	repo   repositories.MensagemRepository
	osRepo repositories.OrdemServicoRepository
	cfg    *config.Config
}

func NewMensagemService(
	repo repositories.MensagemRepository,
	osRepo repositories.OrdemServicoRepository,
	cfg *config.Config,
) *MensagemService {
	return &MensagemService{repo: repo, osRepo: osRepo, cfg: cfg}
}

func (s *MensagemService) ListarPorOS(osID uint) ([]models.MensagemChamado, error) {
	if _, err := s.osRepo.BuscarPorID(osID); err != nil {
		return nil, traduzErroRepo(err)
	}
	return s.repo.ListarPorOS(osID)
}

// ListarPublicasPorOS devolve a conversa SEM as notas internas da equipe —
// usada no portal do solicitante.
func (s *MensagemService) ListarPublicasPorOS(osID uint) ([]models.MensagemChamado, error) {
	if _, err := s.osRepo.BuscarPorID(osID); err != nil {
		return nil, traduzErroRepo(err)
	}
	return s.repo.ListarPublicasPorOS(osID)
}

// EntradaMensagem reúne os dados de uma mensagem enviada no chamado.
type EntradaMensagem struct {
	OrdemServicoID uint
	AutorID        uint
	AutorNome      string
	AutorTipo      models.AutorTipo // tecnico (equipe) ou servidor (solicitante)
	Texto          string
	Interna        bool                  // nota interna — só faz sentido para a equipe
	Arquivo        *multipart.FileHeader // opcional
}

// Enviar registra uma mensagem da equipe no chamado (com anexo opcional).
func (s *MensagemService) Enviar(in EntradaMensagem) (*models.MensagemChamado, error) {
	if _, err := s.osRepo.BuscarPorID(in.OrdemServicoID); err != nil {
		return nil, traduzErroRepo(err)
	}

	texto := strings.TrimSpace(in.Texto)
	if texto == "" && in.Arquivo == nil {
		ev := NovoErroValidacao()
		ev.Add("texto", "Escreva uma mensagem ou anexe um arquivo.")
		return nil, ev
	}

	autorTipo := in.AutorTipo
	if autorTipo == "" {
		autorTipo = models.AutorTecnico
	}

	// Direção e nota interna dependem de quem escreve. Solicitante nunca cria
	// nota interna; sua mensagem é sempre "entrada".
	interna := in.Interna
	var direcao models.DirecaoMensagem
	switch autorTipo {
	case models.AutorServidor:
		direcao = models.MsgEntrada
		interna = false
	default:
		if interna {
			direcao = models.MsgInterna
		} else {
			direcao = models.MsgSaida
		}
	}

	msg := &models.MensagemChamado{
		OrdemServicoID: in.OrdemServicoID,
		Direcao:        direcao,
		AutorTipo:      autorTipo,
		AutorNome:      in.AutorNome,
		Texto:          texto,
		Interna:        interna,
		Status:         models.MsgEnviada,
		EnviadaEm:      time.Now().UTC(),
	}
	if in.AutorID != 0 {
		id := in.AutorID
		msg.AutorID = &id
	}

	if in.Arquivo != nil {
		if err := s.anexar(msg, in.Arquivo); err != nil {
			return nil, err
		}
	}

	if err := s.repo.Criar(msg); err != nil {
		return nil, err
	}
	return s.repo.BuscarPorID(msg.ID)
}

// anexar valida e grava o arquivo em disco, preenchendo os campos de anexo.
func (s *MensagemService) anexar(msg *models.MensagemChamado, fh *multipart.FileHeader) error {
	if fh.Size > tamanhoMaxAnexo {
		ev := NovoErroValidacao()
		ev.Add("arquivo", "Arquivo excede o limite de 10 MB.")
		return ev
	}

	src, err := fh.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dir := s.cfg.AnexosDir
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	nome, err := nomeAleatorio(filepath.Ext(fh.Filename))
	if err != nil {
		return err
	}
	caminho := filepath.Join(dir, nome)

	dst, err := os.Create(caminho)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copia com teto de segurança (evita estouro além do limite declarado).
	n, err := io.Copy(dst, io.LimitReader(src, tamanhoMaxAnexo+1))
	if err != nil {
		os.Remove(caminho)
		return err
	}
	if n > tamanhoMaxAnexo {
		os.Remove(caminho)
		ev := NovoErroValidacao()
		ev.Add("arquivo", "Arquivo excede o limite de 10 MB.")
		return ev
	}

	msg.AnexoNome = filepath.Base(fh.Filename)
	msg.AnexoTipo = fh.Header.Get("Content-Type")
	msg.AnexoCaminho = caminho
	msg.AnexoTamanho = n
	return nil
}

// BuscarAnexo devolve os metadados + caminho do anexo de uma mensagem, validando
// que ela pertence à OS informada.
func (s *MensagemService) BuscarAnexo(osID, msgID uint) (*models.MensagemChamado, error) {
	m, err := s.repo.BuscarPorID(msgID)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	if m.OrdemServicoID != osID || !m.TemAnexo() {
		return nil, ErrNaoEncontrado
	}
	return m, nil
}

func nomeAleatorio(ext string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	ext = strings.ToLower(ext)
	// Sanitiza a extensão (evita path traversal / nomes estranhos).
	if len(ext) > 10 || strings.ContainsAny(ext, "/\\") {
		ext = ""
	}
	return fmt.Sprintf("%s%s", hex.EncodeToString(b), ext), nil
}
