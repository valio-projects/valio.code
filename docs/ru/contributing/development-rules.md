# Участие в разработке и правила

Перед изменениями прочитайте [правила участия](../../../CONTRIBUTING.md) и
[`AGENTS.md`](../../../AGENTS.md). Сохраняйте лицензию AGPL-3.0, remote и
несвязанные изменения пользователя.

Работайте из каталога изменяемого компонента:

```powershell
# Запускайте блок из корня репозитория.
cd src/back-end
go test ./...
go vet ./...
go run ./api/openapi
```

```powershell
# Запускайте блок из корня репозитория.
cd src/front-end
npm ci
npm test
npm run build
```

Для изменений хранилища или Compose запускайте из корня релевантные проверки:

```powershell
.\scripts\test-integration.ps1
docker build -f src/back-end/Dockerfile --target test -t valio-code-test .
```

`test-integration.ps1` изолирует сервисы командой
`docker compose -p valio-code-test -f compose.yaml -f deploy/compose.test.yaml up -d --wait surrealdb jaeger`
и удаляет только этот project без volumes. Его порты database, OTLP и Jaeger UI:
18000, 14318 и 16687; базовый Jaeger UI остаётся на 16686. Для port override
нужен Docker Compose 2.24.4 или новее.

Root-level orchestration запускайте из корня репозитория, а module-level checks —
из каталога компонента; не полагайтесь на каталог, оставшийся от предыдущей команды.
В Windows bundled toolchain вызывается как
`& .\.tools\go\bin\go.exe -C src\back-end test ./...`; в Linux —
`./.tools/go/bin/go -C src/back-end test ./...`. Frontend эквивалентен на обеих
платформах: `cd src/front-end; npm ci; npm test; npm run build`.

Не смешивайте понятия Workspace, Project, Repository, Worktree и Deployment.
SurrealDB — единственное постоянное хранилище приложения. Domain не зависит от
HTTP, SDK базы данных, SDK компилятора и DI; интерфейсы репозиториев реализуются
адаптерами infrastructure. Сервисы не формируют SurrealQL. Operational logs не
должны содержать source или raw configuration payloads. Policy-accepted очищенные
исходники можно capture/upload/spool для повторной доставки; значения конфигурации
и исключённые пути никогда туда не попадают.

Фиксируйте реализованное поведение, проверки и неизвестные ограничения. Не
выдавайте эвристики AST за факты компилятора и не представляйте неизвестные
метаданные нулевыми значениями.
