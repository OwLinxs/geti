#!/usr/bin/env bash
# ==========================================================================
# SIGE-TI - Restauração (PostgreSQL + anexos)
#
# ATENÇÃO: substitui o banco atual (o dump usa --clean). Pare o backend antes
# para evitar escritas concorrentes.
#
# Uso:
#   scripts/restore.sh <backup.sql.gz> [anexos.tar.gz]
#
# Se o tar de anexos não for informado, procura o de mesmo timestamp ao lado.
# ==========================================================================
set -euo pipefail

if [ $# -lt 1 ]; then
  echo "Uso: $0 <backup.sql.gz> [anexos.tar.gz]" >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJ_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
env_get() {
  local val; val="$(grep -E "^$1=" "$PROJ_DIR/.env" 2>/dev/null | head -1 | cut -d= -f2- || true)"
  echo "${val:-$2}"
}

SQL_FILE="$1"
ANEXOS_TAR="${2:-}"
PG_CONTAINER="${PG_CONTAINER:-sige-ti-postgres}"
PG_USER="$(env_get POSTGRES_USER sige)"
PG_DB="$(env_get POSTGRES_DB sige_ti)"
SIGE_VOLUME="${SIGE_VOLUME:-geti_sige_data}"
ANEXOS_SUBDIR="${ANEXOS_SUBDIR:-anexos}"

[ -f "$SQL_FILE" ] || { echo "[restore] arquivo não encontrado: $SQL_FILE" >&2; exit 1; }

echo "[restore] PARE o backend antes (docker compose stop backend)."
read -r -p "Continuar a restauração no banco '${PG_DB}'? (s/N) " RESP
[ "$RESP" = "s" ] || [ "$RESP" = "S" ] || { echo "Cancelado."; exit 0; }

# ---- Banco ----
gunzip -c "$SQL_FILE" | docker exec -i -e PGUSER="$PG_USER" "$PG_CONTAINER" psql -d "$PG_DB"
echo "[restore] banco restaurado."

# ---- Anexos ----
if [ -z "$ANEXOS_TAR" ]; then
  TS="$(basename "$SQL_FILE" | sed -n 's/^sige-ti-\([0-9]\{8\}-[0-9]\{6\}\)\.sql.*/\1/p')"
  CAND="$(dirname "$SQL_FILE")/sige-ti-anexos-${TS}.tar.gz"
  [ -n "$TS" ] && [ -f "$CAND" ] && ANEXOS_TAR="$CAND"
fi
if [ -n "$ANEXOS_TAR" ] && [ -f "$ANEXOS_TAR" ]; then
  echo "[restore] restaurando anexos de: $ANEXOS_TAR"
  docker run --rm \
    -v "${SIGE_VOLUME}:/data" \
    -v "${ANEXOS_TAR}:/anexos.tar.gz:ro" \
    alpine:3.20 sh -c "rm -rf '/data/${ANEXOS_SUBDIR}' && tar xzf /anexos.tar.gz -C /data"
  echo "[restore] anexos restaurados."
else
  echo "[restore] (sem tar de anexos — só o banco foi restaurado)"
fi

echo "[restore] concluído. Suba o backend (docker compose start backend)."
