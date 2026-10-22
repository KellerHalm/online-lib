# Электронная библиотека (вариант 1)

Backend-приложение на Go (Gin + GORM) с PostgreSQL, JWT-аутентификацией и CORS.

## Запуск

```bash
docker compose up --build
```

API доступен на http://localhost:3000.
Swagger UI: http://localhost:3000/swagger (в Docker требуется интернет для CDN Swagger UI).

## Переменные окружения

`.env` не обязателен: в `docker-compose.yml` у всех переменных есть значения по умолчанию, поэтому проект запускается сразу после клонирования. Чтобы переопределить значения (например, `JWT_SECRET`), скопируйте `.env.example` в `.env` — Docker Compose подхватит его автоматически.

| Переменная | По умолчанию | Назначение |
|---|---|---|
| `PORT` | `3000` | Порт API |
| `DB_HOST` | `db` | Хост PostgreSQL (для запуска без Docker — `localhost`) |
| `DB_PORT` | `5432` | Порт PostgreSQL |
| `DB_USER` | `lib` | Пользователь БД |
| `DB_PASSWORD` | `lib_password` | Пароль БД |
| `DB_NAME` | `library` | Имя БД |
| `JWT_SECRET` | `dev_secret_change_me` | Секрет подписи JWT (обязательно задать свой в реальном окружении) |
| `CORS_ALLOWED_ORIGIN` | `http://localhost:5173` | Единственный разрешённый origin |

## Структура проекта

```
online-lib/
├── cmd/
│   └── server/
│       └── main.go              точка входа: конфиг, БД, роутинг, запуск Gin
├── internal/
│   ├── config/
│   │   └── config.go            чтение переменных окружения
│   ├── database/
│   │   └── database.go          подключение к PostgreSQL и AutoMigrate
│   ├── models/
│   │   ├── book.go              сущность Book
│   │   └── user.go              сущность User
│   ├── handlers/
│   │   ├── auth.go              регистрация, вход, /auth/me
│   │   └── books.go             CRUD книг, валидация, фильтр ?author=
│   └── middleware/
│       ├── auth.go              проверка JWT (Bearer)
│       └── cors.go              CORS только для заданного origin
├── docs/
│   ├── index.html               Swagger UI
│   └── openapi.json             OpenAPI-спецификация
├── docker-compose.yml           приложение + PostgreSQL, том db_data
├── Dockerfile                   многоэтапная сборка
├── .env.example                 пример переменных окружения
├── go.mod / go.sum              зависимости
└── README.md
```

## Эндпоинты

| Метод | Путь | Доступ | Описание |
|---|---|---|---|
| POST | /auth/register | открытый | Регистрация (email, пароль не короче 6 символов). 201/400/409 |
| POST | /auth/login | открытый | Вход, ответ `{ "accessToken": "..." }`. 200/401 |
| GET | /auth/me | JWT | Данные текущего пользователя. 200/401 |
| GET | /books | открытый | Список книг, фильтр `?author=` (точное совпадение). 200 |
| GET | /books/:id | открытый | Одна книга. 200/404 |
| POST | /books | JWT | Создать книгу. 201/400/401/409 |
| PATCH | /books/:id | JWT | Частично изменить книгу. 200/400/401/404/409 |
| DELETE | /books/:id | JWT | Удалить книгу. 204/401/404 |

Книга: `title` (1–200), `author` (1–100), `year` (1450–текущий год), `isbn` (ровно 13 цифр, уникальный — дубль даёт 409), `isAvailable` (по умолчанию true).

CORS разрешён только для origin `http://localhost:5173` (заголовки `Authorization`, `Content-Type`; методы GET, POST, PATCH, DELETE).

Токен передаётся в заголовке `Authorization: Bearer <token>`, время жизни — 1 час.
