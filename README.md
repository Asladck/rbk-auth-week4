    # Weather + Users API v2 — Auth & Security

Production-ready REST API на Go с JWT-аутентификацией, ролевой авторизацией, управлением пользователями, городами и получением актуальной погоды.

---

## Стек

| Слой | Технология |
|---|---|
| Язык | Go 1.22 |
| HTTP-роутер | [chi v5](https://github.com/go-chi/chi) |
| База данных | PostgreSQL 16 (pgx v5) |
| Аутентификация | JWT (golang-jwt/jwt v5) |
| Хэширование паролей | bcrypt (cost 12) |
| Логирование | [zap](https://github.com/uber-go/zap) |
| Погода | [weatherapi.com](https://www.weatherapi.com/) |
| Геолокация | [countrystatecity.in](https://countrystatecity.in/) |
| Контейнеризация | Docker + Docker Compose |
| Кэширование | Redis (опционально) |

---

## Архитектура

```
Request → AuthMiddleware → Handler → Service → Repository
```

Проект следует **Hexagonal Architecture (Ports & Adapters)**. Слои зависят только от интерфейсов — ни один слой не знает о реализации другого.

```
cmd/api/              # точка входа, wire-up всех зависимостей
internal/
  domain/             # модели, интерфейсы (порты), context helpers
  config/             # загрузка конфигурации из env
  repository/postgres # DB-адаптеры (реализуют domain-интерфейсы)
  service/            # бизнес-логика
  client/             # HTTP-клиенты к внешним API
  handler/            # HTTP-хэндлеры + роутер
  middleware/         # Auth (JWT), RequireRole, Logger, Recoverer
migrations/           # SQL-миграции
pkg/
  apperrors/          # sentinel errors + ValidationError
  jwtutil/            # JWT sign + parse (изолирован)
  logger/             # инициализация zap
```

---

## Быстрый старт

### Docker (рекомендуется)

```bash
# 1. Клонировать репозиторий
git clone https://github.com/yourorg/weather-auth-api.git
cd weather-auth-api

# 2. Создать .env
cp .env.example .env

# 3. Сгенерировать JWT_SECRET
make gen-secret   # → вставить результат в .env как JWT_SECRET

# 4. Вписать API-ключи в .env
#    WEATHER_API_KEY      — https://www.weatherapi.com/  (бесплатный план)
#    COUNTRY_STATE_CITY_KEY — https://countrystatecity.in/ (бесплатный план)

# 5. Поднять PostgreSQL + API
make docker-up

# 6. Проверить
curl http://localhost:8080/health
```

### Локально (без Docker)

```bash
# Требуется: Go 1.22+, PostgreSQL 16

cp .env.example .env   # настроить DATABASE_URL, JWT_SECRET, API ключи

make migrate-up        # применить миграции
make run               # запустить сервис
```

---

## Переменные окружения

| Переменная | Обязательная | Описание | По умолчанию |
|---|---|---|---|
| `DATABASE_URL` | ✅ | DSN подключения к PostgreSQL | — |
| `JWT_SECRET` | ✅ | Секрет для подписи токенов (`make gen-secret`) | — |
| `WEATHER_API_KEY` | ✅ | Ключ от weatherapi.com | — |
| `COUNTRY_STATE_CITY_KEY` | ✅ | Ключ от countrystatecity.in | — |
| `APP_ENV` | | `development` или `production` | `production` |
| `HTTP_PORT` | | Порт сервера | `8080` |
| `JWT_ACCESS_TTL_MINUTES` | | Время жизни токена (мин) | `60` |
| `DB_MAX_OPEN_CONNS` | | Макс. открытых соединений | `25` |
| `DB_MAX_IDLE_CONNS` | | Макс. idle-соединений | `10` |
| `DB_CONN_MAX_LIFETIME` | | Время жизни соединения (сек) | `300` |
| `REDIS_ADDR` | | Адрес Redis (`host:port`). Пусто = кэширование отключено | `""` |

---

## API Reference

### Формат ответов

Все ответы оборачиваются в единый envelope:

```json
// Успех
{ "data": { ... } }

// Ошибка
{ "error": "описание ошибки" }
```

### HTTP коды

| Код | Причина |
|---|---|
| `200` | Успешный запрос |
| `201` | Ресурс создан |
| `204` | Успешно, тело отсутствует (DELETE) |
| `401` | Нет токена / токен невалиден / неверные credentials |
| `403` | Недостаточно прав (роль) |
| `404` | Ресурс не найден |
| `409` | Конфликт (email уже существует) |
| `410` | Gone — пользователь удалён |
| `422` | Ошибка валидации полей |
| `502` | Ошибка внешнего API |
| `500` | Внутренняя ошибка сервера |

---

## 🔓 Публичные эндпоинты

### Регистрация

```
POST /auth/register
```

```json
// Request
{
  "name": "Amir Seitkali",
  "email": "amir@example.com",
  "password": "secret123"
}

```

Правила:
- `name` — обязательное, до 100 символов
- `email` — обязательное, уникальное среди активных пользователей
- `password` — минимум 8 символов
- Пароль хэшируется bcrypt с cost=12, plain-текст никогда не сохраняется
- Роль по умолчанию: `user`

### Вход

```
POST /auth/login
```

```json
// Request
{
  "email": "amir@example.com",
  "password": "secret123"
}

// Response 200
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer"
  }
}
```

При неверном email **или** пароле возвращается одна и та же ошибка `401 invalid credentials` — защита от перебора пользователей.

---

## 🔐 Аутентификация

Все защищённые эндпоинты требуют заголовок:

```
Authorization: Bearer <access_token>
```

**JWT payload:**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "amir@example.com",
  "role": "user",
  "exp": 1716000000,
  "iat": 1715996400
}
```

Токен проверяется на:
- Валидность подписи (HMAC-SHA256)
- Истечение срока (`exp`)
- Корректный алгоритм (защита от `alg:none` атак)

---

## 👤 Пользователи

### Текущий пользователь

```
GET /users/me
Authorization: Bearer <token>
```

```json
// Response 200
{
  "data": {
    "id": "550e8400-...",
    "name": "Amir Seitkali",
    "email": "amir@example.com",
    "role": "user",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

> `password_hash` никогда не включается в ответ (поле помечено `json:"-"`).

### Список пользователей *(только admin)*

```
GET /users
Authorization: Bearer <admin_token>
```

### Получить пользователя по ID *(только admin)*

```
GET /users/{id}
Authorization: Bearer <admin_token>
```

### Удалить пользователя *(только admin, soft delete)*

```
DELETE /users/{id}
Authorization: Bearer <admin_token>
→ 204 No Content
```

Выставляет `deleted_at = NOW()`. Запись остаётся в БД. Все последующие запросы для удалённого пользователя возвращают `410 Gone`.

---

## 🌍 Города

Все операции привязаны к текущему пользователю из JWT — `user_id` не передаётся в URL.

### Добавить город

```
POST /cities
Authorization: Bearer <token>
```

```json
// Request
{ "name": "Almaty" }

// Response 201
{
  "data": {
    "id": "...",
    "user_id": "...",
    "name": "Almaty",
    "created_at": "2024-01-15T10:31:00Z"
  }
}
```

### Список городов

```
GET /cities
Authorization: Bearer <token>
```

### Удалить город

```
DELETE /cities/{city_id}
Authorization: Bearer <token>
→ 204 No Content
```

---

## 🌦 Погода

### Получить погоду по всем городам

```
GET /weather
Authorization: Bearer <token>
```

Запрашивает WeatherAPI параллельно для каждого города пользователя. Если один город недоступен — он пропускается (логируется), остальные возвращаются. Каждый успешный результат сохраняется в `weather_history`.

```json
// Response 200
{
  "data": {
    "user_id": "550e8400-...",
    "results": [
      {
        "city": "Almaty",
        "temp_c": 18.5,
        "feels_like_c": 17.2,
        "humidity": 45,
        "wind_kph": 12.3,
        "condition": {
          "text": "Partly cloudy",
          "icon": "//cdn.weatherapi.com/weather/64x64/day/116.png"
        }
      },
      {
        "city": "Astana",
        "temp_c": 14.0,
        "feels_like_c": 12.5,
        "humidity": 55,
        "wind_kph": 20.1,
        "condition": {
          "text": "Sunny",
          "icon": "//cdn.weatherapi.com/weather/64x64/day/113.png"
        }
      }
    ]
  }
}
```

### История запросов погоды

```
GET /weather/history?city=Almaty&limit=10
Authorization: Bearer <token>
```

| Параметр | Обязательный | Описание |
|---|---|---|
| `city` | ✅ | Фильтр по городу (case-insensitive) |
| `limit` | | Максимальное количество записей (целое > 0) |

Результаты отсортированы по `requested_at DESC`.

```json
// Response 200
{
  "data": [
    {
      "id": "...",
      "user_id": "...",
      "city": "Almaty",
      "temp_c": 18.5,
      "feels_like_c": 17.2,
      "humidity": 45,
      "wind_kph": 12.3,
      "condition": "Partly cloudy",
      "requested_at": "2024-01-15T10:35:00Z"
    }
  ]
}
```

---

## 🗺 Локации

### Получить штаты/области страны

```
GET /locations/countries/{country}/states
Authorization: Bearer <token>
```

`country` — ISO2 код страны (`KZ`, `US`, `DE` и т.д.). Регистр не важен.

```json
// GET /locations/countries/KZ/states
// Response 200
{
  "data": [
    { "name": "Almaty", "iso2": "ALA" },
    { "name": "Astana", "iso2": "AST" },
    { "name": "Akmola Region", "iso2": "AKM" }
  ]
}
```

---

## 🛡 Роли и авторизация

| Эндпоинт | `user` | `admin` |
|---|---|---|
| `POST /auth/register` | ✅ | ✅ |
| `POST /auth/login` | ✅ | ✅ |
| `GET /users/me` | ✅ | ✅ |
| `POST /cities` | ✅ | ✅ |
| `GET /cities` | ✅ | ✅ |
| `DELETE /cities/{id}` | ✅ | ✅ |
| `GET /weather` | ✅ | ✅ |
| `GET /weather/history` | ✅ | ✅ |
| `GET /locations/countries/{c}/states` | ✅ | ✅ |
| `GET /users` | ❌ | ✅ |
| `GET /users/{id}` | ❌ | ✅ |
| `DELETE /users/{id}` | ❌ | ✅ |

Для присвоения роли `admin` — обновите запись напрямую в БД:
```sql
UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';
```

---

## База данных

### Схема

```sql
users
  id            TEXT PRIMARY KEY
  name          TEXT NOT NULL
  email         TEXT NOT NULL
  password_hash TEXT NOT NULL        -- bcrypt, cost 12
  role          TEXT DEFAULT 'user'  -- CHECK: 'user' | 'admin'
  created_at    TIMESTAMPTZ
  updated_at    TIMESTAMPTZ
  deleted_at    TIMESTAMPTZ          -- NULL = активен (soft delete)

cities
  id          TEXT PRIMARY KEY
  user_id     TEXT REFERENCES users(id) ON DELETE CASCADE
  name        TEXT NOT NULL
  created_at  TIMESTAMPTZ
  UNIQUE (user_id, name)

weather_history
  id           TEXT PRIMARY KEY
  user_id      TEXT REFERENCES users(id) ON DELETE CASCADE
  city         TEXT
  temp_c       NUMERIC(5,2)
  feels_like   NUMERIC(5,2)
  humidity     INT
  wind_kph     NUMERIC(6,2)
  condition    TEXT
  requested_at TIMESTAMPTZ
```

### Индексы

- `users_email_active_idx` — `UNIQUE (email) WHERE deleted_at IS NULL` — удалённый пользователь не блокирует повторную регистрацию того же email
- `cities_user_id_idx` — быстрый lookup городов по пользователю
- `weather_history_user_city_idx` — `(user_id, LOWER(city), requested_at DESC)` — покрывает основной запрос истории без сортировки в памяти

---

## Команды

```bash
make run          # запустить локально
make build        # собрать бинарник → bin/api
make test         # тесты с -race и coverage
make lint         # golangci-lint
make migrate-up   # применить миграции (нужна DATABASE_URL)
make gen-secret   # сгенерировать JWT_SECRET через openssl
make docker-up    # поднять через docker compose
make docker-down  # остановить + удалить volumes
make docker-logs  # стриминг логов API
make tidy         # go mod tidy
```

---

## Полный пример использования

```bash
BASE=http://localhost:8080

# 1. Регистрация
RESP=$(curl -s -X POST $BASE/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Amir","email":"amir@example.com","password":"secret123"}')
TOKEN=$(echo $RESP | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
AUTH="Authorization: Bearer $TOKEN"

# 2. Текущий пользователь
curl -s $BASE/users/me -H "$AUTH" | jq .

# 3. Добавить города
curl -s -X POST $BASE/cities -H "$AUTH" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Almaty"}'

curl -s -X POST $BASE/cities -H "$AUTH" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Astana"}'

# 4. Список городов
curl -s $BASE/cities -H "$AUTH" | jq .

# 5. Погода по всем городам
curl -s $BASE/weather -H "$AUTH" | jq .

# 6. История с фильтром
curl -s "$BASE/weather/history?city=Almaty&limit=5" -H "$AUTH" | jq .

# 7. Штаты Казахстана
curl -s $BASE/locations/countries/KZ/states -H "$AUTH" | jq .

# 8. Удалить город
CITY_ID="<id из шага 4>"
curl -s -X DELETE $BASE/cities/$CITY_ID -H "$AUTH"

# 9. Назначить себя админом (через psql)
psql $DATABASE_URL -c "UPDATE users SET role='admin' WHERE email='amir@example.com';"

# 10. Новый логин (токен будет с role=admin)
RESP=$(curl -s -X POST $BASE/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"amir@example.com","password":"secret123"}')
ADMIN_TOKEN=$(echo $RESP | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
ADMIN_AUTH="Authorization: Bearer $ADMIN_TOKEN"

# 11. Список всех пользователей (admin only)
curl -s $BASE/users -H "$ADMIN_AUTH" | jq .
```

---

## Лицензия

MIT
