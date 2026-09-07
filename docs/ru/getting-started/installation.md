# Установка и запуск

Выберите инструменты для выполняемой работы:

- [Docker Desktop с Compose](https://docs.docker.com/compose/install/) нужен для
  локального стека и изолированной integration topology (Compose 2.24.4 или новее).
- [Git](https://git-scm.com/downloads) нужен для clone репозитория и capture
  worktree агентом.
- [Go 1.26.8](https://go.dev/doc/install) нужен для native-agent и backend
  разработки.
- [Node.js 26.8.1](https://nodejs.org/en/download) нужен для frontend разработки.

В чистом checkout нет `.tools/go`. Сначала проверьте обычный установленный Go;
переносимый toolchain используйте только после отдельной подготовки:

```powershell
docker compose version
go version
node --version
```

В Windows подготовленный portable toolchain запускается
`& .\.tools\go\bin\go.exe version`, в Linux — `./.tools/go/bin/go version`.
Это необязательные альтернативы `go version`.

Из корня репозитория запустите локальный стек:

```powershell
docker compose up -d --build --wait
```

Откройте приложение по адресу <http://127.0.0.1:8080>, Jaeger — по адресу
<http://127.0.0.1:16686>. SurrealDB доступна только внутри сети Compose.

Токен входа для локальной разработки:
`valio-local-development-token-0001`. Браузер обменивает его на HttpOnly cookie.

Разработка backend:

```powershell
# Запускайте блок из корня репозитория.
cd src/back-end
go test ./...
go vet ./...
go run ./api/openapi
```

Проверка frontend:

```powershell
# Запускайте блок из корня репозитория.
cd src/front-end
npm ci
npm test
npm run build
```

Из корня репозитория:

```powershell
docker build -f src/back-end/Dockerfile --target test -t valio-code-test .
.\scripts\test-integration.ps1
.\scripts\smoke.ps1
```

Интеграционный скрипт запускает только изолированные database и Jaeger, а затем
удаляет этот project в `finally`; базовый локальный стек он не останавливает. Для
override портов нужен Docker Compose 2.24.4 или новее:

```powershell
docker compose -p valio-code-test -f compose.yaml -f deploy/compose.test.yaml up -d --wait surrealdb jaeger
```

Изолированные endpoints: SurrealDB `127.0.0.1:18000`, OTLP `127.0.0.1:14318`,
Jaeger UI `127.0.0.1:16687`. Базовый Jaeger UI остаётся на `127.0.0.1:16686`.

Для загрузки репозитория запустите агент из `src/back-end`; его локальный токен по
умолчанию совпадает с токеном Compose:

```powershell
go run ./cmd/valio-agent index --root C:/path/to/repository --server http://127.0.0.1:8080 --repository repo-id
go run ./cmd/valio-agent watch --root C:/path/to/repository --server http://127.0.0.1:8080 --repository repo-id
```

Без `--server` команда `index` создаёт очищенный снимок. `watch` отслеживает
изменения, периодически сверяет дерево и повторяет доставку.

## Первый Project и загрузка репозитория

Текущий CLI не создаёт Workspace, Repository, Project или source-root attachment.
В веб-интерфейсе создайте Repository и Project, прикрепите source roots и
скопируйте возвращённый repository ID. Затем из `src/back-end` запустите
поддерживаемые флаги агента:

```powershell
go run ./cmd/valio-agent index --root C:/src/users-api --server http://127.0.0.1:8080 --workspace workspace-main --repository repo-users
go run ./cmd/valio-agent watch --root C:/src/users-api --server http://127.0.0.1:8080 --workspace workspace-main --repository repo-users --interval 5s
```

В Linux абсолютный root может быть `/work/users-api`. `--server` включает
загрузку; без него агент выводит локальный очищенный snapshot. `--spool` по
умолчанию `<root>/.valio/spool`, `--max-file-bytes` ограничен 2 MiB. Другой
локальный токен нужен только при необходимости:

```powershell
$env:VALIO_API_TOKEN = 'valio-local-development-token-0002'
```

```bash
export VALIO_API_TOKEN='valio-local-development-token-0002'
```

Такой override используйте только с API/server, настроенным на тот же токен. При
изменении Compose обновите API token и выполните
`docker compose up -d --force-recreate api` до отправки запроса. Токен базового
Compose уже совпадает с default агента, поэтому обычному локальному стеку override
не нужен. Будущая модель base-file/profile описана в
[конфигурации локального агента](../design/agent-configuration.md); `--config` и
`--profile` сейчас недопустимы.

## Проверка и диагностика

Из корня репозитория проверьте нужный сервис:

```powershell
docker compose ps
docker compose logs --tail 100 api worker
Invoke-WebRequest http://127.0.0.1:8080/health
Invoke-WebRequest http://127.0.0.1:8080/readyz
```

В Linux используйте `curl -fsS http://127.0.0.1:8080/health` и
`curl -fsS http://127.0.0.1:8080/readyz`. `/health` проверяет nginx, `/readyz`
проксируется к API. При ошибке загрузки изучите JSON diagnostic агента и повторите
отправку из того же spool; не копируйте значения конфигурации в команды или логи.
