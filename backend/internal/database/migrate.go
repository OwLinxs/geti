package database

import (
	"fmt"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

// AutoMigrate cria/atualiza o schema do banco a partir dos models.
// A ordem respeita dependências de chaves estrangeiras.
//
// Optamos por GORM AutoMigrate pela simplicidade nesta fase. Caso a
// complexidade cresça (ex.: migrações de dados, renomeações destrutivas),
// a estrutura está preparada para migrar para golang-migrate sem reescrever
// os models.
func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Usuario{},
		&models.Categoria{},
		&models.Setor{},
		&models.Servidor{},
		&models.Item{},
		&models.Movimentacao{},
		&models.TermoResponsabilidade{},
		&models.RegistroAuditoria{},
	)
	if err != nil {
		return fmt.Errorf("falha na migração automática: %w", err)
	}
	return nil
}
