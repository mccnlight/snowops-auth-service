# Snowops Auth Service

Сервис отвечает за аутентификацию пользователей SnowOps, выпуск access/refresh JWT, управление сессиями и выдачу одноразовых SMS-кодов.

## Требования

- Go 1.23+
- PostgreSQL 15+

## Запуск локально

```bash
cd deploy
docker compose up -d   # PostgreSQL + миграции

cd ..
APP_ENV=development \
DB_DSN="postgres://postgres:postgres@localhost:5431/auth_db?sslmode=disable" \
JWT_ACCESS_SECRET="secret-key" \
go run ./cmd/auth-service
```

## Переменные окружения

| Переменная                | Описание                                                         | Значение по умолчанию                                      |
|---------------------------|------------------------------------------------------------------|-------------------------------------------------------------|
| `APP_ENV`                 | окружение (`development`, `production`)                          | `development`                                              |
| `HTTP_HOST` / `HTTP_PORT` | адрес HTTP-сервера                                               | `0.0.0.0` / `7080`                                         |
| `DB_DSN`                  | строка подключения к PostgreSQL                                  | `postgres://postgres:postgres@localhost:5431/auth_db?sslmode=disable` |
| `DB_MAX_OPEN_CONNS`       | максимум открытых соединений                                     | `25`                                                        |
| `DB_MAX_IDLE_CONNS`       | максимум соединений в пуле                                       | `10`                                                        |
| `DB_CONN_MAX_LIFETIME`    | TTL соединения                                                   | `1h`                                                        |
| `JWT_ACCESS_SECRET`       | секрет для подписи access JWT                                    | `supersecret`                                              |
| `JWT_ACCESS_TTL`          | срок жизни access токена (Go duration)                           | `15m`                                                       |
| `JWT_REFRESH_TTL`         | срок жизни refresh токена                                        | `720h`                                                      |
| `SMS_CODE_TTL`            | срок действия SMS-кода                                           | `5m`                                                        |
| `SMS_CODE_LENGTH`         | длина генерируемого кода                                         | `6`                                                         |
| `SMS_DAILY_LIMIT`         | дневной лимит SMS на номер                                       | `10`                                                        |
| `ADMIN_SEED_ENABLED`      | создавать ли пользователя/организацию при старте                 | `true`                                                      |
| `ADMIN_LOGIN` / `ADMIN_PASSWORD` / `ADMIN_PHONE` | данные сид-админа                                       | `admin` / `admin123` / ``                                   |
| `ADMIN_ORG_NAME` / `ADMIN_ORG_BIN` | данные базовой организации                             | `Akimat` / ``                                              |
| `REG_DEFAULT_ROLE`        | роль по умолчанию для автогенерации пользователей (если нужно)   | пусто                                                       |
| `REG_DEFAULT_ORG_NAME` / `REG_DEFAULT_ORG_BIN` | значения для автосоздания организации                 | пусто                                                       |

## API

Базовый URL: `https://auth.<env>/`. Все ответы (кроме `send-code`/`logout`) оборачиваются в `{"data": ...}`. Эндпоинты `/auth/me` и любые защищённые маршруты требуют `Authorization: Bearer <access_token>`.

### Health

#### `GET /healthz`
```json
{ "status": "ok" }
```

---

### Аутентификация (`/auth`)

#### `POST /auth/login`
- Авторизация по логину/паролю.

```json
{ "login": "dispatcher1", "password": "Secret123" }
```

```json
{
  "data": {
    "access_token": "eyJhbGciOi...",
    "refresh_token": "7d5cba1e-...",
    "user": {
      "id": "f2bd7a74-5a23-4a2a-8ae7-238305cb12ad",
      "organization_id": "c1bd3f4f-9e88-4f1f-80a8-46b1cba6a2d0",
      "organization": "TOO Тест",
      "role": "CONTRACTOR_ADMIN",
      "phone": "+77001234567",
      "login": "dispatcher1"
    }
  }
}
```

#### `POST /auth/send-code`
- Отправка одноразового SMS-кода для входа по номеру телефона.

```json
{ "phone": "+77001234567" }
```

```json
{ "masked_phone": "+7700***4567" }
```

#### `POST /auth/verify-code`
- Ввод SMS-кода, создаёт сессию и выдаёт токены.

```json
{ "phone": "+77001234567", "code": "123456" }
```

```json
{
  "data": {
    "access_token": "eyJhbGciOi...",
    "refresh_token": "55d0d5e4-...",
    "user": {
      "id": "2df028a3-...",
      "organization_id": "a11a83f7-...",
      "organization": "KGU ZHKH",
      "role": "KGU_ZKH_ADMIN",
      "phone": "+77001234567",
      "is_new": false
    }
  }
}
```

#### `POST /auth/refresh`
- Обновление пары токенов. Требует действующий refresh.

```json
{ "refresh_token": "55d0d5e4-..." }
```

Ответ аналогичен `login`.

#### `POST /auth/logout`
- Инвалидирует refresh-токен.

```json
{ "refresh_token": "55d0d5e4-..." }
```

```json
{ "success": true }
```

#### `GET /auth/me`
- Информация о текущем пользователе (по access JWT). Требуется заголовок `Authorization`.

```json
{
  "data": {
    "id": "2df028a3-...",
    "organization_id": "a11a83f7-...",
    "organization": "KGU ZHKH",
    "role": "KGU_ZKH_ADMIN",
    "phone": "+77001234567",
    "login": null
  }
}
```

## Формат ошибок

При ошибках возвращается

```json
{ "error": "описание проблемы" }
```

Соответствие кодов:
- 400 — некорректные данные (`ErrInvalidInput` и т.п.).
- 401 — неверные креды/код/комбинация (`ErrInvalidCredentials`, `ErrCodeInvalid`, `ErrCodeExpired`, отсутствующий токен).
- 403 — недостаточно прав (`ErrPermissionDenied`, `ErrHierarchyViolation`).
- 404 — пользователь/сессия не найдены.
- 409 — конфликты (например, лимит SMS).
- 500 — внутренняя ошибка (лог пишется в stderr / Zerolog).

