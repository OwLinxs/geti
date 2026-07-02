# SIGE-TI — Sistema de Gestão de Estoque de T.I.

Controle de estoque e patrimônio de equipamentos e materiais de consumo para o
Departamento de T.I. de uma Prefeitura. Backend em **Go (Gin + GORM)**, frontend
em **React (Vite + TypeScript + Tailwind)**, banco **SQLite** (troca para
PostgreSQL por configuração). Interface 100% em **português do Brasil**.

- Backend: [`backend/README.md`](backend/README.md)
- Frontend: [`frontend/README.md`](frontend/README.md)

---

## Deploy em servidor de rede interna (LAN)

Pré-requisitos no servidor: **Docker** e **Docker Compose**. O acesso é apenas
interno (sem exposição à internet).

### 1. Configurar segredos

```bash
cp .env.example .env
nano .env
```

Defina **obrigatoriamente** (sem eles o backend não sobe em produção):

- `JWT_SECRET` — segredo do token. Gere com: `openssl rand -base64 48`
- `ADMIN_SENHA` — senha inicial do administrador (troque após o 1º acesso)

Ajuste também `CORS_ALLOWED_ORIGINS` para o IP/hostname do servidor
(ex.: `http://192.168.0.50`) e os campos `PREFEITURA_*`/`TERMO_*` do termo.

> O `.env` contém segredos e **não deve ser versionado** (já está no
> `.gitignore`). Use `.env.example` como referência.

### 2. Subir os serviços

```bash
docker compose up -d --build
```

Isso sobe:

- **backend** (API Go, SQLite em volume `sige_data`) — não exposto diretamente
- **frontend** (Nginx na **porta 80**) — serve a SPA e faz proxy de `/api` para
  o backend

Na **primeira execução**, o seed **base** cria o administrador, as categorias
padrão de T.I. e os setores/secretarias. **Nenhum dado fictício** é inserido em
produção.

### 3. Acessar

De qualquer máquina da rede interna, abra:

```
http://IP-do-servidor          (ex.: http://192.168.0.50)
```

Entre com `ADMIN_EMAIL` / `ADMIN_SENHA` e **troque a senha** do admin na tela
de Usuários. Crie as contas dos operadores (funcionários que cadastram itens).

### 4. Comandos úteis

```bash
docker compose ps                 # status
docker compose logs -f backend    # logs do backend
docker compose down               # parar (mantém o volume/dados)
docker compose up -d --build      # atualizar após mudanças de código
```

---

## Backup e restauração do banco

Os dados ficam no volume Docker `geti_sige_data`. Use os scripts prontos
(backup consistente via `.backup` do sqlite3, com timestamp e retenção):

```bash
scripts/backup.sh                 # gera backups/sige-ti-<timestamp>.db.gz (retém 14 dias)
```

Agende um backup diário às 2h (em `crontab -e`):

```cron
0 2 * * * cd /caminho/para/geti && /caminho/para/geti/scripts/backup.sh >> /var/log/sige-backup.log 2>&1
```

Restaurar a partir de um backup:

```bash
docker compose stop backend
scripts/restore.sh backups/sige-ti-<timestamp>.db.gz
docker compose start backend
```

Detalhes e variáveis de configuração dos scripts no
[`backend/README.md`](backend/README.md#backup-do-banco-sqlite).

---

## Segurança (resumo)

- Segredos **fora do código** — lidos do `.env` (não versionado). Em produção
  (`GIN_MODE=release`), o boot **falha** sem `JWT_SECRET` e `ADMIN_SENHA`.
- **Rate limiting** no login (`POST /auth/login`) por IP, contra força bruta.
- Senhas em **bcrypt**. Contas podem ser **desativadas** (sem exclusão física);
  usuário inativo não autentica.
- **LGPD**: do servidor responsável guardamos apenas **nome e matrícula**.
  Nenhum dado sensível é coletado. Acesso protegido por JWT e perfis.

---

## Perfis de acesso

- **Administrador**: acesso total — gestão de usuários (criar, redefinir senha,
  ativar/desativar), categorias, setores, exclusão de itens cadastrados por
  engano, além de tudo que o operador faz.
- **Operador**: dia a dia — cadastra/edita itens, registra movimentações,
  cadastra servidores, emite termos e consulta relatórios.
