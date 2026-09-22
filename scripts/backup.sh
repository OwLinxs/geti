#!/usr/bin/env bash
# ==========================================================================
# SIGE-TI - Backup (PostgreSQL + anexos)
#
# Faz dump lógico do PostgreSQL (pg_dump) e arquiva os anexos das mensagens
# (que ficam no volume, fora do banco). Gera arquivos com timestamp e aplica
# retenção.
#
# Uso:
#   scripts/backup.sh
#
# Variáveis (opcionais):
#   PG_CONTAINER   Container do Postgres (padrão: sige-ti-postgres)
#   SIGE_VOLUME    Volume Docker dos anexos (padrão: geti_sige_data)
#   ANEXOS_SUBDIR  Subpasta dos anexos no volume (padrão: anexos)
#   BACKUP_DIR     Destino (padrão: ./backups)
#   RETENCAO_DIAS  Dias a manter (padrão: 14)
#
# As credenciais (POSTGRES_USER/DB) são lidas do .env do projeto.
# ==========================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJ_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Lê uma variável do .env (com fallback).
env_get() {
  local chave="$1" padrao="$2"
  local val
  val="$(grep -E "^${chave}=" "$PROJ_DIR/.env" 2>/dev/null | head -1 | cut -d= -f2- || true)"
  echo "${val:-$padrao}"
}

PG_CONTAINER="${PG_CONTAINER:-sige-ti-postgres}"
PG_USER="$(env_get POSTGRES_USER sige)"
PG_DB="$(env_get POSTGRES_DB sige_ti)"
SIGE_VOLUME="${SIGE_VOLUME:-geti_sige_data}"
ANEXOS_SUBDIR="${ANEXOS_SUBDIR:-anexos}"
BACKUP_DIR="${BACKUP_DIR:-$PROJ_DIR/backups}"
RETENCAO_DIAS="${RETENCAO_DIAS:-14}"

TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
SQL_GZ="$BACKUP_DIR/sige-ti-${TIMESTAMP}.sql.gz"
ANEXOS_TAR="$BACKUP_DIR/sige-ti-anexos-${TIMESTAMP}.tar.gz"

mkdir -p "$BACKUP_DIR"
echo "[backup] iniciando (${TIMESTAMP})..."

# ---- Banco: pg_dump do container do Postgres ----
if ! docker ps --format '{{.Names}}' | grep -q "^${PG_CONTAINER}$"; then
  echo "[backup] ERRO: container ${PG_CONTAINER} não está rodando." >&2
  exit 1
fi
docker exec -e PGUSER="$PG_USER" "$PG_CONTAINER" \
  pg_dump -d "$PG_DB" --clean --if-exists | gzip > "$SQL_GZ"
echo "[backup] banco: $SQL_GZ"

# ---- Anexos: tar do volume ----
if command -v docker >/dev/null 2>&1; then
  docker run --rm \
    -v "${SIGE_VOLUME}:/data:ro" \
    -v "${BACKUP_DIR}:/backup" \
    alpine:3.20 sh -c "
      if [ -d '/data/${ANEXOS_SUBDIR}' ]; then
        tar czf '/backup/$(basename "$ANEXOS_TAR")' -C /data '${ANEXOS_SUBDIR}'
      else
        echo '[backup] sem diretório de anexos.'
      fi
    "
fi
[ -f "$ANEXOS_TAR" ] && echo "[backup] anexos: $ANEXOS_TAR"

# ---- Retenção ----
echo "[backup] retenção de ${RETENCAO_DIAS} dias..."
find "$BACKUP_DIR" -name 'sige-ti-*.sql.gz' -type f -mtime +"$RETENCAO_DIAS" -delete || true
find "$BACKUP_DIR" -name 'sige-ti-anexos-*.tar.gz' -type f -mtime +"$RETENCAO_DIAS" -delete || true

echo "[backup] concluído."
