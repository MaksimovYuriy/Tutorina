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

После запуска приложение доступно на <http://localhost:18080>. На стартовой странице отображаются опубликованные преподаватели; служебная проверка API доступна по `/api/v1/ping`.

Создать первую локальную учётную запись с ролями преподавателя и администратора:

```bash
BOOTSTRAP_EMAIL=teacher@example.com \
BOOTSTRAP_PASSWORD='replace-with-a-long-password' \
BOOTSTRAP_ROLES=teacher,admin \
docker compose --profile tools run --rm create-user
```

Пароль должен содержать не менее 12 символов. Команда хранит только bcrypt-хеш; повторный email будет отклонён.

Маршруты авторизации:

- `POST /api/v1/auth/sessions` — вход по `email` и `password`;
- `GET /api/v1/auth/me` — текущий пользователь и его роли;
- `PUT /api/v1/auth/password` — смена собственного пароля с завершением всех сессий;
- `DELETE /api/v1/auth/session` — завершение сессии.

Сессия передаётся в `HttpOnly`, `SameSite=Strict` cookie. Для локального HTTP `AUTH_COOKIE_SECURE=false`; при будущем размещении за HTTPS параметр необходимо включить.

Маршруты преподавателей:

- `GET /api/v1/teachers` — публичный список опубликованных профилей;
- `GET|PUT /api/v1/teacher/profile` — просмотр и редактирование собственного профиля;
- `PUT|DELETE /api/v1/teacher/profile/photo` — управление собственной фотографией;
- `GET|POST /api/v1/admin/teachers/` — список и создание профилей для администратора;
- `PUT|DELETE /api/v1/admin/teachers/{id}/photo` — загрузка/замена и удаление фотографии;
- `PUT /api/v1/admin/teachers/{id}` — редактирование профиля;
- `POST /api/v1/admin/teachers/{id}/account` — создание учётной записи;
- `PUT /api/v1/admin/teachers/{id}/password` — сброс временного пароля с завершением сессий;
- `DELETE /api/v1/admin/teachers/{id}` — архивация профиля и отключение доступа.

Фотографии принимаются в JPEG, PNG или WebP размером до 5 МБ. В Docker они хранятся
в именованном томе `teacher_photos` и не теряются при пересборке контейнера. При
замене и архивации старый файл удаляется.

Остановить сервисы:

```bash
docker compose down
```

Данные PostgreSQL и фотографии сохраняются в именованных Docker volumes. Команда `docker compose down --volumes` явно удалит оба хранилища.

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
