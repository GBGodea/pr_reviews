# PR Reviewer Assignment Service

REST API микросервис для управления автоматическим назначением и перераспределением ревьюверов на Pull Request'ы. Разработан на Go с использованием PostgreSQL и Docker Compose.

## Описание

Сервис обеспечивает автоматизацию процесса назначения ревьюверов на код-ревью в команде разработки. Основные функции:

- Управление командами и участниками
- Автоматическое назначение ревьюверов при создании PR
- Управление статусом активности участников
- Переназначение ревьюверов с учетом ограничений
- Блокировка изменений ревьюверов после merge

## Установка и запуск

### Через Docker Compose

```bash
docker-compose up
```

Сервис будет доступен по адресу `http://localhost:8080`

### Локальный запуск

```bash
go mod download
go run ./cmd/main.go
```

Требуется запущенный PostgreSQL на localhost:5432 с параметрами:
- User: user
- Password: password
- Database: pr_service

---

## API Endpoints

### Health Check

```
GET /health
```

---

## Управление командами

### Создание команды

```
POST /team/add
```

**Заголовки:**
```
Content-Type: application/json
```

**Тело запроса:**
```json
{
  "team_name": "backend",
  "members": [
    {
      "user_id": "u1",
      "username": "Alice",
      "is_active": true
    },
    {
      "user_id": "u2",
      "username": "Bob",
      "is_active": true
    },
    {
      "user_id": "u3",
      "username": "Charlie",
      "is_active": true
    }
  ]
}
```

**curl запрос:**
```bash
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{"team_name": "backend", "members": [{"user_id": "u1", "username": "Alice", "is_active": true}, {"user_id": "u2", "username": "Bob", "is_active": true}, {"user_id": "u3", "username": "Charlie", "is_active": true}]}'
```

---

### Получение информации о команде

```
GET /team/get?team_name=<team_name>
```

**Параметры:**
- `team_name` (string, required) — название команды

**curl запрос:**
```bash
curl "http://localhost:8080/team/get?team_name=backend"
```

---

## Управление пользователями

### Установка статуса активности

```
POST /users/setIsActive
```

**Заголовки:**
```
Content-Type: application/json
```

**Тело запроса:**
```json
{
  "user_id": "u1",
  "is_active": false
}
```

**curl запрос:**
```bash
curl -X POST http://localhost:8080/users/setIsActive \
  -H "Content-Type: application/json" \
  -d '{"user_id": "u1", "is_active": false}'
```

---

### Получение PR'ов для ревью

```
GET /users/getReview?user_id=<user_id>
```

**Параметры:**
- `user_id` (string, required) — идентификатор пользователя

**curl запрос:**
```bash
curl "http://localhost:8080/users/getReview?user_id=u1"
```

---

## Управление Pull Request'ами

### Создание PR

```
POST /pullRequest/create
```

**Заголовки:**
```
Content-Type: application/json
```

**Тело запроса:**
```json
{
  "pull_request_id": "pr-001",
  "pull_request_name": "Add search feature",
  "author_id": "u1"
}
```

**Алгоритм назначения ревьюверов:**
- Выбираются до 2 активных членов команды автора
- Автор исключается из выбора
- Выбор выполняется случайным образом

**curl запрос:**
```bash
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{"pull_request_id": "pr-001", "pull_request_name": "Add feature", "author_id": "u1"}'
```

---

### Merge PR

```
POST /pullRequest/merge
```

**Заголовки:**
```
Content-Type: application/json
```

**Тело запроса:**
```json
{
  "pull_request_id": "pr-001"
}
```

**Примечание:** Операция является идемпотентной. После выполнения изменение ревьюверов становится невозможным.

**curl запрос:**
```bash
curl -X POST http://localhost:8080/pullRequest/merge \
  -H "Content-Type: application/json" \
  -d '{"pull_request_id": "pr-001"}'
```

---

### Переназначение ревьювера

```
POST /pullRequest/reassign
```

**Заголовки:**
```
Content-Type: application/json
```

**Тело запроса:**
```json
{
  "pull_request_id": "pr-001",
  "old_user_id": "u2"
}
```

**Ограничения:**
- PR должен находиться в статусе OPEN
- Указанный ревьювер должен быть назначен на данный PR
- Новый ревьювер выбирается случайно из активных членов команды
- Новый ревьювер не должен быть уже назначен на данный PR

**curl запрос:**
```bash
curl -X POST http://localhost:8080/pullRequest/reassign \
  -H "Content-Type: application/json" \
  -d '{"pull_request_id": "pr-001", "old_user_id": "u2"}'
```
