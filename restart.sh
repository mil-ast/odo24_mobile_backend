#!/usr/bin/env bash
#
# Перезапуск сервиса odo24_mobile_backend.
#
# Скрипт убивает запущенный экземпляр (pkill по точной командной строке)
# и запускает новый из уже скомпилированного файла ./odo24_mobile_backend.
#
# Использование:
#   ./restart.sh
#
# Пути можно переопределить переменными окружения:
#   APP_LOG_FILE  — файл лога  (по умолчанию ./odo24_mobile_backend.log)
#   APP_PID_FILE  — pid-файл   (по умолчанию ./odo24_mobile_backend.pid)
#
set -uo pipefail

APP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_NAME="odo24_mobile_backend"
APP_BIN="$APP_DIR/$APP_NAME"
LOG_FILE="${APP_LOG_FILE:-$APP_DIR/$APP_NAME.log}"
PID_FILE="${APP_PID_FILE:-$APP_DIR/$APP_NAME.pid}"

log() { printf '[restart] %s\n' "$*"; }

# config.json читается относительно текущего каталога, поэтому работаем из каталога скрипта.
cd "$APP_DIR"

if [[ ! -f "$APP_BIN" ]]; then
    log "Ошибка: не найден файл $APP_BIN"
    exit 1
fi
chmod +x "$APP_BIN" 2>/dev/null || true

# Точное совпадение всей командной строки: не заденет tail/grep/редакторы,
# которые просто упоминают имя файла в аргументах.
PROC_PATTERN="^(.*/)?$APP_NAME$"
running_pids() { pgrep -x -f -- "$PROC_PATTERN" || true; }

# --- Остановка текущего процесса ---
pids="$(running_pids | tr '\n' ' ')"
if [[ -n "${pids// /}" ]]; then
    log "Останавливаю: $pids"
    pkill -TERM -x -f -- "$PROC_PATTERN"
    for _ in {1..10}; do
        [[ -z "$(running_pids)" ]] && break
        sleep 1
    done
    if [[ -n "$(running_pids)" ]]; then
        log "Не завершился штатно, применяю SIGKILL"
        pkill -KILL -x -f -- "$PROC_PATTERN"
        sleep 1
    fi
else
    log "Запущенный процесс не найден — останавливать нечего."
fi
rm -f "$PID_FILE"

# --- Запуск нового процесса ---
# setsid отвязывает процесс от сессии терминала, nohup игнорирует SIGHUP.
log "Запускаю $APP_NAME..."
if command -v setsid >/dev/null 2>&1; then
    setsid nohup "$APP_BIN" </dev/null >>"$LOG_FILE" 2>&1 &
else
    nohup "$APP_BIN" </dev/null >>"$LOG_FILE" 2>&1 &
fi

sleep 2
new_pid="$(running_pids | head -n 1)"
if [[ -n "$new_pid" ]]; then
    printf '%s\n' "$new_pid" >"$PID_FILE"
    log "Сервис запущен (PID $new_pid). Лог: $LOG_FILE"
else
    log "Ошибка: процесс завершился сразу после запуска. Последние строки лога:"
    tail -n 50 "$LOG_FILE" 2>/dev/null || true
    exit 1
fi
