#!/usr/bin/env bash
# ==========================================================================
# SIGE-TI - Restauração do banco SQLite a partir de um backup
#
# ATENÇÃO: substitui o banco atual pelo conteúdo do backup. Pare a aplicação
# antes de restaurar para evitar inconsistências.
#
# Uso:
#   scripts/restore.sh <caminho-do-backup.db.gz | .db>
#
# Variáveis (opcional): SIGE_VOLUME, SIGE_DB_PATH (mesmos defaults do backup.sh)
# ==========================================================================
set -euo pipefail

if [ $# -lt 1 ]; then
  echo "Uso: $0 <arquivo-de-backup (.db ou .db.gz)>" >&2
  exit 1
fi

BACKUP_FILE="$1"
SIGE_VOLUME="${SIGE_VOLUME:-geti_sige_data}"
SIGE_DB_PATH="${SIGE_DB_PATH:-/app/data/sige-ti.db}"
DB_BASENAME="$(basename "$SIGE_DB_PATH")"

if [ ! -f "$BACKUP_FILE" ]; then
  echo "[restore] ERRO: arquivo não encontrado: $BACKUP_FILE" >&2
  exit 1
fi

# Descompacta para um temporário se necessário.
TMP_DB="$(mktemp /tmp/sige-restore-XXXXXX.db)"
trap 'rm -f "$TMP_DB"' EXIT
case "$BACKUP_FILE" in
  *.gz) gunzip -c "$BACKUP_FILE" > "$TMP_DB" ;;
  *)    cp "$BACKUP_FILE" "$TMP_DB" ;;
esac

echo "[restore] PARE a aplicação antes de continuar (docker compose stop backend)."
read -r -p "Continuar a restauração? (s/N) " RESP
[ "$RESP" = "s" ] || [ "$RESP" = "S" ] || { echo "Cancelado."; exit 0; }

if command -v docker >/dev/null 2>&1; then
  docker run --rm \
    -v "${SIGE_VOLUME}:/data" \
    -v "${TMP_DB}:/restore.db:ro" \
    alpine:3.20 sh -c "cp /restore.db '/data/${DB_BASENAME}' && rm -f '/data/${DB_BASENAME}-wal' '/data/${DB_BASENAME}-shm'"
else
  cp "$TMP_DB" "$SIGE_DB_PATH"
  rm -f "${SIGE_DB_PATH}-wal" "${SIGE_DB_PATH}-shm"
fi

echo "[restore] concluído. Suba a aplicação novamente (docker compose start backend)."
