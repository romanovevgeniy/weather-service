# Weather Service

Минимальный сервис агрегации погоды на Go. Периодически получает текущую температуру для заданного города из API Open-Meteo и сохраняет показания в PostgreSQL. Предоставляет HTTP-эндпоинт для получения последнего показания по городу.

## Возможности

- Чистая архитектура: domain, ports, adapters, delivery
- Безопасная конфигурация через переменные окружения (без хардкода секретов)
- Плавное завершение работы и таймауты HTTP-сервера
- Внешние HTTP-запросы с контекстом, таймаутами и строгими заголовками
- Подключаемые хранилища (есть адаптер для PostgreSQL)

## Структура проекта

```
cmd/server/                 # Композиция приложения и DI
internal/
  adapter/
    http/                   # Адаптеры к внешним HTTP-сервисам (геокодинг, погода)
    postgres/               # Реализация репозитория на Postgres
  client/http/              # Конкретные HTTP-клиенты для внешних API
  config/                   # Загрузка конфигурации (через env)
  delivery/http/            # HTTP-сервер (роутинг/обработчики)
  domain/                   # Доменные сущности
  ports/                    # Интерфейсы (репозитории, сервисы, часы)
  usecase/                  # Юзкейсы (ингест данных, получение последнего значения)
```

## Требования

- Go (версия соответствует `go.mod`)
- Docker (для PostgreSQL через docker-compose)

## Переменные окружения

- `HTTP_PORT` (по умолчанию: `3000`)
- `DEFAULT_CITY` (по умолчанию: `moscow`)
- `JOB_INTERVAL` (по умолчанию: `10s`) — например, `30s`, `1m`
- `DB_HOST` (по умолчанию: `localhost`)
- `DB_PORT` (по умолчанию: `5432`)
- `DB_USER` (по умолчанию: `admin`)
- `DB_PASSWORD` (по умолчанию нет; обязателен в compose)
- `DB_NAME` (по умолчанию: `weather`)

Строка подключения формируется как: `postgresql://DB_USER:DB_PASSWORD@DB_HOST:DB_PORT/DB_NAME`.

## Быстрый старт

1. Запустите PostgreSQL

```bash
# безопасно задайте пароль в профиле вашей оболочки (пример для локальной разработки)
export DB_PASSWORD="changeme"

docker compose up -d pg
```

2. Создайте таблицу (однократно)

```sql
CREATE TABLE IF NOT EXISTS reading (
  name text NOT NULL,
  timestamp timestamptz NOT NULL,
  temperature double precision NOT NULL
);
CREATE INDEX IF NOT EXISTS reading_name_timestamp_idx
  ON reading (name, timestamp DESC);
```

3. Запустите сервис

```bash
export HTTP_PORT=3000
export DB_HOST=localhost
export DB_PORT=54321
export DB_USER=admin
export DB_PASSWORD="changeme"
export DB_NAME=weather
export DEFAULT_CITY=moscow
export JOB_INTERVAL=10s

go run ./cmd/server
```

4. Запросите последнее показание

```bash
curl -s http://localhost:3000/moscow | jq
```

Пример ответа:

```json
{
	"Name": "moscow",
	"Timestamp": "2025-10-29T09:20:00Z",
	"Temperature": 6.3
}
```

## HTTP API

- `GET /{city}` — возвращает последнее показание по городу
  - 200 OK с JSON в теле
  - 404 Not Found, если данных нет

## Архитектура

- `domain`: сущность `Reading`
- `ports`: интерфейсы `ReadingRepository`, `GeocodingService`, `WeatherService`, `Clock`
- `adapter/postgres`: реализация `ReadingRepository` через pgx
- `adapter/http`: адаптеры, превращающие конкретных клиентов в сервисы `ports`
- `client/http`: конкретные клиенты OpenMeteo (используют context и таймауты)
- `usecase`:
  - `GetLatestReading`
  - `IngestWeather` (периодическая задача)
- `delivery/http`: роутинг и обработчики HTTP
- `cmd/server`: инициализация, DI, планировщик, graceful shutdown

## Безопасность и эксплуатация

- Секреты считываются из переменных окружения или секретов оркестратора (без хардкода).
- HTTP-сервер использует строгие таймауты и корректно завершается по SIGINT/SIGTERM.
- Исходящие HTTP-запросы используют context и явные заголовки `Accept`/`User-Agent`.
- Минимизируй права БД: достаточно INSERT/SELECT на таблицу `reading`.

## TO-DO

- Добавить инструмент миграций (например, golang-migrate или goose) для управления схемой.
- Добавить метрики, трассировку, структурированные логи и health‑чеки.
- При необходимости — ретраи и circuit breaking для внешних API.