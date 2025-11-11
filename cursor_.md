# Обсуждение архитектуры бэкенда проекта
_Exported on 11/8/2025 at 17:45:47 GMT+5 from Cursor (2.0.69)_

---

**User**

общайся на русском.

привет, сейчас тут пустой проект. нам нужно написать бэкенд для одного проекта. давай начнем с того что пропишем архитектуру проекта, какие у нас будут папки и как все это будет выглядеть. я уже создал папку auth-service, начнем с сервиса авторизации. проект будет на go, вместе с postgres. 



вот примерное виденье авторизации:

Авторизация руководителя: (Alt1)

Поле ввода логина

Поле ввода пароля

Кнопка "Войти"

Авторизация пользователя:(Alt2)

Поле ввода номера телефона

Кнопка "Отправить SMS"

После отправки SMS:

Отображается замаскированный номер (последние 4 цифры видны)

Поле для ввода кода из SMS

Кнопка "Подтвердить"

При неверном коде показывается ошибка "Неверный код"

При верном коде происходит успешная авторизация

также, вот структура бд, реши какие таблицы будут относится к нашему сервису авторизации

////////////////////////////////////////////////////////

//  Организации и пользователи

////////////////////////////////////////////////////////

Table organization {

  id              uuid [pk]

  parent_org_id   uuid [ref: > organization.id]   // Иерархия: Akimat -> TOO -> Contractor

  type            varchar(20)                     // AKIMAT, TOO, CONTRACTOR

  name            varchar(255)

  bin             varchar(12)

  head_full_name  varchar(255)

  address         varchar(255)

  phone           varchar(32)

  is_active       bool

  created_at      timestamp

}

Table user {

  id              uuid [pk]

  organization_id uuid [ref: > organization.id]

  role            varchar(30)                     // AKIMAT_ADMIN, TOO_ADMIN, CONTRACTOR_ADMIN, DRIVER

  login           varchar(100)                    // Только для Акимата

  password_hash   varchar(255)                    // Только для Акимата

  phone           varchar(32)                     // Для SMS-авторизации

  driver_id       uuid [ref: > driver.id, null]   // Если это водитель

  is_active       bool

  created_at      timestamp

}

////////////////////////////////////////////////////////

//  Водители и техника

////////////////////////////////////////////////////////

Table driver {

  id              uuid [pk]

  contractor_id   uuid [ref: > organization.id]   // Должен быть CONTRACTOR

  full_name       varchar(255)

  birth_year      int

  iin             varchar(12)

  phone           varchar(32)

  photo_url       varchar(255)

  is_active       bool

  created_at      timestamp

}

Table vehicle {

  id              uuid [pk]

  contractor_id   uuid [ref: > organization.id]   // Владелец техники — подрядчик

  driver_id       uuid [ref: > driver.id, null]   // Текущий закрепленный водитель (опционально)

  plate_number    varchar(16)

  brand           varchar(100)

  model           varchar(100)

  color           varchar(50)

  year            int

  body_volume_m3  numeric(6,2)                    // Объем кузова

  photo_url       varchar(255)

  is_active       bool

  created_at      timestamp

}

////////////////////////////////////////////////////////

//  GPS-оборудование и треки

////////////////////////////////////////////////////////

Table gps_device {

  id              uuid [pk]

  vehicle_id      uuid [ref: > vehicle.id]        // Привязка трекера к машине

  imei            varchar(32)

  provider        varchar(50)

  installed_at    timestamp

  is_active       bool

}

Table gps_point {

  id              bigserial [pk]

  gps_device_id   uuid [ref: > gps_device.id]

  vehicle_id      uuid [ref: > vehicle.id]

  fix_time        timestamp

  latitude        numeric(10,7)

  longitude       numeric(10,7)

  speed_kmh       numeric(6,2)

  heading_deg     numeric(5,1)

  location_point  geometry                         // POINT для гео-запросов

  created_at      timestamp

}

////////////////////////////////////////////////////////

//  Участки уборки и полигоны вывоза снега

////////////////////////////////////////////////////////

Table cleaning_area {

  id              uuid [pk]

  name            varchar(255)

  description     text

  geometry        geometry                        // POLYGON/MULTIPOLYGON — участок уборки

  created_by_org  uuid [ref: > organization.id]   // Akimat или TOO

  is_active       bool

  created_at      timestamp

}

Table polygon {

  id              uuid [pk]

  name            varchar(255)

  address         varchar(255)

  geometry        geometry                        // Контур полигона вывоза снега

  is_active       bool

  created_at      timestamp

}

Table camera {

  id               uuid [pk]

  polygon_id       uuid [ref: > polygon.id]       // Камера установлена на конкретном полигоне

  type             varchar(20)                    // 'LPR' или 'VOLUME'

  entrance_label   varchar(50)                    // Опционально: "Въезд 1", "Въезд 2" — группировка пар камер

  description      varchar(255)

  location_point   geometry

  is_active        bool

  created_at       timestamp

}

////////////////////////////////////////////////////////

//  Тикеты и назначения

////////////////////////////////////////////////////////

Table ticket {

  id                  uuid [pk]

  cleaning_area_id    uuid [ref: > cleaning_area.id]

  created_by_org_id   uuid [ref: > organization.id] // Кто выдал (Akimat/TOO)

  contractor_id       uuid [ref: > organization.id] // Кому поручено (CONTRACTOR)

  status              varchar(20)                   // OPEN, IN_PROGRESS, DONE, EXPIRED, PARTIAL и т.д.

  planned_start_at    timestamp

  planned_end_at      timestamp

  description         text

  created_at          timestamp

}

Table ticket_assignment {

  id              uuid [pk]

  ticket_id       uuid [ref: > ticket.id]

  driver_id       uuid [ref: > driver.id]

  vehicle_id      uuid [ref: > vehicle.id]

  assigned_at     timestamp

}

////////////////////////////////////////////////////////

//  События камер

////////////////////////////////////////////////////////

Table lpr_event {

  id               uuid [pk]

  camera_id        uuid [ref: > camera.id]        // Камера типа LPR

  event_time       timestamp

  event_type       varchar(10)                    // 'ENTRY' или 'EXIT'

  plate_number     varchar(16)

  confidence       numeric(4,2)

  plate_image_url  varchar(255)                   // Фото номера (въезд или выезд)

}

Table volume_event {

  id               uuid [pk]

  camera_id        uuid [ref: > camera.id]        // Камера типа VOLUME

  event_time       timestamp

  event_type       varchar(10)                    // 'ENTRY' или 'EXIT'

  volume_m3        numeric(6,2)                   // Для ENTRY — загруженный кузов, для EXIT — должно быть близко к 0

  body_image_url   varchar(255)                   // Фото кузова (загружен/пустой)

}

////////////////////////////////////////////////////////

//  Рейсы (Trip) — один цикл: участок -> полигон -> выезд

////////////////////////////////////////////////////////

Table trip {

  id                       uuid [pk]

  ticket_id                uuid [ref: > ticket.id]   // Обязательно: рейс всегда относится к конкретному тикету

  driver_id                uuid [ref: > driver.id]

  vehicle_id               uuid [ref: > vehicle.id]

  polygon_id               uuid [ref: > polygon.id]  // На какой полигон привезли снег

  // Времена ключевых точек

  start_time               timestamp                 // Старт работы по рейсу (от участка), по GPS/логике

  polygon_entry_time       timestamp                 // Время въезда на полигон (по ENTRY-событиям)

  polygon_exit_time        timestamp                 // Время выезда с полигона (по EXIT-событиям)

  // Привязка к событиям камер (въезд с загрузкой)

  entry_lpr_event_id       uuid [ref: > lpr_event.id]     // LPR ENTRY с полным кузовом

  entry_volume_event_id    uuid [ref: > volume_event.id]  // VOLUME ENTRY с измеренным объемом

  // Привязка к событиям камер (выезд с пустым кузовом)

  exit_lpr_event_id        uuid [ref: > lpr_event.id]     // LPR EXIT того же номера

  exit_volume_event_id     uuid [ref: > volume_event.id]  // VOLUME EXIT — проверка пустого кузова

  // Объемы

  detected_volume_entry    numeric(6,2)              // Объем при въезде (из entry_volume_event)

  detected_volume_exit     numeric(6,2)              // Объем при выезде (из exit_volume_event), ожидается ≈ 0

  status                   varchar(30)               // Статус рейса (см. ниже)

  created_at               timestamp

  // Возможные значения status:

  // 'CONFIRMED'              - все проверки пройдены:

  // 'NO_AREA_VISIT'          - не зафиксировано захода в целевой участок по тикету

  // 'FOREIGN_AREA'           - работа (долгое пребывание) в чужом участке без назначения

  // 'MISMATCH_PLATE'         - не удалось определить номер

  // 'MISMATCH_VEHICLE'       - машина не в списке разрешенных по тикету

  // 'VOLUME_TOO_LOW'         - объем существенно ниже ожидаемого (подозрение на недогруз)

  // 'EXIT_NOT_EMPTY'         - при выезде с полигона кузов не пустой

  // 'MANUAL_CHECK'           - автоматически помечен как требующий ручной проверки

}

////////////////////////////////////////////////////////

//  SMS-коды для авторизации

////////////////////////////////////////////////////////

Table sms_code {

  id              uuid [pk]

  phone           varchar(32)

  code            varchar(10)

  expires_at      timestamp

  is_used         bool

}

прикрепляю тз проекта для контекста.@Техническое задание (ТЗ).pdf 

---

**Cursor**

## Архитектура Auth Service

- **Основной стек**: Go 1.22+, `go mod`; REST API на `chi`/`echo`; конфигурация через `.env` + Viper; логирование `zerolog`; миграции `golang-migrate`; PostgreSQL 15; контейнеризация `Docker / docker-compose`.
- **Слои**: transport (`http`), usecase/service, repository, domain models, adapters (SMS, hashing), shared (`config`, `logger`, `db`).

### Предлагаемая структура
- `cmd/auth-service/main.go` — точка входа, инициализация модулей.
- `internal/`  
  - `app/` — сборка зависимостей, запуск сервиса.  
  - `config/` — чтение конфигов.  
  - `domain/` — сущности (`User`, `Organization`, `SmsCode`), ошибки.  
  - `transport/http/` — хендлеры, DTO, маршруты; подпапки `admin`, `user`, `middleware`.  
  - `service/` — бизнес-логика: `AdminAuthService`, `SmsAuthService`, `TokenService`.  
  - `repository/` — Postgres-репозитории (user, sms_code, org), `unit` тесты.  
  - `adapter/` — интеграции: `sms` (mock/real provider), `password` (bcrypt/argon2), `token` (JWT).  
  - `pkg/` — утилиты (например, time, tracing).  
- `migrations/` — SQL или goose миграции для таблиц auth-сервиса.
- `deploy/` — `docker-compose.yml`, helm/chart (позже).
- `docs/` — OpenAPI спецификация (`openapi.yaml`), описание потоков авторизации.

### Потоки авторизации

- **Руководитель (логин/пароль)**  
  - `POST /api/v1/admin/login` → проверка `login`, `password_hash`, генерация `access` + `refresh` токенов (JWT/Session).  
  - Возможность смены пароля (`PATCH /password`) и сброса (через future email/SMS).

- **Пользователь (SMS-код)**  
  - `POST /api/v1/user/request-code` → генерация кода, запись в `sms_code`, отправка через SMS-провайдер, маскировка номера в ответе.  
  - `POST /api/v1/user/verify-code` → проверка кода, маркер `is_used`, выдача токенов.  
  - Обработка ошибок: неверный код, истекший, превышен лимит.

- **Токены**  
  - `POST /api/v1/token/refresh`  
  - `POST /api/v1/logout` (инвалидация refresh, возможно `token_blacklist` таблица или короткий TTL).

### Таблицы, относящиеся к auth-сервису

- Обязательные: `organization`, `user`, `sms_code`.  
- Дополнительно при необходимости: отдельная таблица для refresh-токенов (`user_session`), если решим хранить.  
- Остальные сущности (водители, техника, тикеты и т.д.) не входят в зону ответственности auth-сервиса, но `user.driver_id` используется для связи при авторизации водителей.

### Следующие шаги

- Согласовать формат токенов (JWT + срок действия, refresh хранить?).
- Описать детальнее DTO и OpenAPI.
- Подготовить миграции для `user`, `organization`, `sms_code` и, при необходимости, `user_session`.
- Настроить docker-compose с Postgres, запустить базовый скелет приложения.

---

**User**

что это за структура, она обязательно должна быть такая?

---

**Cursor**

Это всего лишь предложенный каркас ― один из распространённых вариантов для Go‑сервисов. Ничего обязательного здесь нет: можно упростить, переименовать, объединить слои или использовать другую архитектурную философию (например, чистую архитектуру, hexagonal, feature-based и т.д.). Главное, чтобы структура помогала команде ориентироваться, отделяла бизнес-логику от инфраструктуры и не мешала развитию. Если хочется начать проще (например, несколько папок `handlers`, `services`, `repositories` без глубоких вложений), это тоже нормальный старт — потом всегда можно реорганизовать, когда станет понятно, каких модулей и интеграций реально много.

---

**User**

а какие еще есть структуры?

---

**Cursor**

Есть несколько популярных подходов — выбор зависит от размера команды, требований к масштабируемости и личных предпочтений.

1. **Classic layered (MVC)**  
   ```
   cmd/
   internal/
     handlers/   // HTTP-слой
     services/   // бизнес-логика
     repositories/ // работа с БД
   pkg/          // общие утилиты
   ```
   Плюсы: простой старт, понятные границы. Минусы: со временем может превращаться в «анемичную» модель, из–за связей слоёв усложняется изменение бизнес-логики.

2. **Feature (by module/domain)**  
   ```
   internal/
     auth/
       handler.go
       service.go
       repository.go
       dto.go
     tickets/
       ...
   shared/...
   ```
   Каждая предметная область в своей папке и содержит полный цикл: transport → use case → repo. Удобно для автономных команд и роста. Минус: дублируются инфраструктурные шаблоны, нужен дисциплинированный подход к общим пакетам.

3. **Hexagonal / Ports & Adapters**  
   ```
   internal/
     domain/
       auth/
         entities.go
         service.go
         ports.go
     adapters/
       http/
       postgres/
       sms/
     app/
       bootstrap.go
   ```
   Бизнес-логика изолирована, внешние интеграции подключаются как адаптеры. Хорошо для сложных систем и тестируемости, но требует больше планирования и дисциплины.

4. **Clean Architecture (use cases + interfaces)**  
   ```
   internal/
     entity/
     usecase/
     interface/
       http/
       repository/
     infrastructure/
       db/
       sms/
   ```
   Похож на hexagonal: домен в центре, остальное — детали. Удобно, когда много сценариев, нужно выдерживать зависимость только «внутрь». Минус — некоторый ceremony, сложнее новичкам.

5. **Package-per-layer без deep nesting**  
   ```
   internal/
     api/
     auth/
       service.go
       store.go
     db/
     config/
   ```
   Упрощённый гибрид; можно комбинировать с feature-подходом. Подходит для средних проектов: минимум вложенности, но есть разделение по зональности.

6. **Go-kit style (service interfaces, endpoints, transports)**  
   ```
   internal/
     auth/
       service.go
       endpoint.go
       transport_http.go
       transport_grpc.go
   ```
   Разделение на service → endpoint → transport. Удобно, если нужен гибкий выбор транспорта (REST, gRPC). Минус — overkill для маленьких сервисов.

7. **Microservice skeletons типа Kratos / Buffalo / Gin boilerplates**  
   Используют свои соглашения, например, Kratos: `internal/biz`, `internal/data`, `internal/service`, `internal/server`. Хорошо, если хочешь следовать стандартам фреймворка.

---

Если мы в начале пути, можно начать с простого layered или feature-based. Позже, когда появятся интеграции (SMS, разные источники данных), можно перейти к hexagonal/clean, сохранив понятную доменную структуру.

```plaintext
   cmd/
   internal/
     handlers/   // HTTP-слой
     services/   // бизнес-логика
     repositories/ // работа с БД
   pkg/          // общие утилиты
```

```plaintext
   internal/
     auth/
       handler.go
       service.go
       repository.go
       dto.go
     tickets/
       ...
   shared/...
```

```plaintext
   internal/
     domain/
       auth/
         entities.go
         service.go
         ports.go
     adapters/
       http/
       postgres/
       sms/
     app/
       bootstrap.go
```

```plaintext
   internal/
     entity/
     usecase/
     interface/
       http/
       repository/
     infrastructure/
       db/
       sms/
```

```plaintext
   internal/
     api/
     auth/
       service.go
       store.go
     db/
     config/
```

```plaintext
   internal/
     auth/
       service.go
       endpoint.go
       transport_http.go
       transport_grpc.go
```

---

**User**

думаю между classic и hexagonal

---

**Cursor**

Если сравнивать глазами нашего auth-сервиса:

### Classic layered
- **Плюсы**: быстрый старт; меньше файлов; проще онбординг (handlers → services → repositories); хорошо ложится на простые CRUD/sync сценарии; легче писать код «на автомате».
- **Минусы**: слой `service` часто превращается в «мешок» бизнес-логики; сложно «инвертировать зависимость» — домен знает об инфраструктуре; при расширении (например, добавлении SMS-провайдера, кэша) зависимости могут «просачиваться» вниз; тестирование use case’ов может требовать моков репозитория + вспомогательных структур.

### Hexagonal (ports & adapters)
- **Плюсы**: домен (`core`) ничего не знает про внешние детали; легко изолировать бизнес-логику (авторизация по логину/паролю, выдача токенов) и покрывать ее тестами; добавление новых адаптеров (новый SMS-провайдер, замена БД на кэш, GRPC поверх HTTP) — без переписывания доменного слоя; удобно разделять ответственность между командами/патчами; можно внедрить репозиторий-подложку для тестов.
- **Минусы**: больше «церемоний» — файлы с интерфейсами (`ports`), отдельные пакеты под адаптеры, bootstrap; требуется дисциплина в зависимостях; новичкам нужно объяснять принципы; на старте выглядит как избыточное дробление.

### Что выбрать
- Если команда маленькая, сервис первый и хочется быстро получить работающую авторизацию без большого количества интеграций — **classic layered** будет проще. Можно заложить базовую дисциплину (чёткий `internal/service`), а потом постепенно реорганизовывать.
- Если планируется интеграция нескольких внешних каналов (SMS, e-mail, push), возможно появление нескольких способов авторизации (логин/пароль, OAuth, SSO), рост числа сервисов, потребность в unit-тестах бизнес-правил — **hexagonal** избавит от боли при рефакторинге.
- Можно начать «light hexagonal»: держать доменные сервисы/интерфейсы в пакете `internal/core`, а адаптеры (HTTP, Postgres, SMS) рядом (`internal/adapter`), но без полного дробления слоями — получится управляемый компромисс.

Если сомневаешься, но предполагаешь рост и внешние интеграции — я бы выбрал hexagonal как базу, даже если сначала он кажется более громоздким: проще сразу соблюдать границы, чем потом выкорчёвывать зависимости из «классического» слоя.

---

**User**

сроки сжатые, в планах сделать легкое но надежное решение

---

**Cursor**

Тогда берём что-то средней сложности: классический layered, но с небольшими ограничениями, чтобы не скатиться в хаос:

- `cmd/auth-service/main.go` — старт сервиса.
- `internal/config`, `internal/logger`, `internal/db` — инфраструктура.
- `internal/http` — роутеры/хендлеры (`admin`, `user`, `middleware`).
- `internal/service` — бизнес-логика (`AdminAuthService`, `SmsAuthService`, `TokenService`).
- `internal/repository` — Postgres-слой (user, sms_code, organization).
- `internal/adapter` (опционально) — SMS-провайдер, хэширование.
- `internal/model` или `internal/domain` — сущности/DTO, чтобы не плодить дубли.

Такая схема:
- запускается быстро (мало файлов/сложности);
- легко объяснить новичкам;
- без проблем допилим тесты (моки на `repository` и `adapter`);
- позже, если потребуется, можно вынести `service` + `model` в `core`, `repository`/`http`/`adapter` в отдельные пакеты и превратить всё в hexagonal.

Главное — договориться:
- сервисы не тянут напрямую инфраструктуру (внедряем через интерфейсы);
- хендлеры не лазят в репозитории, идут через сервисы;
- DTO для API и доменные структуры разделены, чтобы соблюсти границы.

Так мы успеем к срокам, но оставим задел на рост и refactor.

---

**User**

да, давай остановимся на таком решении. теперь, что мы будет использовать для http логики? какие варианты?

---

**Cursor**

Для HTTP-слоя в Go есть несколько проверенных опций:

- **Standard library (`net/http`) + `chi`/`gorilla/mux`**  
  - `chi` — лёгкий router, поддерживает middleware, группировку маршрутов, контекст, удобен для REST.  
  - `gorilla/mux` — классика, но проект поддерживается минимально; лучше `chi`.  
  - Плюсы: минимальная зависимость, хорошая производительность, простота. Минусы: чуть больше кода руками (валидация, JSON-ответы).

- **Echo**  
  - Быстрый фреймворк, дружелюбный API, встроенные middleware, хэндлеры компактнее.  
  - Плюсы: быстро пишутся CRUD, встроенный binder/validator. Минусы: своя модель контекста, специфичный стиль.

- **Fiber**  
  - Вдохновлён Express.js, очень быстрый, простой синтаксис.  
  - Минусы: собственный контекст, не совместим напрямую с `net/http` middleware, чуть менее идиоматичный для Go.

- **Gin**  
  - Считается «де-факто» стандартом, но внутри использует собственный router; хорошая экосистема, много примеров.  
  - Минусы: сложнее тестировать хэндлеры из-за контекста Gin.

- **Goa / go-swagger / kratos**  
  - Более тяжёлые, если нужна декларативная спецификация и генерация кода; для нашего «легко и надёжно» — overkill.

Для auth-сервиса я бы предложил:

1. **`chi`** + стандартная библиотека — минимализм, легко поддерживать.  
2. **Gin** — если нужна богатая экосистема и привычные middlewares.  
3. **Echo** — если нравится синтаксис и встроенные helpers.

Если планируем расширять сервисы, писать middleware сами и хотим оставаться ближе к стандарту — `chi` оптимален.

---

**User**

ок, давай использовать gin. тогда составь план, что мы сейчас должны сделать для разработки сервиса авторизации

---

**Cursor**

## План разработки Auth Service (Gin + PostgreSQL)

- **Инициализация проекта**
  - создать `go.mod`, подтянуть `gin`, `pgx`, `zerolog`, `viper`, `bcrypt`/`argon2`.
  - настроить структуру (`cmd`, `internal/...`, `migrations`), базовый `main.go` с конфигом и логгером.

- **Конфигурация и инфраструктура**
  - `internal/config`: чтение `.env` (HTTP порт, DSN, JWT секреты, SMS-провайдер).
  - `internal/logger`: инициализация `zerolog`.
  - `internal/db`: пул `pgxpool`, миграции (CLI или `golang-migrate`).
  - `internal/http/server`: запуск Gin, подключение middleware (логирование, recovery, CORS, request ID).

- **Модели и репозитории**
  - `internal/model`: `User`, `Organization`, `SmsCode`, DTO для токенов.
  - `internal/repository/user`, `sms_code`, `organization`: CRUD, выборки по логину/телефону, транзакции.

- **Бизнес-логика (`internal/service`)**
  - `AdminAuthService`: проверка логина/пароля, генерация токенов.
  - `SmsAuthService`: выдача/валидация кодов, лимиты, отметка `is_used`.
  - `TokenService`: создание access/refresh, валидация, инвалидация (при необходимости `user_session`).

- **Интеграции (`internal/adapter`)**
  - `sms`: mock + интерфейс для провайдера (позже реальная интеграция).
  - `password`: хэширование/проверка (bcrypt/argon2).
  - `token`: реализация JWT (например, `github.com/golang-jwt/jwt/v5`).

- **HTTP-слой (`internal/http`)**
  - маршруты: `POST /admin/login`, `POST /user/request-code`, `POST /user/verify-code`, `POST /token/refresh`, `POST /logout`.
  - DTO запрос/ответ, биндинг/валидация, оформление ошибок.
  - middleware: аутентификация по access токену, rate limit (позже).

- **Миграции**
  - SQL-файлы для `organization`, `user`, `sms_code`, (опц.) `user_session`.
  - запустить миграции через `docker-compose` или отдельную команду.

- **Тесты и утилиты**
  - unit-тесты сервисов (моки репозитория).
  - integration-тесты минимально для auth flow.
  - Makefile/скрипты: `make run`, `make test`, `make migrate`.

- **Документация**
  - OpenAPI схема для auth endpoint.
  - README: запуск, переменные окружения, миграции.

- **Инфра**
  - `docker-compose` с PostgreSQL, возможно mock SMS сервис.
  - GitHub workflow или локальный check (lint/test) при необходимости.

Это даст работоспособный базовый auth-сервис, готовый к расширению.

---

**User**

тесты и документация не нужны, нужно быстро реализовать проект. также, давай использовать gorm.

---

**Cursor**

## Обновлённый план (без тестов/доков, с GORM)

- **Инициализация**
  - `go mod init auth-service`
  - Зависимости: `gin`, `gorm.io/gorm`, `gorm.io/driver/postgres`, `viper` (конфиг), `zerolog` (логгер), `golang-jwt/jwt/v5`, `bcrypt`.
  - Создать структуру: `cmd/auth-service/main.go`, `internal/...`, `migrations/`.

- **Инфраструктура**
  - `internal/config`: чтение `.env` / переменных.
  - `internal/logger`: инициализация `zerolog`.
  - `internal/db`: подключение GORM к Postgres, настройка пула. Определить миграции через `AutoMigrate` (для скорости).
  - `docker-compose.yml`: Postgres + сервис.

- **Модели и миграции (GORM)**
  - `internal/model`: `Organization`, `User`, `SmsCode`, (при необходимости `UserSession`).
  - В `internal/db` вызвать `AutoMigrate` для перечисленных сущностей.

- **Репозитории (`internal/repository`)**
  - Реализация поверх GORM: поиск пользователя по логину/телефону, создание/обновление `SmsCode`, пометка `is_used`.
  - Минимум CRUD-операций, только необходимые для авторизации.

- **Сервисы (`internal/service`)**
  - `AdminAuthService`: проверка логина/пароля (bcrypt), выдача JWT (access + refresh).
  - `SmsAuthService`: генерация случайного кода, сохранение, проверка/истечение.
  - `TokenService`: создание/парсинг JWT, валидация. Для быстроты можно хранить refresh только в JWT (без БД).

- **Адаптеры (`internal/adapter`)**
  - `password`: обёртка вокруг bcrypt.
  - `sms`: заглушка (логируем код вместо реальной отправки); интерфейс на будущее.
  - `token`: генерация JWT (использовать secret из конфига, указать TTL).

- **HTTP (`internal/http`)**
  - Настроить Gin router, middleware (логирование, recovery, CORS при необходимости).
  - Эндпоинты:
    - `POST /admin/login`
    - `POST /user/request-code`
    - `POST /user/verify-code`
    - `POST /token/refresh`
    - `POST /logout` (можно просто «чёрный список текущего refresh» если понадобится, либо заглушка).
  - DTO запрос/ответ, минимальная валидация через `binding:"required"`.

- **Запуск (`cmd/auth-service/main.go`)**
  - Загрузить конфиг → логгер → GORM → миграции → собрать сервисы и роуты → запустить HTTP.

- **Быстрые проверки**
  - Прогон `go run cmd/auth-service/main.go` и вручную проверить REST запросами (Postman/curl).

Так получаем рабочий минимум без тестов и документации, используя GORM для ускорения разработки.

---

**User**

бд мы будем запускать через docker compose?

---

**Cursor**

Да, так быстрее всего. Добавим `docker-compose.yml` с сервисами:

- `postgres` — образ `postgres:15`, настройки БД/пользователя/пароля через `.env`.
- `auth-service` (по желанию) — чтобы запускать Go-сервис внутри контейнера; на старте можно оставить только Postgres и запускать приложение локально (`go run`), пока идёт разработка.

План действий:
1. Создаём `.env` с `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `DB_HOST=postgres`, `DB_PORT=5432`.
2. В `docker-compose.yml` описываем `postgres`, настраиваем volume для данных, пробрасываем порт `5432`.
3. Запускаем `docker compose up -d`.
4. Приложение в конфигурации получает DSN вида `postgres://user:pass@localhost:5432/db?sslmode=disable` (для локального запуска) или `postgres://user:pass@postgres:5432/db?sslmode=disable` (когда сервис будет внутри compose).

Позже можно добавить сервис `auth-service` в compose и подключить его к той же сети, но для стартовой разработки достаточно отдельной БД в контейнере.

---

**User**

отлично, хороший план. теперь создай auth-service по нашему плану. вот небольшие дополнения:

вот такие должны быть эндпоинты.

POST /auth/login — вход руководителя по логину и паролю

POST /auth/send-code — отправка SMS-кода пользователю

POST /auth/verify-code — проверка SMS-кода пользователем

POST /auth/refresh — обновление access-токена через refresh-токен

POST /auth/logout — завершение сессии, удаление refresh-токена

GET /auth/me — получение информации о текущем пользователе

вот таблицы бд, используй те, которые нужны для auth-service:

////////////////////////////////////////////////////////
//  Организации и пользователи
////////////////////////////////////////////////////////

Table organization {
  id              uuid [pk]
  parent_org_id   uuid [ref: > organization.id]   // Иерархия: Akimat -> TOO -> Contractor
  type            varchar(20)                     // AKIMAT, TOO, CONTRACTOR
  name            varchar(255)
  bin             varchar(12)
  head_full_name  varchar(255)
  address         varchar(255)
  phone           varchar(32)
  is_active       bool
  created_at      timestamp
}

Table user {
  id              uuid [pk]
  organization_id uuid [ref: > organization.id]
  role            varchar(30)                     // AKIMAT_ADMIN, TOO_ADMIN, CONTRACTOR_ADMIN, DRIVER
  login           varchar(100)                    // Только для Акимата
  password_hash   varchar(255)                    // Только для Акимата
  phone           varchar(32)                     // Для SMS-авторизации
  driver_id       uuid [ref: > driver.id, null]   // Если это водитель
  is_active       bool
  created_at      timestamp
}

////////////////////////////////////////////////////////
//  Водители и техника
////////////////////////////////////////////////////////

Table driver {
  id              uuid [pk]
  contractor_id   uuid [ref: > organization.id]   // Должен быть CONTRACTOR
  full_name       varchar(255)
  birth_year      int
  iin             varchar(12)
  phone           varchar(32)
  photo_url       varchar(255)
  is_active       bool
  created_at      timestamp
}

Table vehicle {
  id              uuid [pk]
  contractor_id   uuid [ref: > organization.id]   // Владелец техники — подрядчик
  driver_id       uuid [ref: > driver.id, null]   // Текущий закрепленный водитель (опционально)
  plate_number    varchar(16)
  brand           varchar(100)
  model           varchar(100)
  color           varchar(50)
  year            int
  body_volume_m3  numeric(6,2)                    // Объем кузова
  photo_url       varchar(255)
  is_active       bool
  created_at      timestamp
}

////////////////////////////////////////////////////////
//  GPS-оборудование и треки
////////////////////////////////////////////////////////

Table gps_device {
  id              uuid [pk]
  vehicle_id      uuid [ref: > vehicle.id]        // Привязка трекера к машине
  imei            varchar(32)
  provider        varchar(50)
  installed_at    timestamp
  is_active       bool
}

Table gps_point {
  id              bigserial [pk]
  gps_device_id   uuid [ref: > gps_device.id]
  vehicle_id      uuid [ref: > vehicle.id]
  fix_time        timestamp
  latitude        numeric(10,7)
  longitude       numeric(10,7)
  speed_kmh       numeric(6,2)
  heading_deg     numeric(5,1)
  location_point  geometry                         // POINT для гео-запросов
  created_at      timestamp
}

////////////////////////////////////////////////////////
//  Участки уборки и полигоны вывоза снега
////////////////////////////////////////////////////////

Table cleaning_area {
  id              uuid [pk]
  name            varchar(255)
  description     text
  geometry        geometry                        // POLYGON/MULTIPOLYGON — участок уборки
  created_by_org  uuid [ref: > organization.id]   // Akimat или TOO
  is_active       bool
  created_at      timestamp
}

Table polygon {
  id              uuid [pk]
  name            varchar(255)
  address         varchar(255)
  geometry        geometry                        // Контур полигона вывоза снега
  is_active       bool
  created_at      timestamp
}

Table camera {
  id               uuid [pk]
  polygon_id       uuid [ref: > polygon.id]       // Камера установлена на конкретном полигоне
  type             varchar(20)                    // 'LPR' или 'VOLUME'
  entrance_label   varchar(50)                    // Опционально: "Въезд 1", "Въезд 2" — группировка пар камер
  description      varchar(255)
  location_point   geometry
  is_active        bool
  created_at       timestamp
}

////////////////////////////////////////////////////////
//  Тикеты и назначения
////////////////////////////////////////////////////////

Table ticket {
  id                  uuid [pk]
  cleaning_area_id    uuid [ref: > cleaning_area.id]
  created_by_org_id   uuid [ref: > organization.id] // Кто выдал (Akimat/TOO)
  contractor_id       uuid [ref: > organization.id] // Кому поручено (CONTRACTOR)
  status              varchar(20)                   // OPEN, IN_PROGRESS, DONE, EXPIRED, PARTIAL и т.д.
  planned_start_at    timestamp
  planned_end_at      timestamp
  description         text
  created_at          timestamp
}

Table ticket_assignment {
  id              uuid [pk]
  ticket_id       uuid [ref: > ticket.id]
  driver_id       uuid [ref: > driver.id]
  vehicle_id      uuid [ref: > vehicle.id]
  assigned_at     timestamp
}

////////////////////////////////////////////////////////
//  События камер
////////////////////////////////////////////////////////

Table lpr_event {
  id               uuid [pk]
  camera_id        uuid [ref: > camera.id]        // Камера типа LPR
  event_time       timestamp
  event_type       varchar(10)                    // 'ENTRY' или 'EXIT'
  plate_number     varchar(16)
  confidence       numeric(4,2)
  plate_image_url  varchar(255)                   // Фото номера (въезд или выезд)
}

Table volume_event {
  id               uuid [pk]
  camera_id        uuid [ref: > camera.id]        // Камера типа VOLUME
  event_time       timestamp
  event_type       varchar(10)                    // 'ENTRY' или 'EXIT'
  volume_m3        numeric(6,2)                   // Для ENTRY — загруженный кузов, для EXIT — должно быть близко к 0
  body_image_url   varchar(255)                   // Фото кузова (загружен/пустой)
}

////////////////////////////////////////////////////////
//  Рейсы (Trip) — один цикл: участок -> полигон -> выезд
////////////////////////////////////////////////////////

Table trip {
  id                       uuid [pk]

  ticket_id                uuid [ref: > ticket.id]   // Обязательно: рейс всегда относится к конкретному тикету

  driver_id                uuid [ref: > driver.id]
  vehicle_id               uuid [ref: > vehicle.id]

  polygon_id               uuid [ref: > polygon.id]  // На какой полигон привезли снег

  // Времена ключевых точек
  start_time               timestamp                 // Старт работы по рейсу (от участка), по GPS/логике
  polygon_entry_time       timestamp                 // Время въезда на полигон (по ENTRY-событиям)
  polygon_exit_time        timestamp                 // Время выезда с полигона (по EXIT-событиям)

  // Привязка к событиям камер (въезд с загрузкой)
  entry_lpr_event_id       uuid [ref: > lpr_event.id]     // LPR ENTRY с полным кузовом
  entry_volume_event_id    uuid [ref: > volume_event.id]  // VOLUME ENTRY с измеренным объемом

  // Привязка к событиям камер (выезд с пустым кузовом)
  exit_lpr_event_id        uuid [ref: > lpr_event.id]     // LPR EXIT того же номера
  exit_volume_event_id     uuid [ref: > volume_event.id]  // VOLUME EXIT — проверка пустого кузова

  // Объемы
  detected_volume_entry    numeric(6,2)              // Объем при въезде (из entry_volume_event)
  detected_volume_exit     numeric(6,2)              // Объем при выезде (из exit_volume_event), ожидается ≈ 0

  status                   varchar(30)               // Статус рейса (см. ниже)
  created_at               timestamp

  // Возможные значения status:
  // 'CONFIRMED'              - все проверки пройдены:
  // 'NO_AREA_VISIT'          - не зафиксировано захода в целевой участок по тикету
  // 'FOREIGN_AREA'           - работа (долгое пребывание) в чужом участке без назначения
  // 'MISMATCH_PLATE'         - не удалось определить номер
  // 'MISMATCH_VEHICLE'       - машина не в списке разрешенных по тикету
  // 'VOLUME_TOO_LOW'         - объем существенно ниже ожидаемого (подозрение на недогруз)
  // 'EXIT_NOT_EMPTY'         - при выезде с полигона кузов не пустой
  // 'MANUAL_CHECK'           - автоматически помечен как требующий ручной проверки
}

////////////////////////////////////////////////////////
//  SMS-коды для авторизации
////////////////////////////////////////////////////////

Table sms_code {
  id              uuid [pk]
  phone           varchar(32)
  code            varchar(10)
  expires_at      timestamp
  is_used         bool
}


работай в папке auth-service, сейчас мы только в ней.

---

**Cursor**

```dockercompose
version: "3.9"

services:
  postgres:
    image: postgres:15
    restart: unless-stopped
    environment:
      POSTGRES_USER: ${POSTGRES_USER:-auth_user}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-auth_pass}
      POSTGRES_DB: ${POSTGRES_DB:-auth_db}
    volumes:
      - pgdata:/var/lib/postgresql/data
    ports:
      - "${POSTGRES_PORT:-5432}:5432"

volumes:
  pgdata:

```

```go
package config

import (
	"time"

	"github.com/spf13/viper"
)

type HTTPConfig struct {
	Host string
	Port int
}

type DBConfig struct {
	DSN            string
	MaxOpenConns   int
	MaxIdleConns   int
	ConnMaxLifetime time.Duration
}

type JWTConfig struct {
	AccessSecret        string
	AccessTTL           time.Duration
	RefreshTTL          time.Duration
}

type SMSConfig struct {
	CodeTTL        time.Duration
	CodeLength     int
	DailySendLimit int
}

type Config struct {
	Environment string
	HTTP        HTTPConfig
	DB          DBConfig
	JWT         JWTConfig
	SMS         SMSConfig
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("./deploy")
	v.AddConfigPath("./internal/config")

	v.AutomaticEnv()

	setDefaults(v)

	_ = v.ReadInConfig()

	cfg := &Config{
		Environment: v.GetString("APP_ENV"),
		HTTP: HTTPConfig{
			Host: v.GetString("HTTP_HOST"),
			Port: v.GetInt("HTTP_PORT"),
		},
		DB: DBConfig{
			DSN:             v.GetString("DB_DSN"),
			MaxOpenConns:    v.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    v.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: v.GetDuration("DB_CONN_MAX_LIFETIME"),
		},
		JWT: JWTConfig{
			AccessSecret: v.GetString("JWT_ACCESS_SECRET"),
			AccessTTL:    v.GetDuration("JWT_ACCESS_TTL"),
			RefreshTTL:   v.GetDuration("JWT_REFRESH_TTL"),
		},
		SMS: SMSConfig{
			CodeTTL:        v.GetDuration("SMS_CODE_TTL"),
			CodeLength:     v.GetInt("SMS_CODE_LENGTH"),
			DailySendLimit: v.GetInt("SMS_DAILY_LIMIT"),
		},
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("HTTP_HOST", "0.0.0.0")
	v.SetDefault("HTTP_PORT", 8080)

	v.SetDefault("DB_DSN", "postgres://auth_user:auth_pass@localhost:5432/auth_db?sslmode=disable")
	v.SetDefault("DB_MAX_OPEN_CONNS", 25)
	v.SetDefault("DB_MAX_IDLE_CONNS", 10)
	v.SetDefault("DB_CONN_MAX_LIFETIME", time.Hour)

	v.SetDefault("JWT_ACCESS_SECRET", "supersecret")
	v.SetDefault("JWT_ACCESS_TTL", time.Minute*15)
	v.SetDefault("JWT_REFRESH_TTL", time.Hour*24*30)

	v.SetDefault("SMS_CODE_TTL", time.Minute*5)
	v.SetDefault("SMS_CODE_LENGTH", 6)
	v.SetDefault("SMS_DAILY_LIMIT", 10)
}

```

```go
package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

func New(env string) zerolog.Logger {
	level := zerolog.InfoLevel
	if strings.EqualFold(env, "development") {
		level = zerolog.DebugLevel
	}

	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	return zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()
}

```

```go
package db

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm"

	"github.com/nurpe/snowops-auth/internal/config"
	"github.com/nurpe/snowops-auth/internal/model"
)

func New(cfg config.DBConfig, log zerolog.Logger, appEnv string) (*gorm.DB, error) {
	gormLog := gormlogger.New(
		zerologAdapter{logger: log},
		gormlogger.Config{
			SlowThreshold:             time.Second,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			LogLevel:                  selectLogLevel(appEnv),
		},
	)

	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		Logger: gormLog,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		return nil, err
	}
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`).Error; err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&model.Organization{},
		&model.User{},
		&model.SmsCode{},
		&model.UserSession{},
	); err != nil {
		return nil, err
	}

	return db, nil
}

func HealthCheck(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec("SELECT 1").Error
}

func selectLogLevel(env string) gormlogger.LogLevel {
	if env == "development" {
		return gormlogger.Info
	}
	return gormlogger.Warn
}

type zerologAdapter struct {
	logger zerolog.Logger
}

func (z zerologAdapter) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return z
}

func (z zerologAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	z.logger.Info().Msgf(msg, data...)
}

func (z zerologAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	z.logger.Warn().Msgf(msg, data...)
}

func (z zerologAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	z.logger.Error().Msgf(msg, data...)
}

func (z zerologAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	event := z.logger.Debug()
	if err != nil {
		event = z.logger.Error().Err(err)
	}
	event.
		Dur("elapsed", elapsed).
		Int64("rows", rows).
		Msg(sql)
}

```

```go
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrganizationType string

const (
	OrganizationTypeAkimat     OrganizationType = "AKIMAT"
	OrganizationTypeToo        OrganizationType = "TOO"
	OrganizationTypeContractor OrganizationType = "CONTRACTOR"
)

type UserRole string

const (
	UserRoleAkimatAdmin     UserRole = "AKIMAT_ADMIN"
	UserRoleTooAdmin        UserRole = "TOO_ADMIN"
	UserRoleContractorAdmin UserRole = "CONTRACTOR_ADMIN"
	UserRoleDriver          UserRole = "DRIVER"
)

type Organization struct {
	ID            uuid.UUID        `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ParentOrgID   *uuid.UUID       `gorm:"type:uuid"`
	Type          OrganizationType `gorm:"type:varchar(20);not null"`
	Name          string           `gorm:"type:varchar(255);not null"`
	BIN           string           `gorm:"type:varchar(12)"`
	HeadFullName  string           `gorm:"type:varchar(255)"`
	Address       string           `gorm:"type:varchar(255)"`
	Phone         string           `gorm:"type:varchar(32)"`
	IsActive      bool             `gorm:"default:true"`
	CreatedAt     time.Time        `gorm:"autoCreateTime"`
}

type User struct {
	ID             uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null"`
	Role           UserRole   `gorm:"type:varchar(30);not null"`
	Login          *string    `gorm:"type:varchar(100)"`
	PasswordHash   *string    `gorm:"type:varchar(255)"`
	Phone          *string    `gorm:"type:varchar(32)"`
	DriverID       *uuid.UUID `gorm:"type:uuid"`
	IsActive       bool       `gorm:"default:true"`
	CreatedAt      time.Time  `gorm:"autoCreateTime"`

	Organization Organization `gorm:"foreignKey:OrganizationID"`
}

type SmsCode struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Phone     string    `gorm:"type:varchar(32);index;not null"`
	Code      string    `gorm:"type:varchar(10);not null"`
	ExpiresAt time.Time `gorm:"not null"`
	IsUsed    bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type UserSession struct {
	ID               uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID           uuid.UUID  `gorm:"type:uuid;not null;index"`
	RefreshTokenHash string     `gorm:"type:char(64);not null;uniqueIndex"`
	ExpiresAt        time.Time  `gorm:"not null"`
	RevokedAt        *time.Time
	UserAgent        string     `gorm:"type:varchar(255)"`
	ClientIP         string     `gorm:"type:varchar(45)"`
	CreatedAt        time.Time  `gorm:"autoCreateTime"`

	User User `gorm:"foreignKey:UserID"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (o *Organization) BeforeCreate(_ *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

func (s *SmsCode) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (s *UserSession) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

```

```go
package password

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Hasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(cost int) *BcryptHasher {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{cost: cost}
}

func (b *BcryptHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (b *BcryptHasher) Compare(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrInvalidPassword
	}
	return err
}

var ErrInvalidPassword = errors.New("invalid password")

```

```go
package sms

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
)

type Sender interface {
	Send(ctx context.Context, phone, message string) error
}

type LoggerSender struct {
	log zerolog.Logger
}

func NewLoggerSender(log zerolog.Logger) *LoggerSender {
	return &LoggerSender{log: log}
}

func (s *LoggerSender) Send(_ context.Context, phone, message string) error {
	s.log.Info().
		Str("phone", phone).
		Msgf("SMS sent: %s", message)
	return nil
}

func FormatAuthCode(code string) string {
	return fmt.Sprintf("Ваш код для входа: %s", code)
}

```

```go
package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/nurpe/snowops-auth/internal/model"
)

type Claims struct {
	SessionID uuid.UUID       `json:"sid"`
	UserID    uuid.UUID       `json:"sub"`
	Role      model.UserRole  `json:"role"`
	OrgID     uuid.UUID       `json:"org_id"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret []byte
}

func NewManager(secret string) *Manager {
	return &Manager{secret: []byte(secret)}
}

func (m *Manager) GenerateAccessToken(user *model.User, sessionID uuid.UUID, ttl time.Duration) (string, error) {
	now := time.Now()

	claims := Claims{
		SessionID: sessionID,
		UserID:    user.ID,
		Role:      user.Role,
		OrgID:     user.OrganizationID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) ParseAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

```

```go
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/nurpe/snowops-auth/internal/model"
)

type UserRepository interface {
	FindByLogin(ctx context.Context, login string) (*model.User, error)
	FindByPhone(ctx context.Context, phone string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByLogin(ctx context.Context, login string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("login = ? AND is_active = TRUE", login).Preload("Organization").First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByPhone(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("phone = ? AND is_active = TRUE", phone).Preload("Organization").First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).Preload("Organization").First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

```

```go
package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/nurpe/snowops-auth/internal/model"
)

type SmsCodeRepository interface {
	Create(ctx context.Context, code *model.SmsCode) error
	FindLatest(ctx context.Context, phone string) (*model.SmsCode, error)
	CountActiveInRange(ctx context.Context, phone string, from time.Time) (int64, error)
	MarkUsed(ctx context.Context, id string) error
}

type smsCodeRepository struct {
	db *gorm.DB
}

func NewSmsCodeRepository(db *gorm.DB) SmsCodeRepository {
	return &smsCodeRepository{db: db}
}

func (r *smsCodeRepository) Create(ctx context.Context, code *model.SmsCode) error {
	return r.db.WithContext(ctx).Create(code).Error
}

func (r *smsCodeRepository) FindLatest(ctx context.Context, phone string) (*model.SmsCode, error) {
	var smsCode model.SmsCode
	err := r.db.WithContext(ctx).
		Where("phone = ?", phone).
		Order("created_at DESC").
		First(&smsCode).Error
	if err != nil {
		return nil, err
	}
	return &smsCode, nil
}

func (r *smsCodeRepository) CountActiveInRange(ctx context.Context, phone string, from time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.SmsCode{}).
		Where("phone = ? AND created_at >= ?", phone, from).
		Count(&count).Error
	return count, err
}

func (r *smsCodeRepository) MarkUsed(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&model.SmsCode{}).
		Where("id = ?", id).
		Update("is_used", true).
		Error
}

```

```go
package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/nurpe/snowops-auth/internal/model"
)

type UserSessionRepository interface {
	Create(ctx context.Context, session *model.UserSession) error
	FindByTokenHash(ctx context.Context, hash string) (*model.UserSession, error)
	Revoke(ctx context.Context, id string) error
	RevokeByUser(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context, before time.Time) error
}

type userSessionRepository struct {
	db *gorm.DB
}

func NewUserSessionRepository(db *gorm.DB) UserSessionRepository {
	return &userSessionRepository{db: db}
}

func (r *userSessionRepository) Create(ctx context.Context, session *model.UserSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *userSessionRepository) FindByTokenHash(ctx context.Context, hash string) (*model.UserSession, error) {
	var session model.UserSession
	if err := r.db.WithContext(ctx).
		Where("refresh_token_hash = ? AND revoked_at IS NULL", hash).
		First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *userSessionRepository) Revoke(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.UserSession{}).
		Where("id = ?", id).
		Update("revoked_at", now).
		Error
}

func (r *userSessionRepository) RevokeByUser(ctx context.Context, userID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.UserSession{}).
		Where("user_id = ?", userID).
		Update("revoked_at", now).
		Error
}

func (r *userSessionRepository) DeleteExpired(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).
		Where("expires_at <= ?", before).
		Delete(&model.UserSession{}).
		Error
}

```

```go
package service

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrCodeExpired        = errors.New("sms code expired")
	ErrCodeInvalid        = errors.New("sms code invalid")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionRevoked     = errors.New("session revoked")
)

```

```go
package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/nurpe/snowops-auth/internal/adapter/password"
	"github.com/nurpe/snowops-auth/internal/adapter/sms"
	"github.com/nurpe/snowops-auth/internal/adapter/token"
	"github.com/nurpe/snowops-auth/internal/config"
	"github.com/nurpe/snowops-auth/internal/model"
	"github.com/nurpe/snowops-auth/internal/repository"
)

type AuthService struct {
	users     repository.UserRepository
	sessions  repository.UserSessionRepository
	smsCodes  repository.SmsCodeRepository
	password  password.Hasher
	smsSender sms.Sender
	tokens    *token.Manager
	cfg       *config.Config
}

type AuthMeta struct {
	UserAgent string
	ClientIP  string
}

type AuthResult struct {
	AccessToken  string
	RefreshToken string
	User         UserInfo
}

type UserInfo struct {
	ID             uuid.UUID       `json:"id"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	Organization   string          `json:"organization"`
	Role           model.UserRole  `json:"role"`
	Phone          *string         `json:"phone,omitempty"`
	Login          *string         `json:"login,omitempty"`
}

func NewAuthService(
	users repository.UserRepository,
	sessions repository.UserSessionRepository,
	smsCodes repository.SmsCodeRepository,
	password password.Hasher,
	smsSender sms.Sender,
	tokens *token.Manager,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		users:     users,
		sessions:  sessions,
		smsCodes:  smsCodes,
		password:  password,
		smsSender: smsSender,
		tokens:    tokens,
		cfg:       cfg,
	}
}

func (s *AuthService) Login(ctx context.Context, login, pass string, meta AuthMeta) (*AuthResult, error) {
	user, err := s.users.FindByLogin(ctx, login)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return nil, ErrInvalidCredentials
	}

	if err := s.password.Compare(*user.PasswordHash, pass); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.createSessionAndTokens(ctx, user, meta)
}

func (s *AuthService) SendCode(ctx context.Context, phone string) (string, error) {
	user, err := s.users.FindByPhone(ctx, phone)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", err
	}

	now := time.Now()
	if s.cfg.SMS.DailySendLimit > 0 {
		from := now.Add(-24 * time.Hour)
		count, err := s.smsCodes.CountActiveInRange(ctx, phone, from)
		if err != nil {
			return "", err
		}
		if count >= int64(s.cfg.SMS.DailySendLimit) {
			return "", fmt.Errorf("daily sms limit reached")
		}
	}

	code := generateNumericCode(s.cfg.SMS.CodeLength)
	smsCode := &model.SmsCode{
		Phone:     phone,
		Code:      code,
		ExpiresAt: now.Add(s.cfg.SMS.CodeTTL),
	}

	if err := s.smsCodes.Create(ctx, smsCode); err != nil {
		return "", err
	}

	if err := s.smsSender.Send(ctx, phone, sms.FormatAuthCode(code)); err != nil {
		return "", err
	}

	return maskPhone(phone), nil
}

func (s *AuthService) VerifyCode(ctx context.Context, phone, code string, meta AuthMeta) (*AuthResult, error) {
	smsCode, err := s.smsCodes.FindLatest(ctx, phone)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCodeInvalid
	}
	if err != nil {
		return nil, err
	}

	if smsCode.IsUsed {
		return nil, ErrCodeInvalid
	}
	if smsCode.ExpiresAt.Before(time.Now()) {
		return nil, ErrCodeExpired
	}
	if smsCode.Code != code {
		return nil, ErrCodeInvalid
	}

	if err := s.smsCodes.MarkUsed(ctx, smsCode.ID.String()); err != nil {
		return nil, err
	}

	user, err := s.users.FindByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}

	return s.createSessionAndTokens(ctx, user, meta)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string, meta AuthMeta) (*AuthResult, error) {
	hash := hashToken(refreshToken)

	session, err := s.sessions.FindByTokenHash(ctx, hash)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	if session.RevokedAt != nil {
		return nil, ErrSessionRevoked
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionRevoked
	}

	user, err := s.users.FindByID(ctx, session.UserID.String())
	if err != nil {
		return nil, err
	}

	if err := s.sessions.Revoke(ctx, session.ID.String()); err != nil {
		return nil, err
	}

	return s.createSessionAndTokens(ctx, user, meta)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	hash := hashToken(refreshToken)
	session, err := s.sessions.FindByTokenHash(ctx, hash)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrSessionNotFound
	}
	if err != nil {
		return err
	}
	return s.sessions.Revoke(ctx, session.ID.String())
}

func (s *AuthService) GetMe(ctx context.Context, userID string) (*UserInfo, error) {
	user, err := s.users.FindByID(ctx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	info := toUserInfo(user)
	return &info, nil
}

func (s *AuthService) createSessionAndTokens(ctx context.Context, user *model.User, meta AuthMeta) (*AuthResult, error) {
	sessionID := uuid.New()
	refreshToken := uuid.NewString()
	refreshHash := hashToken(refreshToken)

	session := &model.UserSession{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		ExpiresAt:        time.Now().Add(s.cfg.JWT.RefreshTTL),
		UserAgent:        meta.UserAgent,
		ClientIP:         meta.ClientIP,
	}

	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, err
	}

	accessToken, err := s.tokens.GenerateAccessToken(user, session.ID, s.cfg.JWT.AccessTTL)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toUserInfo(user),
	}, nil
}

func generateNumericCode(length int) string {
	if length <= 0 {
		length = 6
	}
	const digits = "0123456789"
	b := make([]byte, length)
	_, _ = rand.Read(b)
	for i := 0; i < length; i++ {
		b[i] = digits[int(b[i])%len(digits)]
	}
	return string(b)
}

func maskPhone(phone string) string {
	if len(phone) <= 4 {
		return phone
	}
	last4 := phone[len(phone)-4:]
	return fmt.Sprintf("%s***%s", phonePrefix(phone), last4)
}

func phonePrefix(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	if len(phone) <= 7 {
		return phone[:len(phone)-4]
	}
	return phone[:len(phone)-7]
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func toUserInfo(user *model.User) UserInfo {
	var orgName string
	if user.Organization.ID != uuid.Nil {
		orgName = user.Organization.Name
	}
	return UserInfo{
		ID:             user.ID,
		OrganizationID: user.OrganizationID,
		Organization:   orgName,
		Role:           user.Role,
		Phone:          user.Phone,
		Login:          user.Login,
	}
}

```

```go
package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/nurpe/snowops-auth/internal/adapter/password"
	"github.com/nurpe/snowops-auth/internal/adapter/sms"
	"github.com/nurpe/snowops-auth/internal/adapter/token"
	"github.com/nurpe/snowops-auth/internal/config"
	"github.com/nurpe/snowops-auth/internal/model"
	"github.com/nurpe/snowops-auth/internal/repository"
)

type AuthService struct {
	users     repository.UserRepository
	sessions  repository.UserSessionRepository
	smsCodes  repository.SmsCodeRepository
	password  password.Hasher
	smsSender sms.Sender
	tokens    *token.Manager
	cfg       *config.Config
}

type AuthMeta struct {
	UserAgent string
	ClientIP  string
}

type AuthResult struct {
	AccessToken  string
	RefreshToken string
	User         UserInfo
}

type UserInfo struct {
	ID             uuid.UUID       `json:"id"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	Organization   string          `json:"organization"`
	Role           model.UserRole  `json:"role"`
	Phone          *string         `json:"phone,omitempty"`
	Login          *string         `json:"login,omitempty"`
}

func NewAuthService(
	users repository.UserRepository,
	sessions repository.UserSessionRepository,
	smsCodes repository.SmsCodeRepository,
	password password.Hasher,
	smsSender sms.Sender,
	tokens *token.Manager,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		users:     users,
		sessions:  sessions,
		smsCodes:  smsCodes,
		password:  password,
		smsSender: smsSender,
		tokens:    tokens,
		cfg:       cfg,
	}
}

func (s *AuthService) Login(ctx context.Context, login, pass string, meta AuthMeta) (*AuthResult, error) {
	user, err := s.users.FindByLogin(ctx, login)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return nil, ErrInvalidCredentials
	}

	if err := s.password.Compare(*user.PasswordHash, pass); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.createSessionAndTokens(ctx, user, meta)
}

func (s *AuthService) SendCode(ctx context.Context, phone string) (string, error) {
	user, err := s.users.FindByPhone(ctx, phone)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", err
	}

	now := time.Now()
	if s.cfg.SMS.DailySendLimit > 0 {
		from := now.Add(-24 * time.Hour)
		count, err := s.smsCodes.CountActiveInRange(ctx, phone, from)
		if err != nil {
			return "", err
		}
		if count >= int64(s.cfg.SMS.DailySendLimit) {
			return "", fmt.Errorf("daily sms limit reached")
		}
	}

	code := generateNumericCode(s.cfg.SMS.CodeLength)
	smsCode := &model.SmsCode{
		Phone:     phone,
		Code:      code,
		ExpiresAt: now.Add(s.cfg.SMS.CodeTTL),
	}

	if err := s.smsCodes.Create(ctx, smsCode); err != nil {
		return "", err
	}

	if err := s.smsSender.Send(ctx, phone, sms.FormatAuthCode(code)); err != nil {
		return "", err
	}

	return maskPhone(phone), nil
}

func (s *AuthService) VerifyCode(ctx context.Context, phone, code string, meta AuthMeta) (*AuthResult, error) {
	smsCode, err := s.smsCodes.FindLatest(ctx, phone)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCodeInvalid
	}
	if err != nil {
		return nil, err
	}

	if smsCode.IsUsed {
		return nil, ErrCodeInvalid
	}
	if smsCode.ExpiresAt.Before(time.Now()) {
		return nil, ErrCodeExpired
	}
	if smsCode.Code != code {
		return nil, ErrCodeInvalid
	}

	if err := s.smsCodes.MarkUsed(ctx, smsCode.ID.String()); err != nil {
		return nil, err
	}

	user, err := s.users.FindByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}

	return s.createSessionAndTokens(ctx, user, meta)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string, meta AuthMeta) (*AuthResult, error) {
	hash := hashToken(refreshToken)

	session, err := s.sessions.FindByTokenHash(ctx, hash)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	if session.RevokedAt != nil {
		return nil, ErrSessionRevoked
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionRevoked
	}

	user, err := s.users.FindByID(ctx, session.UserID.String())
	if err != nil {
		return nil, err
	}

	if err := s.sessions.Revoke(ctx, session.ID.String()); err != nil {
		return nil, err
	}

	return s.createSessionAndTokens(ctx, user, meta)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	hash := hashToken(refreshToken)
	session, err := s.sessions.FindByTokenHash(ctx, hash)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrSessionNotFound
	}
	if err != nil {
		return err
	}
	return s.sessions.Revoke(ctx, session.ID.String())
}

func (s *AuthService) GetMe(ctx context.Context, userID string) (*UserInfo, error) {
	user, err := s.users.FindByID(ctx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	info := toUserInfo(user)
	return &info, nil
}

func (s *AuthService) createSessionAndTokens(ctx context.Context, user *model.User, meta AuthMeta) (*AuthResult, error) {
	sessionID := uuid.New()
	refreshToken := uuid.NewString()
	refreshHash := hashToken(refreshToken)

	session := &model.UserSession{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		ExpiresAt:        time.Now().Add(s.cfg.JWT.RefreshTTL),
		UserAgent:        meta.UserAgent,
		ClientIP:         meta.ClientIP,
	}

	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, err
	}

	accessToken, err := s.tokens.GenerateAccessToken(user, session.ID, s.cfg.JWT.AccessTTL)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toUserInfo(user),
	}, nil
}

func generateNumericCode(length int) string {
	if length <= 0 {
		length = 6
	}
	const digits = "0123456789"
	b := make([]byte, length)
	_, _ = rand.Read(b)
	for i := 0; i < length; i++ {
		b[i] = digits[int(b[i])%len(digits)]
	}
	return string(b)
}

func maskPhone(phone string) string {
	if len(phone) <= 4 {
		return phone
	}
	last4 := phone[len(phone)-4:]
	return fmt.Sprintf("%s***%s", phonePrefix(phone), last4)
}

func phonePrefix(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	if len(phone) <= 7 {
		return phone[:len(phone)-4]
	}
	return phone[:len(phone)-7]
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func toUserInfo(user *model.User) UserInfo {
	var orgName string
	if user.Organization.ID != uuid.Nil {
		orgName = user.Organization.Name
	}
	return UserInfo{
		ID:             user.ID,
		OrganizationID: user.OrganizationID,
		Organization:   orgName,
		Role:           user.Role,
		Phone:          user.Phone,
		Login:          user.Login,
	}
}


```

```go
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/nurpe/snowops-auth/internal/adapter/token"
	"github.com/nurpe/snowops-auth/internal/service"
)

type Handler struct {
	authService *service.AuthService
	log         zerolog.Logger
}

func NewHandler(authService *service.AuthService, log zerolog.Logger) *Handler {
	return &Handler{
		authService: authService,
		log:         log,
	}
}

func (h *Handler) Register(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	auth := r.Group("/auth")

	auth.POST("/login", h.login)
	auth.POST("/send-code", h.sendCode)
	auth.POST("/verify-code", h.verifyCode)
	auth.POST("/refresh", h.refresh)
	auth.POST("/logout", h.logout)
	auth.GET("/me", authMiddleware, h.me)
}

type loginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type sendCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type verifyCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	result, err := h.authService.Login(
		c.Request.Context(),
		req.Login,
		req.Password,
		metaFromContext(c),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

func (h *Handler) sendCode(c *gin.Context) {
	var req sendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	masked, err := h.authService.SendCode(c.Request.Context(), req.Phone)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"masked_phone": masked,
	})
}

func (h *Handler) verifyCode(c *gin.Context) {
	var req verifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	result, err := h.authService.VerifyCode(
		c.Request.Context(),
		req.Phone,
		req.Code,
		metaFromContext(c),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

func (h *Handler) refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	result, err := h.authService.Refresh(
		c.Request.Context(),
		req.RefreshToken,
		metaFromContext(c),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

func (h *Handler) logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) me(c *gin.Context) {
	claims, ok := c.Get(tokenClaimsKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	tokenClaims := claims.(*token.Claims)
	info, err := h.authService.GetMe(c.Request.Context(), tokenClaims.UserID.String())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(info))
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials),
		errors.Is(err, service.ErrCodeInvalid),
		errors.Is(err, service.ErrCodeExpired):
		c.JSON(http.StatusUnauthorized, errorResponse(err.Error()))
	case errors.Is(err, service.ErrUserNotFound),
		errors.Is(err, service.ErrSessionNotFound):
		c.JSON(http.StatusNotFound, errorResponse(err.Error()))
	default:
		h.log.Error().Err(err).Msg("handler error")
		c.JSON(http.StatusInternalServerError, errorResponse("internal error"))
	}
}

func successResponse(data interface{}) gin.H {
	return gin.H{
		"data": data,
	}
}

func errorResponse(message string) gin.H {
	return gin.H{
		"error": message,
	}
}

```

```go
package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/nurpe/snowops-auth/internal/adapter/token"
	"github.com/nurpe/snowops-auth/internal/service"
)

type Handler struct {
	authService *service.AuthService
	log         zerolog.Logger
}

func NewHandler(authService *service.AuthService, log zerolog.Logger) *Handler {
	return &Handler{
		authService: authService,
		log:         log,
	}
}

func (h *Handler) Register(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	auth := r.Group("/auth")

	auth.POST("/login", h.login)
	auth.POST("/send-code", h.sendCode)
	auth.POST("/verify-code", h.verifyCode)
	auth.POST("/refresh", h.refresh)
	auth.POST("/logout", h.logout)
	auth.GET("/me", authMiddleware, h.me)
}

type loginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type sendCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type verifyCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	result, err := h.authService.Login(
		c.Request.Context(),
		req.Login,
		req.Password,
		metaFromContext(c),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

func (h *Handler) sendCode(c *gin.Context) {
	var req sendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	masked, err := h.authService.SendCode(c.Request.Context(), req.Phone)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"masked_phone": masked,
	})
}

func (h *Handler) verifyCode(c *gin.Context) {
	var req verifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	result, err := h.authService.VerifyCode(
		c.Request.Context(),
		req.Phone,
		req.Code,
		metaFromContext(c),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

func (h *Handler) refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	result, err := h.authService.Refresh(
		c.Request.Context(),
		req.RefreshToken,
		metaFromContext(c),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

func (h *Handler) logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) me(c *gin.Context) {
	claims, ok := c.Get(tokenClaimsKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	tokenClaims := claims.(*token.Claims)
	info, err := h.authService.GetMe(c.Request.Context(), tokenClaims.UserID.String())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(info))
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials),
		errors.Is(err, service.ErrCodeInvalid),
		errors.Is(err, service.ErrCodeExpired):
		c.JSON(http.StatusUnauthorized, errorResponse(err.Error()))
	case errors.Is(err, service.ErrUserNotFound),
		errors.Is(err, service.ErrSessionNotFound):
		c.JSON(http.StatusNotFound, errorResponse(err.Error()))
	default:
		h.log.Error().Err(err).Msg("handler error")
		c.JSON(http.StatusInternalServerError, errorResponse("internal error"))
	}
}

func successResponse(data interface{}) gin.H {
	return gin.H{
		"data": data,
	}
}

func errorResponse(message string) gin.H {
	return gin.H{
		"error": message,
	}
}


```

```go
package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/nurpe/snowops-auth/internal/adapter/token"
	"github.com/nurpe/snowops-auth/internal/service"
)

type Handler struct {
	authService *service.AuthService
	log         zerolog.Logger
}

func NewHandler(authService *service.AuthService, log zerolog.Logger) *Handler {
	return &Handler{
		authService: authService,
		log:         log,
	}
}

func (h *Handler) Register(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	auth := r.Group("/auth")

	auth.POST("/login", h.login)
	auth.POST("/send-code", h.sendCode)
	auth.POST("/verify-code", h.verifyCode)
	auth.POST("/refresh", h.refresh)
	auth.POST("/logout", h.logout)
	auth.GET("/me", authMiddleware, h.me)
}

func metaFromContext(c *gin.Context) service.AuthMeta {
	return service.AuthMeta{
		UserAgent: c.GetHeader("User-Agent"),
		ClientIP:  c.ClientIP(),
	}
}

type loginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type sendCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type verifyCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	result, err := h.authService.Login(
		c.Request.Context(),
		req.Login,
		req.Password,
		metaFromContext(c),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

func (h *Handler) sendCode(c *gin.Context) {
	var req sendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	masked, err := h.authService.SendCode(c.Request.Context(), req.Phone)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"masked_phone": masked,
	})
}

func (h *Handler) verifyCode(c *gin.Context) {
	var req verifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	result, err := h.authService.VerifyCode(
		c.Request.Context(),
		req.Phone,
		req.Code,
		metaFromContext(c),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

func (h *Handler) refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	result, err := h.authService.Refresh(
		c.Request.Context(),
		req.RefreshToken,
		metaFromContext(c),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

func (h *Handler) logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) me(c *gin.Context) {
	claims, ok := c.Get(tokenClaimsKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	tokenClaims := claims.(*token.Claims)
	info, err := h.authService.GetMe(c.Request.Context(), tokenClaims.UserID.String())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(info))
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials),
		errors.Is(err, service.ErrCodeInvalid),
		errors.Is(err, service.ErrCodeExpired):
		c.JSON(http.StatusUnauthorized, errorResponse(err.Error()))
	case errors.Is(err, service.ErrUserNotFound),
		errors.Is(err, service.ErrSessionNotFound):
		c.JSON(http.StatusNotFound, errorResponse(err.Error()))
	default:
		h.log.Error().Err(err).Msg("handler error")
		c.JSON(http.StatusInternalServerError, errorResponse("internal error"))
	}
}

func successResponse(data interface{}) gin.H {
	return gin.H{
		"data": data,
	}
}

func errorResponse(message string) gin.H {
	return gin.H{
		"error": message,
	}
}


```

```go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nurpe/snowops-auth/internal/adapter/token"
)

const tokenClaimsKey = "tokenClaims"

func Auth(tokenManager *token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header missing"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		claims, err := tokenManager.ParseAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(tokenClaimsKey, claims)
		c.Next()
	}
}

```

```go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nurpe/snowops-auth/internal/adapter/token"
)

const TokenClaimsKey = "tokenClaims"

func Auth(tokenManager *token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header missing"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		claims, err := tokenManager.ParseAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(TokenClaimsKey, claims)
		c.Next()
	}
}


```

```go
package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/nurpe/snowops-auth/internal/adapter/token"
	"github.com/nurpe/snowops-auth/internal/http/middleware"
	"github.com/nurpe/snowops-auth/internal/service"
)

type Handler struct {
	authService *service.AuthService
	log         zerolog.Logger
}

func NewHandler(authService *service.AuthService, log zerolog.Logger) *Handler {
	return &Handler{
		authService: authService,
		log:         log,
	}
}

func (h *Handler) Register(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	auth := r.Group("/auth")

	auth.POST("/login", h.login)
	auth.POST("/send-code", h.sendCode)
	auth.POST("/verify-code", h.verifyCode)
	auth.POST("/refresh", h.refresh)
	auth.POST("/logout", h.logout)
	auth.GET("/me", authMiddleware, h.me)
}

func metaFromContext(c *gin.Context) service.AuthMeta {
	return service.AuthMeta{
		UserAgent: c.GetHeader("User-Agent"),
		ClientIP:  c.ClientIP(),
	}
}

type loginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type sendCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type verifyCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	result, err := h.authService.Login(
		c.Request.Context(),
		req.Login,
		req.Password,
		metaFromContext(c),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

func (h *Handler) sendCode(c *gin.Context) {
	var req sendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	masked, err := h.authService.SendCode(c.Request.Context(), req.Phone)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"masked_phone": masked,
	})
}

func (h *Handler) verifyCode(c *gin.Context) {
	var req verifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	result, err := h.authService.VerifyCode(
		c.Request.Context(),
		req.Phone,
		req.Code,
		metaFromContext(c),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

func (h *Handler) refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	result, err := h.authService.Refresh(
		c.Request.Context(),
		req.RefreshToken,
		metaFromContext(c),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

func (h *Handler) logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) me(c *gin.Context) {
	claims, ok := c.Get(middleware.TokenClaimsKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	tokenClaims := claims.(*token.Claims)
	info, err := h.authService.GetMe(c.Request.Context(), tokenClaims.UserID.String())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(info))
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials),
		errors.Is(err, service.ErrCodeInvalid),
		errors.Is(err, service.ErrCodeExpired):
		c.JSON(http.StatusUnauthorized, errorResponse(err.Error()))
	case errors.Is(err, service.ErrUserNotFound),
		errors.Is(err, service.ErrSessionNotFound):
		c.JSON(http.StatusNotFound, errorResponse(err.Error()))
	default:
		h.log.Error().Err(err).Msg("handler error")
		c.JSON(http.StatusInternalServerError, errorResponse("internal error"))
	}
}

func successResponse(data interface{}) gin.H {
	return gin.H{
		"data": data,
	}
}

func errorResponse(message string) gin.H {
	return gin.H{
		"error": message,
	}
}


```

```go
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(handler *Handler, authMiddleware gin.HandlerFunc, env string) *gin.Engine {
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	handler.Register(router, authMiddleware)

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

```

```go
package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/nurpe/snowops-auth/internal/adapter/password"
	"github.com/nurpe/snowops-auth/internal/adapter/sms"
	"github.com/nurpe/snowops-auth/internal/adapter/token"
	"github.com/nurpe/snowops-auth/internal/config"
	"github.com/nurpe/snowops-auth/internal/db"
	httphandler "github.com/nurpe/snowops-auth/internal/http"
	"github.com/nurpe/snowops-auth/internal/http/middleware"
	"github.com/nurpe/snowops-auth/internal/logger"
	"github.com/nurpe/snowops-auth/internal/repository"
	"github.com/nurpe/snowops-auth/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	appLogger := logger.New(cfg.Environment)

	database, err := db.New(cfg.DB, appLogger, cfg.Environment)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("failed to connect database")
	}

	userRepo := repository.NewUserRepository(database)
	smsRepo := repository.NewSmsCodeRepository(database)
	sessionRepo := repository.NewUserSessionRepository(database)

	passwordHasher := password.NewBcryptHasher(0)
	tokenManager := token.NewManager(cfg.JWT.AccessSecret)
	smsSender := sms.NewLoggerSender(appLogger)

	authService := service.NewAuthService(
		userRepo,
		sessionRepo,
		smsRepo,
		passwordHasher,
		smsSender,
		tokenManager,
		cfg,
	)

	handler := httphandler.NewHandler(authService, appLogger)
	authMiddleware := middleware.Auth(tokenManager)
	router := httphandler.NewRouter(handler, authMiddleware, cfg.Environment)

	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	appLogger.Info().Msgf("starting auth service on %s", addr)

	if err := router.Run(addr); err != nil {
		appLogger.Error().Err(err).Msg("failed to start server")
		os.Exit(1)
	}
}

```

```go
package main

import (
	"fmt"
	"os"

	"github.com/nurpe/snowops-auth/internal/adapter/password"
	"github.com/nurpe/snowops-auth/internal/adapter/sms"
	"github.com/nurpe/snowops-auth/internal/adapter/token"
	"github.com/nurpe/snowops-auth/internal/config"
	"github.com/nurpe/snowops-auth/internal/db"
	httphandler "github.com/nurpe/snowops-auth/internal/http"
	"github.com/nurpe/snowops-auth/internal/http/middleware"
	"github.com/nurpe/snowops-auth/internal/logger"
	"github.com/nurpe/snowops-auth/internal/repository"
	"github.com/nurpe/snowops-auth/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	appLogger := logger.New(cfg.Environment)

	database, err := db.New(cfg.DB, appLogger, cfg.Environment)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("failed to connect database")
	}

	userRepo := repository.NewUserRepository(database)
	smsRepo := repository.NewSmsCodeRepository(database)
	sessionRepo := repository.NewUserSessionRepository(database)

	passwordHasher := password.NewBcryptHasher(0)
	tokenManager := token.NewManager(cfg.JWT.AccessSecret)
	smsSender := sms.NewLoggerSender(appLogger)

	authService := service.NewAuthService(
		userRepo,
		sessionRepo,
		smsRepo,
		passwordHasher,
		smsSender,
		tokenManager,
		cfg,
	)

	handler := httphandler.NewHandler(authService, appLogger)
	authMiddleware := middleware.Auth(tokenManager)
	router := httphandler.NewRouter(handler, authMiddleware, cfg.Environment)

	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	appLogger.Info().Msgf("starting auth service on %s", addr)

	if err := router.Run(addr); err != nil {
		appLogger.Error().Err(err).Msg("failed to start server")
		os.Exit(1)
	}
}


```

```go
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm"

	"github.com/nurpe/snowops-auth/internal/config"
	"github.com/nurpe/snowops-auth/internal/model"
)

func New(cfg config.DBConfig, log zerolog.Logger, appEnv string) (*gorm.DB, error) {
	gormLog := gormlogger.New(
		zerologAdapter{logger: log},
		gormlogger.Config{
			SlowThreshold:             time.Second,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			LogLevel:                  selectLogLevel(appEnv),
		},
	)

	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		Logger: gormLog,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		return nil, err
	}
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`).Error; err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&model.Organization{},
		&model.User{},
		&model.SmsCode{},
		&model.UserSession{},
	); err != nil {
		return nil, err
	}

	return db, nil
}

func HealthCheck(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec("SELECT 1").Error
}

func selectLogLevel(env string) gormlogger.LogLevel {
	if env == "development" {
		return gormlogger.Info
	}
	return gormlogger.Warn
}

type zerologAdapter struct {
	logger zerolog.Logger
}

func (z zerologAdapter) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return z
}

func (z zerologAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	z.logger.Info().Msg(fmt.Sprintf(msg, data...))
}

func (z zerologAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	z.logger.Warn().Msg(fmt.Sprintf(msg, data...))
}

func (z zerologAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	z.logger.Error().Msg(fmt.Sprintf(msg, data...))
}

func (z zerologAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	event := z.logger.Debug()
	if err != nil {
		event = z.logger.Error().Err(err)
	}
	event.
		Dur("elapsed", elapsed).
		Int64("rows", rows).
		Msg(sql)
}


```

```go
package sms

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
)

type Sender interface {
	Send(ctx context.Context, phone, message string) error
}

type LoggerSender struct {
	log zerolog.Logger
}

func NewLoggerSender(log zerolog.Logger) *LoggerSender {
	return &LoggerSender{log: log}
}

func (s *LoggerSender) Send(_ context.Context, phone, message string) error {
	s.log.Info().
		Str("phone", phone).
		Str("message", message).
		Msg("sms sent")
	return nil
}

func FormatAuthCode(code string) string {
	return fmt.Sprintf("Ваш код для входа: %s", code)
}


```

```go
package main

import (
	"fmt"
	"os"

	"github.com/nurpe/snowops-auth/internal/adapter/password"
	"github.com/nurpe/snowops-auth/internal/adapter/sms"
	"github.com/nurpe/snowops-auth/internal/adapter/token"
	"github.com/nurpe/snowops-auth/internal/config"
	"github.com/nurpe/snowops-auth/internal/db"
	httphandler "github.com/nurpe/snowops-auth/internal/http"
	"github.com/nurpe/snowops-auth/internal/http/middleware"
	"github.com/nurpe/snowops-auth/internal/logger"
	"github.com/nurpe/snowops-auth/internal/repository"
	"github.com/nurpe/snowops-auth/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	appLogger := logger.New(cfg.Environment)

	database, err := db.New(cfg.DB, appLogger, cfg.Environment)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("failed to connect database")
	}

	userRepo := repository.NewUserRepository(database)
	smsRepo := repository.NewSmsCodeRepository(database)
	sessionRepo := repository.NewUserSessionRepository(database)

	passwordHasher := password.NewBcryptHasher(0)
	tokenManager := token.NewManager(cfg.JWT.AccessSecret)
	smsSender := sms.NewLoggerSender(appLogger)

	authService := service.NewAuthService(
		userRepo,
		sessionRepo,
		smsRepo,
		passwordHasher,
		smsSender,
		tokenManager,
		cfg,
	)

	handler := httphandler.NewHandler(authService, appLogger)
	authMiddleware := middleware.Auth(tokenManager)
	router := httphandler.NewRouter(handler, authMiddleware, cfg.Environment)

	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	appLogger.Info().Str("addr", addr).Msg("starting auth service")

	if err := router.Run(addr); err != nil {
		appLogger.Error().Err(err).Msg("failed to start server")
		os.Exit(1)
	}
}


```

