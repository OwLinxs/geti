package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pmfb/sige-ti/internal/models"
)

// AuthService emite e valida tokens JWT.
type AuthService struct {
	usuarioService *UsuarioService
	segredo        []byte
	expiraEm       time.Duration
}

func NewAuthService(usuarioService *UsuarioService, segredo string, expiraEm time.Duration) *AuthService {
	return &AuthService{
		usuarioService: usuarioService,
		segredo:        []byte(segredo),
		expiraEm:       expiraEm,
	}
}

// Claims carregadas no token. Mantemos o mínimo necessário (id, perfil, nome).
type Claims struct {
	UsuarioID uint          `json:"uid"`
	Perfil    models.Perfil `json:"perfil"`
	Nome      string        `json:"nome"`
	jwt.RegisteredClaims
}

// ResultadoLogin é devolvido após autenticação bem-sucedida.
type ResultadoLogin struct {
	Token     string         `json:"token"`
	ExpiraEm  time.Time      `json:"expira_em"`
	Usuario   *models.Usuario `json:"usuario"`
}

// Login autentica e emite o token.
func (s *AuthService) Login(email, senha string) (*ResultadoLogin, error) {
	u, err := s.usuarioService.Autenticar(email, senha)
	if err != nil {
		return nil, err
	}
	token, exp, err := s.gerarToken(u)
	if err != nil {
		return nil, err
	}
	return &ResultadoLogin{Token: token, ExpiraEm: exp, Usuario: u}, nil
}

func (s *AuthService) gerarToken(u *models.Usuario) (string, time.Time, error) {
	exp := time.Now().UTC().Add(s.expiraEm)
	claims := Claims{
		UsuarioID: u.ID,
		Perfil:    u.Perfil,
		Nome:      u.Nome,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.Email,
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(exp),
			Issuer:    "sige-ti",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	assinado, err := token.SignedString(s.segredo)
	if err != nil {
		return "", time.Time{}, err
	}
	return assinado, exp, nil
}

// ValidarToken verifica a assinatura/validade e devolve as claims.
func (s *AuthService) ValidarToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de assinatura inesperado")
		}
		return s.segredo, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrNaoAutorizado
	}
	return claims, nil
}
