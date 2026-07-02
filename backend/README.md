# SIGE-TI — Backend

API REST do **Sistema de Gestão de Estoque de T.I.** (SIGE-TI), para o
Departamento de T.I. de uma Prefeitura. Controle de estoque e patrimônio de
equipamentos e materiais de consumo.

Stack: **Go + Gin + GORM**, banco **SQLite** por padrão (troca para PostgreSQL
apenas por configuração), autenticação **JWT** com dois perfis
(`administrador` e `operador`).

---

## Arquitetura em camadas

```
cmd/server            -> ponto de entrada (boot, shutdown gracioso)
internal/
  config              -> carregamento de variáveis de ambiente
  database            -> conexão GORM (dialeto abstraído) + AutoMigrate
  models              -> entidades
  repositories        -> acesso a dados (interfaces + GORM)
  services            -> regras de negócio e validações (camada crítica)
  handlers            -> controladores HTTP (Gin)
  middlewares         -> autenticação JWT e checagem de perfil
  router              -> registro de rotas
  container           -> injeção de dependências (repos -> services -> handlers)
  seed                -> dados iniciais de exemplo
```

Fluxo de dependências: **models → repositories → services → handlers**.

---

## Requisitos

- Go 1.26+
- Compilador C (CGO) — necessário para o driver SQLite. No macOS já vem com as
  Command Line Tools; no Linux instale `build-essential`.

---

## Configuração (variáveis de ambiente)

Copie `.env.example` para `.env` e ajuste. Principais variáveis:

| Variável | Padrão | Descrição |
|---|---|---|
| `PORT` | `8080` | Porta HTTP |
| `GIN_MODE` | `debug` | `debug` ou `release` |
| `DB_DRIVER` | `sqlite` | `sqlite` ou `postgres` |
| `DB_DSN` | `sige-ti.db` | Caminho do arquivo (SQLite) ou DSN (PostgreSQL) |
| `JWT_SECRET` | (inseguro em dev) | **Obrigatório em produção** (`GIN_MODE=release`), mín. 16 caracteres. Sem ele, o boot falha em produção |
| `JWT_EXPIRES_HOURS` | `8` | Validade do token |
| `ADMIN_NOME` | `Administrador do Sistema` | Nome do admin criado no seed base |
| `ADMIN_EMAIL` | `admin@sige-ti.local` | E-mail do admin criado no seed base |
| `ADMIN_SENHA` | `admin123` (só em dev) | **Obrigatória em produção**. Troque após o 1º acesso |
| `LOGIN_RATE_LIMITE` | `10` | Tentativas de login por IP por janela |
| `LOGIN_RATE_JANELA_SEG` | `60` | Janela (segundos) do rate limiting de login |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173` | Origens permitidas (vírgula) |
| `SEED_ON_START` | — | `true` roda o seed na inicialização (idempotente) |
| `SEED_DEMO` | `false` | `true` insere dados fictícios (**só dev**; proibido em produção) |
| `PREFEITURA_NOME` | `Prefeitura Municipal` | Cabeçalho do termo (PDF) |
| `PREFEITURA_DEPTO` | `Departamento de T.I.` | Subtítulo do termo |
| `PREFEITURA_LOGO_PATH` | — | Caminho do brasão/logo (PNG/JPG) |
| `TERMO_CABECALHO` | `Termo de Responsabilidade...` | Título do documento |
| `TERMO_CIDADE_UF` | — | Cidade/UF no fecho do termo |

### Trocar para PostgreSQL

Basta configuração — o código abstrai o dialeto:

```bash
DB_DRIVER=postgres
DB_DSN="host=localhost user=sige password=sige dbname=sige_ti port=5432 sslmode=disable TimeZone=America/Sao_Paulo"
```

---

## Executar em desenvolvimento

```bash
cd backend
cp .env.example .env        # ajuste JWT_SECRET, ADMIN_SENHA, etc.
go mod download

# Popular dados (base + demo). Em dev, habilite SEED_DEMO para os fictícios:
SEED_DEMO=true go run ./cmd/server -seed

# Subir a API:
go run ./cmd/server
```

A API sobe em `http://localhost:8080`. Verifique: `GET /health`.

### Seed: base x demo

O seed é dividido em duas trilhas (idempotentes):

- **Seed BASE** (sempre): usuário **administrador** + **categorias** padrão de
  T.I. (patrimoniadas e de consumo) + **setores/secretarias** reais. É o
  necessário para começar a usar o sistema **em produção** (sem dados fictícios).
- **Seed DEMO** (`SEED_DEMO=true`, **só em desenvolvimento**): operador,
  servidores, itens, movimentações (incluindo um alerta de estoque) e um termo,
  todos fictícios. **É bloqueado em produção** (`GIN_MODE=release` recusa o boot
  com `SEED_DEMO=true`).

**Credenciais do admin (seed base):** definidas por `ADMIN_EMAIL`/`ADMIN_SENHA`
(em dev, padrão `admin@sige-ti.local` / `admin123`). Em produção, `ADMIN_SENHA`
é obrigatória e deve ser trocada após o primeiro acesso.
Operador de exemplo (só com demo): `operador@sige-ti.local` / `operador123`.

### Segurança em produção

Com `GIN_MODE=release`, o backend **recusa subir** se faltar `JWT_SECRET`
(mín. 16 caracteres) ou `ADMIN_SENHA`, evitando defaults inseguros. O endpoint
`POST /auth/login` tem **rate limiting por IP** (`LOGIN_RATE_LIMITE` tentativas
a cada `LOGIN_RATE_JANELA_SEG` segundos) para mitigar força bruta. Usuários
**inativos não conseguem autenticar**.

---

## Executar com Docker / Deploy em rede interna

Na raiz do projeto há um `docker-compose.yml` preparado para **deploy em rede
interna (LAN)**: o frontend (Nginx) é publicado na **porta 80** do host e faz
proxy de `/api` para o backend. Segredos vêm do arquivo `.env` da raiz.

```bash
cd ..                      # raiz do projeto
cp .env.example .env       # defina JWT_SECRET e ADMIN_SENHA (obrigatórios)
docker compose up -d --build
```

Acesse de qualquer máquina da rede: `http://IP-do-servidor` (ex.:
`http://192.168.0.50`). Ajuste `CORS_ALLOWED_ORIGINS` no `.env` para esse IP.
Consulte o README da raiz/frontend para detalhes do deploy.

---

## Testes

Cobrem as regras de negócio críticas de estoque (entrada, saída, bloqueio de
saída maior que o estoque, alerta de estoque baixo, baixa patrimonial):

```bash
cd backend
go test ./...
```

---

## Endpoints principais

Base: `/api/v1`. Todas as rotas (exceto `/auth/login` e `/health`) exigem
`Authorization: Bearer <token>`.

- `POST /auth/login` — autenticação (público)
- `GET  /auth/eu` — dados do usuário logado
- `GET/POST/PUT/DELETE /categorias` `/setores` — leitura: ambos; escrita: admin
- `GET/POST/PUT /servidores` — operador+admin; `DELETE`: admin
- `GET/POST/PUT /itens` — cadastro/consulta/busca (filtros via query: `q`,
  `categoria_id`, `setor_id`, `responsavel_id`, `estado`, `baixado`,
  `abaixo_minimo`, `pagina`, `tamanho`)
- `DELETE /itens/:id` — exclui (soft delete) item criado por engano
  (**somente administrador**). Recusado se o item tiver histórico de
  movimentações ou termos (use a baixa patrimonial nesses casos)
- `GET /itens/:id/historico` — histórico de movimentações do item
- `GET /itens/alertas/estoque-baixo` — consumíveis abaixo do mínimo
- `POST/GET /movimentacoes` — registrar/consultar (filtros: `item_id`, `tipo`,
  `data_inicio`, `data_fim`)
- `POST/GET /termos`, `GET /termos/:id`, `GET /termos/:id/pdf` — termo de
  responsabilidade (PDF parametrizado)
- `GET /relatorios/itens-por-setor` `/itens-por-responsavel` `/estoque-baixo`
  `/inventario` `/movimentacoes` — aceitam `?formato=csv|pdf`
- `GET/POST /usuarios`, `GET /usuarios/:id` — gestão de usuários (**admin**)
- `PATCH /usuarios/:id/senha` — redefinir senha de um usuário (**admin**)
- `PATCH /usuarios/:id/ativo` — ativar/desativar usuário (**admin**, sem
  exclusão física; impede desativar o último administrador ativo)

---

## Backup do banco (SQLite)

O banco SQLite é um único arquivo. Para backup **consistente**, use o comando
`.backup` (seguro mesmo com a aplicação rodando, graças ao WAL). **Nunca** copie
o `.db` com `cp` enquanto a app roda — pode corromper a cópia.

### Script pronto (recomendado)

Há um script na raiz que faz backup com timestamp, compactação e **retenção**:

```bash
scripts/backup.sh            # gera backups/sige-ti-AAAAMMDD-HHMMSS.db.gz
```

Por padrão lê do volume Docker `geti_sige_data` (não precisa instalar `sqlite3`
no host — usa um container efêmero), guarda em `./backups` e mantém **14 dias**.
Configurável por env: `SIGE_VOLUME`, `SIGE_DB_PATH`, `BACKUP_DIR`, `RETENCAO_DIAS`.

### Agendamento via cron

Faça backup diário às 2h da manhã (edite com `crontab -e`):

```cron
0 2 * * * cd /caminho/para/geti && /caminho/para/geti/scripts/backup.sh >> /var/log/sige-backup.log 2>&1
```

### Restaurar

Pare a aplicação, restaure e suba novamente:

```bash
docker compose stop backend
scripts/restore.sh backups/sige-ti-AAAAMMDD-HHMMSS.db.gz
docker compose start backend
```

> Para PostgreSQL, use `pg_dump`/`pg_restore`.

---

## LGPD

O sistema coleta o **mínimo necessário** de dados pessoais do servidor
responsável: **apenas nome e matrícula** (entidade `Servidor`). **Não** são
coletados CPF, RG, endereço, telefone ou quaisquer dados sensíveis. O acesso é
protegido por autenticação JWT e perfis de acesso. Senhas são armazenadas como
hash **bcrypt**. O histórico de movimentações é **imutável** (auditoria);
correções são feitas por movimentações compensatórias. A baixa patrimonial
**não apaga** registros — mantém rastreabilidade total.

Usuários **não são apagados fisicamente**: são **desativados** (flag `Ativo`),
preservando a autoria das movimentações já registradas. A **exclusão de itens**
(soft delete) existe apenas para corrigir um cadastro feito por engano e é
**bloqueada** quando o item já tem histórico (movimentações/termos), preservando
a rastreabilidade — nesses casos usa-se a baixa patrimonial.
