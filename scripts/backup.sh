#!/usr/bin/env bash
# ==========================================================================
# SIGE-TI - Backup consistente do banco SQLite
#
# Usa o comando ".backup" do sqlite3, que produz uma cópia consistente mesmo
# com o banco em uso (respeita o WAL). NÃO copie o arquivo .db diretamente com
# cp enquanto a aplicação roda — isso pode gerar uma cópia corrompida.
#
# Gera um arquivo com timestamp e aplica retenção (apaga backups antigos).
#
# Uso:
#   scripts/backup.sh
#
# Configuração por variáveis de ambiente (opcional):
#   SIGE_VOLUME      Nome do volume Docker do banco (padrão: geti_sige_data)
#   SIGE_DB_PATH     Caminho do .db DENTRO do volume/container (padrão: /app/data/sige-ti.db)
#   BACKUP_DIR       Pasta de destino dos backups (padrão: ./backups)
#   RETENCAO_DIAS    Dias a manter (padrão: 14)
#
# Pré-requisito: Docker (o script usa um container efêmero com sqlite3, então
# não é preciso instalar sqlite3 no host). Para instalação direta (sem Docker),
# defere para o modo local — veja a seção "MODO LOCAL" abaixo.
# ==========================================================================
set -euo pipefail

# Resolve a raiz do projeto (pasta-pai deste script).
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJ_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

SIGE_VOLUME="${SIGE_VOLUME:-geti_sige_data}"
SIGE_DB_PATH="${SIGE_DB_PATH:-/app/data/sige-ti.db}"
BACKUP_DIR="${BACKUP_DIR:-$PROJ_DIR/backups}"
RETENCAO_DIAS="${RETENCAO_DIAS:-14}"

TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
ARQUIVO="$BACKUP_DIR/sige-ti-${TIMESTAMP}.db"

mkdir -p "$BACKUP_DIR"

echo "[backup] iniciando backup do SQLite (${TIMESTAMP})..."

# ---- MODO DOCKER (padrão) ----
# Sobe um container efêmero (alpine + sqlite) com o volume do banco montado e
# o diretório de backup do host, e executa o ".backup" para um arquivo.
if command -v docker >/dev/null 2>&1; then
  # Dentro do container, o volume é montado em /data; o .db é o basename do
  # caminho configurado (ex.: sige-ti.db).
  DB_BASENAME="$(basename "$SIGE_DB_PATH")"
  docker run --rm \
    -v "${SIGE_VOLUME}:/data:ro" \
    -v "${BACKUP_DIR}:/backup" \
    alpine:3.20 sh -c "
      set -e
      apk add --no-cache sqlite >/dev/null 2>&1
      sqlite3 '/data/${DB_BASENAME}' \".backup '/backup/$(basename "$ARQUIVO")'\"
    "
else
  # ---- MODO LOCAL (sem Docker) ----
  # Requer sqlite3 instalado no host e SIGE_DB_PATH apontando para o arquivo
  # real no disco (ex.: /var/lib/sige-ti/sige-ti.db).
  if ! command -v sqlite3 >/dev/null 2>&1; then
    echo "[backup] ERRO: nem docker nem sqlite3 disponíveis no host." >&2
    exit 1
  fi
  sqlite3 "$SIGE_DB_PATH" ".backup '$ARQUIVO'"
fi

# Compacta o backup para economizar espaço.
gzip -f "$ARQUIVO"
ARQUIVO="${ARQUIVO}.gz"
echo "[backup] gerado: $ARQUIVO"

# ---- Retenção: remove backups com mais de RETENCAO_DIAS dias ----
echo "[backup] aplicando retenção de ${RETENCAO_DIAS} dias..."
find "$BACKUP_DIR" -name 'sige-ti-*.db.gz' -type f -mtime +"$RETENCAO_DIAS" -print -delete || true

echo "[backup] concluído."
