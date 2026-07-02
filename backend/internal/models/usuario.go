package models

// Perfil define o nível de acesso do usuário do sistema.
type Perfil string

const (
	PerfilAdministrador Perfil = "administrador"
	PerfilOperador      Perfil = "operador"
)

// PerfilValido verifica se o perfil informado é aceito.
func PerfilValido(p Perfil) bool {
	return p == PerfilAdministrador || p == PerfilOperador
}

// Usuario representa um usuário autenticável do sistema (não confundir com
// Servidor, que é a pessoa responsável por um equipamento).
//
// LGPD: armazenamos apenas o necessário para autenticação e auditoria
// (nome, e-mail funcional e hash de senha). Nenhum dado sensível é coletado.
type Usuario struct {
	Base
	Nome      string `gorm:"size:120;not null" json:"nome"`
	Email     string `gorm:"size:160;uniqueIndex;not null" json:"email"`
	SenhaHash string `gorm:"size:255;not null" json:"-"` // nunca serializado
	Perfil    Perfil `gorm:"size:20;not null;index" json:"perfil"`
	Ativo     bool   `gorm:"not null;default:true" json:"ativo"`
}

func (Usuario) TableName() string { return "usuarios" }
