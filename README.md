# PR Reviewer Checker

Сервис реализует HTTP API сервиса по управлению pull requests из `openapi.yml`: управление командами/пользователями и автоматическое назначение ревьюеров на Pull Request. Реализация на Go, хранение в PostgreSQL, миграции через `migrate`.

## Запуск

```bash
# основной docker compose (приложение + БД + миграции)
make up
# остановить
make down
```

Локальные переменные и подключения задаются в `docker-compose.yml`. Сервис по умолчанию слушает `:8080`.

## Тестирование

```bash
make unit-test        # go generate + go test
make integration-test # docker-compose-test: БД + migrate + ./tests
```

Интеграционные тесты используют `docker-compose-test.yml`: поднимают временную БД, накатывают миграции и прогоняют сценарии из `./tests`. Перед запуском убедитесь, что `docker` и `docker compose` доступны.

### Нагрузочное тестирование

Для быстрого локального теста можно использовать `scripts/load_test.sh` (требует установленного `hey`). По умолчанию грузит `POST /team/add` 5 RPS в течение 30 секунд:

```bash
./scripts/load_test.sh               # по умолчанию http://localhost:8080/team/add
./scripts/load_test.sh http://...    # свой адрес
```

## Миграции

SQL миграции лежат в папке `migrations/`. `docker-compose.yml` автоматически выполняет `000001_init`, `000002_add_idx`, `000003_init_test_data`.

## API

Полная спецификация в `openapi.yml`. Примеры запросов:

```bash
curl -X POST http://localhost:8080/team/add -H 'Content-Type: application/json' \
  -d '{"team_name":"backend","members":[{"user_id":"u1","username":"Alice","is_active":true}]}'
```

Для проверки готовности сервиса используйте `GET /health`.
