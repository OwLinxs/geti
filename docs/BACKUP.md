# Backup e restauração — SIGE-TI

O SIGE-TI guarda os dados em dois lugares:
- Banco **PostgreSQL** (serviço `postgres`, volume `sige_pgdata`)
- **Anexos** das mensagens de chamado (volume `sige_data`, em `/app/data/anexos/`)

Os scripts fazem backup dos **dois**: banco via `pg_dump` (dump lógico `.sql.gz`)
e anexos via `tar`.

## Fazer backup

```bash
cd /opt/geti
./scripts/backup.sh
```

Gera em `./backups/`:
- `sige-ti-AAAAMMDD-HHMMSS.sql.gz` (banco PostgreSQL)
- `sige-ti-anexos-AAAAMMDD-HHMMSS.tar.gz` (anexos)

Retenção padrão: 14 dias (configurável). Requer o container `sige-ti-postgres`
rodando (o script usa `pg_dump`). Credenciais lidas do `.env`.

### Variáveis (opcionais)
| Variável | Padrão | Descrição |
|---|---|---|
| `PG_CONTAINER` | `sige-ti-postgres` | container do Postgres |
| `SIGE_VOLUME` | `geti_sige_data` | volume dos anexos |
| `ANEXOS_SUBDIR` | `anexos` | subpasta dos anexos |
| `BACKUP_DIR` | `./backups` | destino |
| `RETENCAO_DIAS` | `14` | dias a manter |

> O nome do volume costuma ser `<pasta-do-projeto>_sige_data`. Confirme com
> `docker volume ls`. Se o projeto está em `/opt/geti`, é `geti_sige_data`.

## Agendar (cron) — todo dia às 2h

```bash
crontab -e
# adicione:
0 2 * * * cd /opt/geti && ./scripts/backup.sh >> /var/log/sige-backup.log 2>&1
```

**Recomendado:** copie os backups para **fora da VM** (outro servidor, storage,
nuvem). Backup no mesmo disco não protege contra falha de disco.

```bash
# exemplo: enviar para outro host por rsync após o backup
rsync -az /opt/geti/backups/ usuario@servidor-backup:/backups/sige-ti/
```

## Restaurar

⚠️ Substitui o banco/anexos atuais. **Pare o backend antes.**

```bash
cd /opt/geti
docker compose stop backend
./scripts/restore.sh backups/sige-ti-AAAAMMDD-HHMMSS.sql.gz
# (o tar de anexos do mesmo horário é restaurado automaticamente se estiver ao lado)
docker compose start backend
```

Restaurar anexos de um arquivo específico:
```bash
./scripts/restore.sh backups/sige-ti-<ts>.sql.gz backups/sige-ti-anexos-<ts>.tar.gz
```

## Teste periódico
Faça uma **restauração de teste** de vez em quando (num ambiente separado) —
backup que nunca foi testado não é backup.
