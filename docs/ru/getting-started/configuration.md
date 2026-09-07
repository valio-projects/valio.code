# Конфигурация

Compose-файл уже содержит значения для локальной разработки: `.env`, `setup.ps1`
и `setup.sh` не нужны. Выполните из корня репозитория:

```powershell
docker compose up -d --build --wait
```

У SurrealDB одна локальная root-учётная запись: пользователь `root`, пароль
`valio`, аутентификация уровня root. API, worker и миграция используют эту же
root-аутентификацию. База данных остаётся во внутренней сети Compose; приложение
доступно на `127.0.0.1:8080`, Jaeger — на `127.0.0.1:16686`.

Локальный API-токен задан в Compose:
`valio-local-development-token-0001`; CLI использует тот же токен по умолчанию.
После входа браузер хранит только HttpOnly cookie.

При пересоздании сервисов именованный volume сохраняется. В новом checkout
миграция учётных данных не нужна. Перед пересозданием старого стека с существующим
volume смените root-пароль, пока старый контейнер `surrealdb` ещё запущен:

```powershell
"DEFINE USER OVERWRITE root ON ROOT PASSWORD 'valio' ROLES OWNER;" | docker compose exec -T surrealdb /surreal sql --endpoint http://localhost:8000 --username root --auth-level root --hide-welcome --json
```

Команда сохраняет данные. Если старый контейнер уже недоступен, выполните ту же
смену с известным старым root-паролем до запуска новой конфигурации. Старый `.env`
больше не нужен: локальные значения заданы прямо в Compose. Не используйте
`docker compose down -v`, если эти данные нужны.

Для отдельных native/deployment запусков API читает `VALIO_WORKSPACE_ID`,
`VALIO_API_ADDRESS`, `VALIO_TRUSTED_ORIGINS`, а также параметры базы данных
`VALIO_DB_URL`, `VALIO_DB_NAMESPACE`, `VALIO_DB_DATABASE`, `VALIO_DB_USER`,
`VALIO_DB_PASSWORD` и `VALIO_DB_AUTH_LEVEL`. Ограничения HTTP и сессии описаны
в [OpenAPI contract](../../../src/back-end/api/openapi/openapi.json).

У native API значение `VALIO_API_ADDRESS` по умолчанию равно `127.0.0.1:8090`.
Локальный ingress остаётся `http://127.0.0.1:8080`, поэтому native agent обычно
использует ingress, если специально не тестируется API listener. В Windows
process-only override задаётся `$env:VALIO_API_TOKEN = '...'`; в Linux —
`export VALIO_API_TOKEN='...'`. Эти команды не переписывают Compose или
существующий volume.

Сейчас агент читает только `VALIO_API_TOKEN`, а не файл конфигурации агента.
Предлагаемые base server URL, spool/watch settings и Project profiles описаны в
[конфигурации локального агента](../design/agent-configuration.md) и пока не
реализованы.

Политика источников и исключения конфигурации применяются до spool, хеширования
и загрузки. Текущий поиск секретов эвристический, поэтому для конфиденциальных
данных задавайте явные исключения источников.
