param([string]$Server = 'http://127.0.0.1:8080')
$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot
$previousToken = $env:VALIO_API_TOKEN
try {
  $settings = @{}
  Get-Content .env | ForEach-Object { $pair = $_ -split '=', 2; $settings[$pair[0]] = $pair[1] }
  $env:VALIO_API_TOKEN = $settings['VALIO_API_TOKEN']
  $headers = @{ Authorization = "Bearer $($env:VALIO_API_TOKEN)" }
  function Send-API([string]$Path, $Body) {
    Invoke-RestMethod -Uri "$Server/api/v1/$Path" -Method Post -Headers $headers -ContentType 'application/json' -Body ($Body | ConvertTo-Json -Depth 20)
  }
  # A distinct fixture remains available for manual inspection without touching user repositories.
  $id = 'demo-' + [Guid]::NewGuid().ToString('N').Substring(0, 12)
  $fixture = Join-Path $repoRoot ".valio/$id"
  New-Item -ItemType Directory -Path $fixture -Force | Out-Null
  $source = @'
package users

// User is a small source fixture for type inspection.
type User struct {
    ID int64 `json:"id"`
    Name string `json:"name"`
}

// Label exposes a value derived from a declared field.
func (u User) Label(prefix string) string { return prefix + u.Name }
'@
  Set-Content -LiteralPath (Join-Path $fixture 'user.go') -Value $source -Encoding utf8NoBOM
  Set-Content -LiteralPath (Join-Path $fixture '.env') -Value 'PASSWORD=fixture-secret-must-not-appear' -Encoding utf8NoBOM
  git -C $fixture init --quiet --initial-branch=main
  if ($LASTEXITCODE -ne 0) { throw 'Fixture Git init failed' }
  Send-API 'repositories' @{ id = $id; workspaceId = 'workspace-main' } | Out-Null
  Send-API 'projects' @{ project = @{ id = $id; workspaceId = 'workspace-main'; key = $id; name = 'Users demo'; kind = 'service'; status = 'active' }; roots = @(@{ repositoryId = $id; path = '.'; version = '1'; role = 'code' }) } | Out-Null
  $goPath = if (Test-Path .tools/go/bin/go.exe) { '.tools/go/bin/go.exe' } else { 'go' }
  $receiptText = & $goPath run ./cmd/valio-agent index --root $fixture --server $Server --repository $id
  if ($LASTEXITCODE -ne 0) { throw 'Agent upload failed' }
  $receipt = $receiptText | ConvertFrom-Json
  $scope = @{ workspaceId = 'workspace-main'; projectIds = @($id); viewId = $receipt.viewId }
  $search = Send-API 'search' @{ query = 'User'; scope = $scope }
  if ($search.total -ne 1 -or !$search.complete) { throw 'Verified source search failed' }
  $types = Invoke-RestMethod -Uri "$Server/api/v1/types?name=User&projectId=$id&viewId=$($receipt.viewId)" -Headers $headers
  if (@($types.candidates).Count -ne 1) { throw 'Rich type lookup failed' }
  $secret = Send-API 'search' @{ query = 'fixture-secret-must-not-appear'; scope = $scope }
  if ($secret.total -ne 0) { throw 'Configuration policy failed' }
  $again = & $goPath run ./cmd/valio-agent index --root $fixture --server $Server --repository $id
  if ($LASTEXITCODE -ne 0 -or ($again | ConvertFrom-Json).viewId -ne $receipt.viewId) { throw 'Idempotent upload failed' }
  Write-Output "Smoke passed: project=$id view=$($receipt.viewId); search, types, config policy and duplicate delivery."
} finally {
  $env:VALIO_API_TOKEN = $previousToken
  Pop-Location
}
