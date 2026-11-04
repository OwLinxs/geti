package services

import (
	"fmt"
	"sort"
	"time"
)

// EventoTimeline é um evento unificado da linha do tempo de um item: pode ter
// origem em uma movimentação, uma ordem de serviço ou um termo de responsabilidade.
type EventoTimeline struct {
	Tipo    string    `json:"tipo"` // "movimentacao" | "os" | "termo"
	Data    time.Time `json:"data"`
	Titulo  string    `json:"titulo"`
	Detalhe string    `json:"detalhe,omitempty"`
	RefID   uint      `json:"ref_id"` // id do registro de origem (para navegação)
}

// Timeline monta a linha do tempo consolidada do item: movimentações + ordens de
// serviço + termos, em ordem cronológica decrescente.
func (s *ItemService) Timeline(itemID uint) ([]EventoTimeline, error) {
	if _, err := s.repo.BuscarPorID(itemID); err != nil {
		return nil, traduzErroRepo(err)
	}

	eventos := []EventoTimeline{}

	// Movimentações.
	movs, err := s.movRepo.ListarPorItem(itemID)
	if err != nil {
		return nil, err
	}
	for _, m := range movs {
		detalhe := m.Observacao
		if m.OrigemDescricao != "" || m.DestinoDescricao != "" {
			local := m.OrigemDescricao
			if m.DestinoDescricao != "" {
				if local != "" {
					local += " → " + m.DestinoDescricao
				} else {
					local = m.DestinoDescricao
				}
			}
			if detalhe != "" {
				detalhe = local + " · " + detalhe
			} else {
				detalhe = local
			}
		}
		eventos = append(eventos, EventoTimeline{
			Tipo:    "movimentacao",
			Data:    m.DataEvento,
			Titulo:  rotuloMov(m.Tipo),
			Detalhe: detalhe,
			RefID:   m.ID,
		})
	}

	// Ordens de serviço (manutenção) — dependência opcional.
	if s.osRepo != nil {
		oss, err := s.osRepo.ListarPorItem(itemID)
		if err != nil {
			return nil, err
		}
		for _, os := range oss {
			status := rotuloStatusOS[os.Status]
			if status == "" {
				status = string(os.Status)
			}
			detalhe := os.Assunto
			if detalhe == "" {
				detalhe = os.DefeitoRelatado
			}
			eventos = append(eventos, EventoTimeline{
				Tipo:    "os",
				Data:    os.DataAbertura,
				Titulo:  fmt.Sprintf("OS %s · %s", os.Numero, status),
				Detalhe: detalhe,
				RefID:   os.ID,
			})
		}
	}

	// Termos de responsabilidade.
	termos, err := s.termoRepo.ListarPorItem(itemID)
	if err != nil {
		return nil, err
	}
	for _, t := range termos {
		nome := t.ServidorNomeSnapshot
		if nome == "" && t.Servidor != nil {
			nome = t.Servidor.Nome
		}
		eventos = append(eventos, EventoTimeline{
			Tipo:    "termo",
			Data:    t.DataEmissao,
			Titulo:  fmt.Sprintf("Termo %s", t.Numero),
			Detalhe: "Entregue a " + nome,
			RefID:   t.ID,
		})
	}

	// Ordena do mais recente ao mais antigo.
	sort.Slice(eventos, func(i, j int) bool {
		return eventos[i].Data.After(eventos[j].Data)
	})

	return eventos, nil
}
