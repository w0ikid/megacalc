# Распределенный калькулятор арифметических выражений

## Описание
Данный проект представляет собой распределенный калькулятор, который выполняет обработку арифметических выражений. Архитектура включает:
- **Оркестратор** — принимает выражения, разбивает их на задачи и управляет вычислениями.
- **Агенты** — выполняют математические операции параллельно.
- **Веб-интерфейс** — визуализирует процесс вычислений

## Возможности
- Поддержка арифметических операций: `+`, `-`, `*`, `/`
- Приоритет операций и работа со скобками
- Поддержка больших чисел и точных вычислений
- Проверка валидности выражений
- Параллельные вычисления с распределением задач
- Логирование выполнения задач

## Предварительные требования
- **Go 1.24.0**
- **Docker**
- **Git**

## Установка
```sh
git clone git@github.com:w0ikid/megacalc.git
cd megacalc
go mod tidy  # Загрузка и синхронизация зависимостей
```

## Запуск
### Запуск с Docker
```sh
 docker-compose up --build -d  
```

### Web (в своем браузере)
```
localhost
```
or

```
http://localhost
```

### Важно

- Фронтенд (web/) обрабатывается Nginx, который монтируется в контейнере.
- Если веб-интерфейс не отображает данные, попробуйте:

```sh
curl -X GET "http://localhost:8080/api/v1/expressions"
```

## API эндпоинты
### Отправка выражения на вычисление
```sh
curl -X POST "http://localhost:8080/api/v1/calculate" \
     -H "Content-Type: application/json" \
     -d '{"expression":"2+2*2"}'
```
**Ответ:**
```json
{"id":"123e4567-e89b-12d3-a456-426614174000"}
```

### Получение списка всех выражений
```sh
curl -X GET "http://localhost:8080/api/v1/expressions"
```
**Ответ:**
```json
{
  "expressions": [
    {"id": "123e4567-e89b-12d3-a456-426614174000", "status": "completed", "result": 6}
  ]
}
```

### Получение конкретного выражения по ID
```sh
curl -X GET "http://localhost:8080/api/v1/expressions/123e4567-e89b-12d3-a456-426614174000"
```
### 1. **Неуспешное вычисление — Пустое выражение:**
```sh
curl -L 'http://localhost:8080/api/v1/calculate' -H 'Content-Type: application/json' --data '{"expression":""}'
```
Ответ (HTTP 400 Bad Request):
```sh
{
    "error": "Key: 'ExpressionRequest.Expression' Error:Field validation for 'Expression' failed on the 'required' tag"
}
```
### 2. Неуспешное вычисление — Неверный формат выражения:
```sh
curl -L 'http://localhost:8080/api/v1/calculate' -H 'Content-Type: application/json' --data '{"expression":"2++2"}'
```
Ответ (HTTP 400 Bad Request):
```sh
{
    "error": "invalid expression: not enough operands for operator +"
}
```
### 3. Выражение не найдено:
```sh
curl --location 'http://localhost:8080/api/v1/expressions/non-existent-id'
```
Ответ (HTTP 404 Not Found):
```sh
{
    "error": "expression not found"
}
```
4. Список всех выражений:
```sh
curl --location 'http://localhost:8080/api/v1/expressions'
```
Ответ (HTTP 200 OK):
```sh
{
    "expressions": [
        {"id": "ba0e2320-19e0-4595-90d8-7ce24d1a618f", "status": "failed"},
        {"id": "5bd93f00-8fdd-43a4-87be-844193212778", "status": "completed", "result": 10},
        {"id": "6fad18b7-3a22-460e-be79-a2ba2de37e3c", "status": "completed", "result": 2},
        {"id": "89fe8f3f-92c3-4abe-ab39-3a069697a01b", "status": "completed", "result": 10},
        {"id": "c07a3f9b-0f57-4a4e-b568-5b9e333d2b49", "status": "completed", "result": 10},
        {"id": "fe751363-69a3-4eb2-aa28-e45870bb0ab8", "status": "completed", "result": 10},
        {"id": "0f720b1a-4880-4139-b06f-a95d1059bbc0", "status": "completed", "result": 16},
        {"id": "537cdd8a-f5eb-4d99-902c-65ff6598ab2e", "status": "completed", "result": 7.2},
        {"id": "e37c7b28-d6c4-4d5d-857c-07ad2a217202", "status": "completed", "result": 10},
        {"id": "3602eaf7-a141-488e-a7d0-5e958fafe3fd", "status": "completed", "result": 4}
    ]
}
```
## Тестирование
Запуск всех тестов
```sh
go test ./...
```

## Структура проекта
```plaintext
.
├── cmd
│   ├── agent
│   │   └── agent.go              # Агент, выполняющий вычисления
│   └── orchestrator
│       └── main.go               # Оркестратор, управляющий вычислениями
├── docker-compose.yml            # Конфигурация для Docker
├── Dockerfile.agent              # Dockerfile для агента
├── Dockerfile.orchestrator       # Dockerfile для оркестратора
├── go.mod                        # Зависимости Go
├── go.sum                        # Контрольные суммы зависимостей
├── internal
│   ├── agent
│   │   └── agent.go              # Логика агента
│   ├── api
│   │   ├── handler.go            # API обработчик
│   │   └── handler_test.go       # Тесты для обработчика API
│   └── service
│       ├── service.go            # Сервис, выполняющий обработку выражений
│       └── service_test.go       # Тесты для сервиса
├── nginx.conf                    # Конфигурация для Nginx
├── orchestrator                  # Папка с кодом оркестратора
├── pkg
│   └── errors
│       └── errors.go             # Обработка ошибок
├── readme.md                     # Описание проекта
└── web
    ├── index.html                # HTML страница для веб-интерфейса
    └── static
        ├── js
        │   └── main.js           # JavaScript для работы с веб-интерфейсом
        └── style.css             # Стили для веб-интерфейса
```

# Эта структура включает все ключевые компоненты, такие как:
```
cmd — основное место для запуска различных частей приложения (агентов и оркестратора).
docker-compose.yml — конфигурация для запуска проекта с помощью Docker.
internal — внутренние пакеты, реализующие логику вычислений и API.
web — фронтенд для визуализации результатов вычислений.
```
