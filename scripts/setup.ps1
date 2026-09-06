$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$envPath = Join-Path $repoRoot '.env'
if (Test-Path -LiteralPath $envPath) {
  Write-Output 'Existing .env preserved.'
  exit 0
}
function New-Secret {
  $bytes = [byte[]]::new(32)
  [Security.Cryptography.RandomNumberGenerator]::Fill($bytes)
  return [Convert]::ToHexString($bytes).ToLowerInvariant()
}
$content = @(
  'SURREAL_ROOT_PASSWORD=' + (New-Secret)
  'VALIO_API_DB_PASSWORD=' + (New-Secret)
  'VALIO_WORKER_DB_PASSWORD=' + (New-Secret)
  'VALIO_API_TOKEN=' + (New-Secret)
)
[IO.File]::WriteAllLines($envPath, $content)
Write-Output 'Generated local credentials in ignored .env. Keep this file private.'
