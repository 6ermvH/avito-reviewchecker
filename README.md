# PR Reviewe Checker

Сервис реализует HTTP API сервиса по управлению pull requests из `openapi.yml`: управление командами/пользователями и автоматическое назначение ревьюеров на Pull Request. Реализация на Go, хранение в PostgreSQL, миграции через `migrate`.

## Запуск

```bash
# основной docker compose (приложение + БД + миграции)
make up
# остановить
make down
```

Локальные переменные и подключения задаются в `docker-compose.yml`. Сервис по умолчанию слушает `:8080`.

## Запуск(unit тестов)
```bash
make unit-test
```

## Запуск интеграционных тестов
```bash
make integration-test
```

Интеграционные тесты используют `docker-compose-test.yml`: поднимают временную БД, накатывают миграции и прогоняют сценарии из `./tests`. Перед запуском убедитесь, что `docker` и `docker compose` доступны.

## Миграции

SQL миграции лежат в папке `migrations/`. `docker-compose.yml` автоматически выполняет `000001_init`, `000002_add_idx`, `000003_init_test_data`.

## API

Полная спецификация в `openapi.yml`. Примеры запросов:

```bash
curl -X POST http://localhost:8080/team/add -H 'Content-Type: application/json' \
  -d '{"team_name":"backend","members":[{"user_id":"u1","username":"Alice","is_active":true}]}'
```
