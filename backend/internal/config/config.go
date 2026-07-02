package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config concentra todas as configurações da aplicação, carregadas de
// variáveis de ambiente. Mantém o sistema preparado para troca de banco
// (SQLite <-> PostgreSQL) e parametrização do termo de responsabilidade.
type Config struct {
	// Servidor
	Port    string
	GinMode string

	// Banco de dados
	DBDriver string // "sqlite" ou "postgres"
	DBDSN    string // string de conexão (caminho do arquivo ou DSN do postgres)

	// Autenticação JWT
	JWTSecret    string
	JWTExpiresIn time.Duration

	// Seed / primeira execução
	SeedDemo   bool   // insere itens/movimentações fictícios (apenas dev)
	AdminEmail string // e-mail do admin criado no seed base
	AdminSenha string // senha do admin criado no seed base
	AdminNome  string // nome do admin criado no seed base

	// Rate limiting do login (mitigação de força bruta).
	LoginRateLimite int           // tentativas permitidas por janela, por IP
	LoginRateJanela time.Duration // duração da janela

	// Parametrização do termo de responsabilidade (PDF)
	PrefeituraNome      string
	PrefeituraDepto     string
	PrefeituraLogoPath  string // caminho para o brasão/logo (opcional)
	TermoCabecalho      string
	TermoCidadeUF       string

	// CORS
	CORSAllowedOrigins []string
}

// Load lê as variáveis de ambiente e devolve a configuração validada.
// Aplica defaults seguros para desenvolvimento; valores sensíveis em
// produção devem vir sempre do ambiente.
func Load() (*Config, error) {
	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		GinMode:            getEnv("GIN_MODE", "debug"),
		DBDriver:           strings.ToLower(getEnv("DB_DRIVER", "sqlite")),
		DBDSN:              getEnv("DB_DSN", "sige-ti.db"),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		SeedDemo:           strings.EqualFold(getEnv("SEED_DEMO", "false"), "true"),
		AdminEmail:         strings.ToLower(strings.TrimSpace(getEnv("ADMIN_EMAIL", "admin@sige-ti.local"))),
		AdminSenha:         getEnv("ADMIN_SENHA", ""),
		AdminNome:          getEnv("ADMIN_NOME", "Administrador do Sistema"),
		PrefeituraNome:     getEnv("PREFEITURA_NOME", "Prefeitura Municipal"),
		PrefeituraDepto:    getEnv("PREFEITURA_DEPTO", "Departamento de Tecnologia da Informação"),
		PrefeituraLogoPath: getEnv("PREFEITURA_LOGO_PATH", ""),
		TermoCabecalho:     getEnv("TERMO_CABECALHO", "Termo de Responsabilidade de Equipamento"),
		TermoCidadeUF:      getEnv("TERMO_CIDADE_UF", ""),
	}

	expHours, err := strconv.Atoi(getEnv("JWT_EXPIRES_HOURS", "8"))
	if err != nil || expHours <= 0 {
		return nil, fmt.Errorf("JWT_EXPIRES_HOURS inválido: deve ser inteiro positivo")
	}
	cfg.JWTExpiresIn = time.Duration(expHours) * time.Hour

	loginLimite, err := strconv.Atoi(getEnv("LOGIN_RATE_LIMITE", "10"))
	if err != nil || loginLimite <= 0 {
		return nil, fmt.Errorf("LOGIN_RATE_LIMITE inválido: deve ser inteiro positivo")
	}
	cfg.LoginRateLimite = loginLimite

	loginJanelaSeg, err := strconv.Atoi(getEnv("LOGIN_RATE_JANELA_SEG", "60"))
	if err != nil || loginJanelaSeg <= 0 {
		return nil, fmt.Errorf("LOGIN_RATE_JANELA_SEG inválido: deve ser inteiro positivo")
	}
	cfg.LoginRateJanela = time.Duration(loginJanelaSeg) * time.Second

	origins := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")
	cfg.CORSAllowedOrigins = splitAndTrim(origins)

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Producao indica se a aplicação roda em modo produção (GIN_MODE=release).
// Em produção, segredos e senhas não podem cair em valores padrão inseguros.
func (c *Config) Producao() bool {
	return strings.EqualFold(c.GinMode, "release")
}

func (c *Config) validate() error {
	if c.DBDriver != "sqlite" && c.DBDriver != "postgres" {
		return fmt.Errorf("DB_DRIVER inválido %q: use \"sqlite\" ou \"postgres\"", c.DBDriver)
	}
	if c.DBDSN == "" {
		return fmt.Errorf("DB_DSN não pode ser vazio")
	}

	// JWT_SECRET: obrigatório em produção. Em dev, usa um valor padrão inseguro
	// apenas para não travar o boot, com aviso explícito.
	if c.JWTSecret == "" {
		if c.Producao() {
			return fmt.Errorf("JWT_SECRET é obrigatório em produção (GIN_MODE=release): defina um segredo aleatório de pelo menos 16 caracteres no ambiente")
		}
		c.JWTSecret = "sige-ti-dev-secret-INSEGURO-troque-em-producao"
	}
	if len(c.JWTSecret) < 16 {
		return fmt.Errorf("JWT_SECRET muito curto: use ao menos 16 caracteres")
	}

	// ADMIN_SENHA: obrigatória em produção (sem fallback inseguro).
	// Em dev, cai para "admin123" apenas para facilitar testes locais.
	if c.AdminSenha == "" {
		if c.Producao() {
			return fmt.Errorf("ADMIN_SENHA é obrigatória em produção (GIN_MODE=release): defina a senha inicial do administrador no ambiente e troque-a após o primeiro acesso")
		}
		c.AdminSenha = "admin123"
	}
	if len(c.AdminSenha) < 6 {
		return fmt.Errorf("ADMIN_SENHA muito curta: use ao menos 6 caracteres")
	}

	// SEED_DEMO não deve ser usado em produção: dados fictícios não entram no
	// banco de produção. Recusa explicitamente para evitar acidentes.
	if c.SeedDemo && c.Producao() {
		return fmt.Errorf("SEED_DEMO=true não é permitido em produção (GIN_MODE=release): dados fictícios só devem ser inseridos em desenvolvimento")
	}

	return nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
