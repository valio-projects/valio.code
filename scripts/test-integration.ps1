$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot
$savedEnvironment = @{}
foreach ($name in @('VALIO_TEST_DB_URL', 'VALIO_TEST_DB_PASSWORD', 'VALIO_TEST_OTLP_URL', 'VALIO_TEST_SYNTAX_HELPER')) {
  $savedEnvironment[$name] = [Environment]::GetEnvironmentVariable($name)
}
try {
  Push-Location src/back-end/analyzers/syntax
  try {
    npm ci --ignore-scripts
    if ($LASTEXITCODE -ne 0) { throw 'Syntax helper dependency installation failed' }
    npm test
    if ($LASTEXITCODE -ne 0) { throw 'Syntax helper fixtures failed' }
  } finally { Pop-Location }
  $env:VALIO_TEST_SYNTAX_HELPER = (Resolve-Path src/back-end/analyzers/syntax/index.js).Path
  $env:VALIO_TEST_DB_URL = 'http://127.0.0.1:18000'
  $env:VALIO_TEST_DB_PASSWORD = 'valio'
  $env:VALIO_TEST_OTLP_URL = 'http://127.0.0.1:14318'
  docker compose -p valio-code-test -f compose.yaml -f deploy/compose.test.yaml up -d --wait surrealdb jaeger
  if ($LASTEXITCODE -ne 0) { throw 'Compose database startup failed' }
  $goPath = if (Test-Path .tools/go/bin/go.exe) { Join-Path $repoRoot '.tools/go/bin/go.exe' } else { 'go' }
  & $goPath -C src/back-end test ./...
  if ($LASTEXITCODE -ne 0) { throw 'Integration tests failed' }
} finally {
  docker compose -p valio-code-test -f compose.yaml -f deploy/compose.test.yaml down
  foreach ($name in $savedEnvironment.Keys) {
    [Environment]::SetEnvironmentVariable($name, $savedEnvironment[$name])
  }
  Pop-Location
}
