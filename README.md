# PR Reviewer Checker

Сервис реализует HTTP API сервиса по управлению pull requests из `openapi.yml`: управление командами/пользователями и автоматическое назначение ревьюеров на Pull Request. Реализация на Go, хранение в PostgreSQL, миграции через `migrate`.

## Запуск

```bash
# основной docker compose (приложение + БД + миграции)
make up
# остановить
make down
```

Локальные переменные и подключения задаются в `docker-compose.yml`. и `.env` А параметры сервера и логгера задаются в файлах из папки `configs/`

## Тестирование

```bash
make unit-test        # go generate + go test
```

## Нагрузочное тестирование

Для быстрого локального теста можно использовать `scripts/load_test.sh` (требует установленного `apache-utils`). По умолчанию грузит `POST /team/add` 3000 запросов из 10 потоков:

```bash
./scripts/load_test.sh               # по умолчанию http://localhost:8080/team/add
./scripts/load_test.sh http://...    # свой адрес
```

### Результаты

Московский сервер `4CPU`, `8 RAM`, Нагрузочное тестирование проводилось на машине в СПБ. Результаты: 70/300 ms на запрос. RPS: 140, Перцентиль успешности на 3000 запросов 100%
```bash
Server Software:
  Server Hostname:        31.130.149.163
  Server Port:            8080

  Document Path:          /team/add
  Document Length:        248 bytes

  Concurrency Level:      10
  Time taken for tests:   21.294 seconds
  Complete requests:      3000
  Failed requests:        0
  Total transferred:      1086000 bytes
  Total body sent:        1260000
  HTML transferred:       744000 bytes
  Requests per second:    140.88 [#/sec] (mean)
  Time per request:       70.981 [ms] (mean)
  Time per request:       7.098 [ms] (mean, across all concurrent requests)
  Transfer rate:          49.80 [Kbytes/sec] received
                          57.78 kb/s sent
                          107.59 kb/s total

  Connection Times (ms)
                min  mean[+/-sd] median   max
  Connect:       10   55 424.2     12    7225
  Processing:    11   16   5.9     15     129
  Waiting:       11   15   5.2     15     128
  Total:         22   71 424.5     27    7245

  Percentage of the requests served within a certain time (ms)
    50%     27
    66%     30
    75%     31
    80%     31
    90%     34
    95%     38
    98%   1031
    99%   1048
   100%   7245 (longest request)
```

Замеры с включённым VPN(Германия), 8000 запросов Всё равно удовлетворяет условию
```bash
Server Software:        
Server Hostname:        31.130.149.163
Server Port:            8080

Document Path:          /team/add
Document Length:        248 bytes

Concurrency Level:      10
Time taken for tests:   147.929 seconds
Complete requests:      8000
Failed requests:        0
Total transferred:      2896000 bytes
Total body sent:        3360000
HTML transferred:       1984000 bytes
Requests per second:    54.08 [#/sec] (mean)
Time per request:       184.911 [ms] (mean)
Time per request:       18.491 [ms] (mean, across all concurrent requests)
Transfer rate:          19.12 [Kbytes/sec] received
                        22.18 kb/s sent
                        41.30 kb/s total

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:       81   91  56.3     83    1230
Processing:    83   94  39.7     86     954
Waiting:       83   93  37.2     86     954
Total:        165  184  73.3    169    1387

Percentage of the requests served within a certain time (ms)
  50%    169
  66%    171
  75%    172
  80%    174
  90%    186
  95%    287
  98%    313
  99%    422
 100%   1387 (longest request)
```
## Миграции

SQL миграции лежат в папке `migrations/`. `docker-compose.yml` автоматически выполняет `000001_init`, `000002_add_idx`

## API

Полная спецификация в `openapi.yml`. Примеры запросов:

```bash
curl -X POST http://localhost:8080/team/add -H 'Content-Type: application/json' \
  -d '{"team_name":"backend","members":[{"user_id":"u1","username":"Alice","is_active":true}]}'
```
