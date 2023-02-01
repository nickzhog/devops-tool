# «Сервис сбора метрик и алертинга»
Я.Практикум: Курс — Golang advanced

Реализовано:
-Агент по сбору рантайм-метрик, который отправляет метрики по протоколу HTTP.
-Сервер по сбору рантайм-метрик, который собирает репорты от агентов по протоколу HTTP, отображает их различными способами

### Настройки приложений
Приложения можно настраивать через переменные окружения или через аргументы, флаги. Приоритет для значений полученных через переменные окружения
Агент:
- `ADDRESS` или `-a`: адрес на который будут отправляться метрики (по умолчанию `http://127.0.0.1:8080`)
- `POLL_INTERVAL` или `-p`: интервал обновления метрик (по умолчанию `2s`)
- `REPORT_INTERVAL` или `-r`: интервал отправки метрик (по умолчанию `10s`)
- `KEY` или `-k`: если указан ключ, то каждая метрика будет хэшироваться с его использованием

Сервер:
- `ADDRESS` или `-a`: адрес на котором запускается сервер  (по умолчанию `:8080`)
- `DATABASE_DSN` или `-d`: если указан, то сервер будет использовать базу данных (по умолчанию используется кэширование)
- `STORE_FILE` или `-f`: указывается путь файла, если указан, то сервер будет сохранять все метрики в JSON формате в этот файл(по умолчанию `/tmp/devops-metrics-db.json`)
- `STORE_INTERVAL` или `-i`: интервал обновления метрик в файле (по умолчанию `1s`)
- `RESTORE` или `-r`: восстановление метрик из файла при запуске программы (по умолчанию `true`)
- `KEY` или `-k`: если указан ключ, то перед обновлением каждой метрики будет проверена хэш-сумма, с использованием ключа
