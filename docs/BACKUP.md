# Backup e restauração — SIGE-TI

O SIGE-TI guarda os dados em **um único volume Docker** (`sige_data`):
- Banco **SQLite** (`/app/data/sige-ti.db`)
- **Anexos** das mensagens de chamado (`/app/data/anexos/`)

Os scripts fazem backup consistente dos **dois** (banco via `.backup` do SQLite,
que respeita o WAL; anexos via `tar`).

## Fazer backup

```bash
cd /opt/geti
./scripts/backup.sh
```

Gera em `./backups/`:
- `sige-ti-AAAAMMDD-HHMMSS.db.gz` (banco)
- `sige-ti-anexos-AAAAMMDD-HHMMSS.tar.gz` (anexos)

Retenção padrão: 14 dias (configurável).

### Variáveis (opcionais)
| Variável | Padrão | Descrição |
|---|---|---|
| `SIGE_VOLUME` | `geti_sige_data` | nome do volume Docker |
| `SIGE_DB_PATH` | `/app/data/sige-ti.db` | caminho do `.db` no volume |
| `ANEXOS_SUBDIR` | `anexos` | subpasta dos anexos no volume |
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
./scripts/restore.sh backups/sige-ti-AAAAMMDD-HHMMSS.db.gz
# (o tar de anexos do mesmo horário é restaurado automaticamente se estiver ao lado)
docker compose start backend
```

Restaurar anexos de um arquivo específico:
```bash
./scripts/restore.sh backups/sige-ti-<ts>.db.gz backups/sige-ti-anexos-<ts>.tar.gz
```

## Teste periódico
Faça uma **restauração de teste** de vez em quando (num ambiente separado) —
backup que nunca foi testado não é backup.
