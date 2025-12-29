#!/bin/bash

if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

nohup go run main.go > app.log 2>&1 &
PID=$!

echo "Сервер запущен (PID: $PID)."
sleep 5

if kill -0 $PID 2>/dev/null; then
    echo "Приложение работает."
    echo "=== Tailing logs (first 20 lines) ==="
    head -n 20 app.log
else
    echo "Ошибка запуска приложения!"
    cat app.log
    exit 1
fi