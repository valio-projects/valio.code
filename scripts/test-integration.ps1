$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot
try {
  $settings = @{}
  Get-Content .env | ForEach-Object { $parts = $_ -split '=', 2; $settings[$parts[0]] = $parts[1] }
  $env:VALIO_TEST_DB_URL = 'http://127.0.0.1:18000'
  $env:VALIO_TEST_DB_PASSWORD = $settings['SURREAL_ROOT_PASSWORD']
  $env:VALIO_TEST_OTLP_URL = 'http://127.0.0.1:14318'
  docker compose -f compose.yaml -f deploy/compose.test.yaml up -d --wait surrealdb jaeger
  if ($LASTEXITCODE -ne 0) { throw 'Compose database startup failed' }
  $goPath = if (Test-Path .tools/go/bin/go.exe) { '.tools/go/bin/go.exe' } else { 'go' }
  & $goPath test ./...
  if ($LASTEXITCODE -ne 0) { throw 'Integration tests failed' }
} finally {
  Remove-Item Env:VALIO_TEST_DB_PASSWORD -ErrorAction SilentlyContinue
  Remove-Item Env:VALIO_TEST_DB_URL -ErrorAction SilentlyContinue
  Remove-Item Env:VALIO_TEST_OTLP_URL -ErrorAction SilentlyContinue
  Pop-Location
}
