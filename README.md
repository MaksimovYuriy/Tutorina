# Tutorina

Стартовый монорепозиторий для разработки Tutorina. Каркас основан на проверенных инфраструктурных решениях Artfolio, но не содержит его предметной логики, дизайна и конфигурации деплоя.

## Состав

- `frontend` — React 19, TypeScript, Vite и Material UI;
- `backend` — Go API на chi со структурированными логами и graceful shutdown;
- `postgres` — PostgreSQL 18 и миграции Goose;
- `caddy` — единая точка входа: `/api/*` направляется в backend, остальные запросы — во frontend;
- `compose.yml` — локальная сборка всех образов из исходников.

Снаружи публикуется только Caddy на порту `18080`. Backend и PostgreSQL доступны лишь внутри Docker-сети.

## Быстрый старт

```bash
cp .env.example .env
docker compose up --build
```

После запуска приложение доступно на <http://localhost:18080>. Стартовый экран проверяет всю цепочку `frontend → Caddy → backend → PostgreSQL` запросом к `/api/v1/ping`.

Остановить сервисы:

```bash
docker compose down
```

Данные PostgreSQL сохраняются в именованном Docker volume. Для удаления данных явно выполните `docker compose down --volumes`.

## Разработка без Docker

Для backend нужен локальный PostgreSQL. Параметры подключения можно переопределить переменными из `.env.example`.

```bash
cd backend
go run ./cmd/migrate
go run ./cmd/api
```

Frontend запускается отдельно; Vite проксирует `/api` на `localhost:8081`:

```bash
cd frontend
npm install
npm run dev
```

## Проверки

```bash
cd backend && go test ./...
cd frontend && npm run lint && npm run build
docker compose config --quiet
```

## Перед размещением на сервере

Текущий Compose-файл предназначен для разработки и первичной сборки. Перед публикацией задайте сильный `DB_PASSWORD`, определите домен и TLS-политику в `Caddyfile`, а схему доставки образов и серверного доступа оформите отдельно. В каркасе нет CI/CD, SSH-пользователя `deploy`, ключей или серверных секретов.
