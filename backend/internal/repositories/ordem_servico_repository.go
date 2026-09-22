package repositories

import (
	"errors"
	"fmt"
	"time"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

// FiltroOrdemServico concentra os filtros de consulta da fila de manutenção.
type FiltroOrdemServico struct {
	Status             string
	TecnicoID          *uint
	CategoriaChamadoID *uint
	AbertoPorID        *uint // filtra pelos chamados abertos por um usuário (portal)
	De                 *time.Time
	Ate                *time.Time
	Pagina             int
	Tamanho            int
}

// ContagemRotulo é um par nome/total para agregações do dashboard.
type ContagemRotulo struct {
	Rotulo string `json:"rotulo"`
	Total  int64  `json:"total"`
}

// DashboardChamados agrega indicadores do módulo de chamados.
type DashboardChamados struct {
	AbertosTotal             int64            `json:"abertos_total"`
	PorStatus                []ContagemRotulo `json:"por_status"`
	PorPrioridade            []ContagemRotulo `json:"por_prioridade"`
	PorCategoria             []ContagemRotulo `json:"por_categoria"`
	SemTecnico               int64            `json:"sem_tecnico"`
	RespostaAtrasada         int64            `json:"resposta_atrasada"`
	ResolucaoAtrasada        int64            `json:"resolucao_atrasada"`
	ConcluidosUltimos30      int64            `json:"concluidos_ultimos_30"`
	TempoMedioResolucaoHoras float64          `json:"tempo_medio_resolucao_horas"`
	NotaMedia                float64          `json:"nota_media"`
	TotalAvaliacoes          int64            `json:"total_avaliacoes"`
}

type OrdemServicoRepository interface {
	CriarComTx(tx *gorm.DB, os *models.OrdemServico) error
	Dashboard() (DashboardChamados, error)
	Atualizar(os *models.OrdemServico) error
	BuscarPorID(id uint) (*models.OrdemServico, error)
	// BuscarPorReferenciaExterna localiza a OS pelo id do card externo (sync).
	BuscarPorReferenciaExterna(ref string) (*models.OrdemServico, error)
	Listar(f FiltroOrdemServico) ([]models.OrdemServico, int64, error)
	Remover(id uint) error
	// SubstituirPassos troca todos os passos de uma OS numa transação.
	SubstituirPassos(osID uint, passos []models.OrdemServicoPasso) error
	// DefinirPrimeiraResposta marca a 1ª resposta da equipe (só se ainda nula).
	DefinirPrimeiraResposta(osID uint, t time.Time) error
	// SLAEstourado devolve chamados ativos que furaram o prazo e ainda não
	// foram notificados. tipo = "resposta" ou "resolucao".
	SLAEstourado(tipo string, agora time.Time) ([]models.OrdemServico, error)
	// MarcarSLANotificado marca o chamado como já notificado para aquele tipo.
	MarcarSLANotificado(osID uint, tipo string) error
	// ProximoNumero gera o próximo número sequencial no formato OS-AAAA-NNNN.
	ProximoNumero(tx *gorm.DB, ano int) (string, error)
	DB() *gorm.DB
}

type ordemServicoRepository struct {
	db *gorm.DB
}

func NewOrdemServicoRepository(db *gorm.DB) OrdemServicoRepository {
	return &ordemServicoRepository{db: db}
}

func (r *ordemServicoRepository) DB() *gorm.DB { return r.db }

func (r *ordemServicoRepository) preloads(q *gorm.DB) *gorm.DB {
	return q.
		Preload("Item").
		Preload("Setor").
		Preload("CategoriaChamado").
		Preload("Solicitante").
		Preload("Tecnico").
		Preload("AbertoPor").
		Preload("Passos", func(db *gorm.DB) *gorm.DB {
			return db.Order("ordem ASC, id ASC")
		})
}

func (r *ordemServicoRepository) CriarComTx(tx *gorm.DB, os *models.OrdemServico) error {
	return tx.Create(os).Error
}

func (r *ordemServicoRepository) Atualizar(os *models.OrdemServico) error {
	return r.db.Save(os).Error
}

func (r *ordemServicoRepository) BuscarPorID(id uint) (*models.OrdemServico, error) {
	var os models.OrdemServico
	if err := r.preloads(r.db).First(&os, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return &os, nil
}

func (r *ordemServicoRepository) BuscarPorReferenciaExterna(ref string) (*models.OrdemServico, error) {
	if ref == "" {
		return nil, ErrNaoEncontrado
	}
	var os models.OrdemServico
	if err := r.preloads(r.db).Where("referencia_externa = ?", ref).First(&os).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return &os, nil
}

func (r *ordemServicoRepository) Listar(f FiltroOrdemServico) ([]models.OrdemServico, int64, error) {
	q := r.db.Model(&models.OrdemServico{})

	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.TecnicoID != nil {
		q = q.Where("tecnico_id = ?", *f.TecnicoID)
	}
	if f.AbertoPorID != nil {
		q = q.Where("aberto_por_id = ?", *f.AbertoPorID)
	}
	if f.CategoriaChamadoID != nil {
		q = q.Where("categoria_chamado_id = ?", *f.CategoriaChamadoID)
	}
	if f.De != nil {
		q = q.Where("data_abertura >= ?", *f.De)
	}
	if f.Ate != nil {
		q = q.Where("data_abertura <= ?", *f.Ate)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if f.Tamanho <= 0 {
		f.Tamanho = 20
	}
	if f.Pagina <= 0 {
		f.Pagina = 1
	}
	offset := (f.Pagina - 1) * f.Tamanho

	var lista []models.OrdemServico
	err := r.preloads(q).
		Order("data_abertura DESC, id DESC").
		Limit(f.Tamanho).Offset(offset).
		Find(&lista).Error
	if err != nil {
		return nil, 0, err
	}
	return lista, total, nil
}

func (r *ordemServicoRepository) Remover(id uint) error {
	return r.db.Delete(&models.OrdemServico{}, id).Error
}

func (r *ordemServicoRepository) SubstituirPassos(osID uint, passos []models.OrdemServicoPasso) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("ordem_servico_id = ?", osID).
			Delete(&models.OrdemServicoPasso{}).Error; err != nil {
			return err
		}
		if len(passos) == 0 {
			return nil
		}
		for i := range passos {
			passos[i].OrdemServicoID = osID
		}
		return tx.Create(&passos).Error
	})
}

func (r *ordemServicoRepository) Dashboard() (DashboardChamados, error) {
	var d DashboardChamados
	ativos := []models.StatusOS{models.OSAberta, models.OSEmAndamento, models.OSAguardandoPeca}
	agora := time.Now().UTC()
	base := func() *gorm.DB { return r.db.Model(&models.OrdemServico{}) }

	base().Where("status IN ?", ativos).Count(&d.AbertosTotal)

	// Por status (todos os status).
	{
		var rows []struct {
			Status string
			Total  int64
		}
		base().Select("status, count(*) as total").Group("status").Scan(&rows)
		for _, x := range rows {
			d.PorStatus = append(d.PorStatus, ContagemRotulo{Rotulo: x.Status, Total: x.Total})
		}
	}
	// Por prioridade (apenas abertos).
	{
		var rows []struct {
			Prioridade string
			Total      int64
		}
		base().Select("prioridade, count(*) as total").
			Where("status IN ?", ativos).Group("prioridade").Scan(&rows)
		for _, x := range rows {
			d.PorPrioridade = append(d.PorPrioridade, ContagemRotulo{Rotulo: x.Prioridade, Total: x.Total})
		}
	}
	// Por categoria (abertos) — junta o nome.
	{
		var rows []struct {
			Nome  string
			Total int64
		}
		r.db.Model(&models.OrdemServico{}).
			Select("COALESCE(categorias_chamado.nome, 'Sem categoria') as nome, count(*) as total").
			Joins("LEFT JOIN categorias_chamado ON categorias_chamado.id = ordens_servico.categoria_chamado_id").
			Where("ordens_servico.status IN ?", ativos).
			Group("nome").Scan(&rows)
		for _, x := range rows {
			d.PorCategoria = append(d.PorCategoria, ContagemRotulo{Rotulo: x.Nome, Total: x.Total})
		}
	}

	base().Where("status IN ? AND tecnico_id IS NULL", ativos).Count(&d.SemTecnico)
	base().Where("status IN ? AND primeira_resposta_em IS NULL AND prazo_resposta_em IS NOT NULL AND prazo_resposta_em < ?", ativos, agora).
		Count(&d.RespostaAtrasada)
	base().Where("status IN ? AND prazo_resolucao_em IS NOT NULL AND prazo_resolucao_em < ?", ativos, agora).
		Count(&d.ResolucaoAtrasada)

	trintaDias := agora.Add(-30 * 24 * time.Hour)
	base().Where("status = ? AND data_conclusao >= ?", models.OSConcluida, trintaDias).
		Count(&d.ConcluidosUltimos30)

	// Tempo médio de resolução (horas) dos concluídos com data de conclusão.
	{
		var concluidos []models.OrdemServico
		r.db.Select("data_abertura, data_conclusao").
			Where("status = ? AND data_conclusao IS NOT NULL", models.OSConcluida).
			Find(&concluidos)
		if len(concluidos) > 0 {
			var soma float64
			for _, os := range concluidos {
				if os.DataConclusao != nil {
					soma += os.DataConclusao.Sub(os.DataAbertura).Hours()
				}
			}
			d.TempoMedioResolucaoHoras = soma / float64(len(concluidos))
		}
	}

	// Avaliações.
	base().Where("avaliacao_nota IS NOT NULL").Count(&d.TotalAvaliacoes)
	if d.TotalAvaliacoes > 0 {
		var media *float64
		r.db.Model(&models.OrdemServico{}).
			Where("avaliacao_nota IS NOT NULL").
			Select("AVG(avaliacao_nota)").Scan(&media)
		if media != nil {
			d.NotaMedia = *media
		}
	}

	return d, nil
}

func (r *ordemServicoRepository) DefinirPrimeiraResposta(osID uint, t time.Time) error {
	return r.db.Model(&models.OrdemServico{}).
		Where("id = ? AND primeira_resposta_em IS NULL", osID).
		Update("primeira_resposta_em", t).Error
}

func (r *ordemServicoRepository) SLAEstourado(tipo string, agora time.Time) ([]models.OrdemServico, error) {
	ativos := []models.StatusOS{models.OSAberta, models.OSEmAndamento, models.OSAguardandoPeca}
	q := r.db.Model(&models.OrdemServico{}).Where("status IN ?", ativos)
	if tipo == "resposta" {
		q = q.Where("primeira_resposta_em IS NULL AND prazo_resposta_em IS NOT NULL AND prazo_resposta_em < ? AND sla_resposta_notificada = ?", agora, false)
	} else {
		q = q.Where("prazo_resolucao_em IS NOT NULL AND prazo_resolucao_em < ? AND sla_resolucao_notificada = ?", agora, false)
	}
	var lista []models.OrdemServico
	err := q.Find(&lista).Error
	return lista, err
}

func (r *ordemServicoRepository) MarcarSLANotificado(osID uint, tipo string) error {
	coluna := "sla_resolucao_notificada"
	if tipo == "resposta" {
		coluna = "sla_resposta_notificada"
	}
	return r.db.Model(&models.OrdemServico{}).Where("id = ?", osID).
		Update(coluna, true).Error
}

func (r *ordemServicoRepository) ProximoNumero(tx *gorm.DB, ano int) (string, error) {
	var count int64
	inicio := time.Date(ano, 1, 1, 0, 0, 0, 0, time.UTC)
	fim := time.Date(ano+1, 1, 1, 0, 0, 0, 0, time.UTC)
	// Conta inclusive registros soft-deletados para não reaproveitar números.
	err := tx.Unscoped().Model(&models.OrdemServico{}).
		Where("data_abertura >= ? AND data_abertura < ?", inicio, fim).
		Count(&count).Error
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("OS-%04d-%04d", ano, count+1), nil
}
