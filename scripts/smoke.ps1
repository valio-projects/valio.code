param([string]$Server = 'http://127.0.0.1:8080')
$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot
$previousToken = $env:VALIO_API_TOKEN
try {
  $env:VALIO_API_TOKEN = 'valio-local-development-token-0001'
  $headers = @{ Authorization = "Bearer $($env:VALIO_API_TOKEN)" }
  function Send-API([string]$Path, $Body) {
    try {
      Invoke-RestMethod -Uri "$Server/api/v1/$Path" -Method Post -Headers $headers -ContentType 'application/json' -Body ($Body | ConvertTo-Json -Depth 20)
    } catch {
      throw "Smoke operation '$Path' failed with HTTP status $([int]$_.Exception.Response.StatusCode)"
    }
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
  $cSource = @'
typedef struct User { int id; } User;
'@
  $cppSource = @'
class User { public: int id; };
'@
  $csharpSource = @'
[Serializable] public class User { public string Name; }
[Flags] public enum State : byte { Active = 1, Disabled = 2 }
'@
  $javaSource = @'
@Deprecated public class User { private String name; public String label(String prefix) { return prefix + name; } }
'@
  $javascriptSource = @'
class User { constructor(name) { this.name = name; } label(prefix) { return prefix + this.name; } }
'@
  $typescriptSource = @'
export class User { constructor(public readonly name: string) {} label(prefix: string): string { return prefix + this.name; } }
'@
  Set-Content -LiteralPath (Join-Path $fixture 'user.go') -Value $source -Encoding utf8NoBOM
  Set-Content -LiteralPath (Join-Path $fixture 'user.c') -Value $cSource -Encoding utf8NoBOM
  Set-Content -LiteralPath (Join-Path $fixture 'user.cpp') -Value $cppSource -Encoding utf8NoBOM
  Set-Content -LiteralPath (Join-Path $fixture 'user.cs') -Value $csharpSource -Encoding utf8NoBOM
  Set-Content -LiteralPath (Join-Path $fixture 'User.java') -Value $javaSource -Encoding utf8NoBOM
  Set-Content -LiteralPath (Join-Path $fixture 'user.js') -Value $javascriptSource -Encoding utf8NoBOM
  Set-Content -LiteralPath (Join-Path $fixture 'user.ts') -Value $typescriptSource -Encoding utf8NoBOM
  Set-Content -LiteralPath (Join-Path $fixture '.env') -Value 'PASSWORD=fixture-secret-must-not-appear' -Encoding utf8NoBOM
  git -C $fixture init --quiet --initial-branch=main
  if ($LASTEXITCODE -ne 0) { throw 'Fixture Git init failed' }
  Send-API 'repositories' @{ id = $id; workspaceId = 'workspace-main' } | Out-Null
  Send-API 'projects' @{ project = @{ id = $id; workspaceId = 'workspace-main'; key = $id; name = 'Users demo'; kind = 'service'; status = 'active' }; roots = @(@{ repositoryId = $id; path = '.'; version = '1'; role = 'code' }) } | Out-Null
  $goPath = if (Test-Path .tools/go/bin/go.exe) { '.tools/go/bin/go.exe' } else { 'go' }
  $receiptText = & $goPath -C src/back-end run ./cmd/valio-agent index --root $fixture --server $Server --repository $id
  if ($LASTEXITCODE -ne 0) { throw 'Agent upload failed' }
  $receipt = $receiptText | ConvertFrom-Json
  $scope = @{ workspaceId = 'workspace-main'; projectIds = @($id); viewId = $receipt.viewId }
  $search = Send-API 'search' @{ query = 'User'; scope = $scope }
  if ($search.total -ne 7 -or !$search.complete) { throw 'Verified source search failed' }
  $symbolSearch = Send-API 'search' @{ query = 'symbol:User'; scope = $scope }
  if ($symbolSearch.total -ne 7 -or !$symbolSearch.complete) { throw 'Cross-language symbol search failed' }
  $types = Invoke-RestMethod -Uri "$Server/api/v1/types?name=User&projectId=$id&viewId=$($receipt.viewId)" -Headers $headers
  if (@($types.candidates).Count -ne 7) { throw 'Multi-language rich type lookup failed' }
  $structure = Send-API 'structure/graph' @{ name = 'User'; depth = 3; scope = $scope }
  if (@($structure.files).Count -ne 6 -or $structure.truncated) { throw 'Multi-language declaration graph failed' }
  $syntax = Send-API 'syntax/query' @{ scope = $scope; name = 'User' }
  $expectedSyntaxLanguages = @('c', 'cpp', 'csharp', 'java', 'javascript', 'typescript')
  $actualSyntaxLanguages = @($syntax.reports | ForEach-Object { $_.language } | Sort-Object)
  if (@($syntax.reports).Count -ne 6 -or ($actualSyntaxLanguages -join ',') -ne ($expectedSyntaxLanguages -join ',')) { throw 'Cross-language syntax reports failed' }
  $stateSyntax = Send-API 'syntax/query' @{ scope = $scope; name = 'State' }
  $stateFile = @($stateSyntax.reports | Where-Object { $_.language -eq 'csharp' })[0]
  if ($null -eq $stateFile) { throw 'C# syntax report missing' }
  $stateReport = $stateFile.report
  if ($stateReport -is [string]) { $stateReport = $stateReport | ConvertFrom-Json }
  $state = @($stateReport.symbols | Where-Object { $_.name -eq 'State' -and $_.kind -eq 'enum' })[0]
  if ($null -eq $state -or $state.underlyingType -ne 'byte' -or @($state.attributes) -notcontains '[Flags]') { throw 'C# enum syntax metadata failed' }
  $lexical = Send-API 'retrieval/search' @{ scope = $scope; query = 'User'; mode = 'lexical'; limit = 10 }
  if (@($lexical.hits).Count -eq 0) { throw 'Lexical retrieval failed' }
  $context = Send-API 'context' @{ scope = $scope; chunkId = $lexical.hits[0].chunk.id; maxBytes = 4096 }
  if (@($context.items).Count -eq 0) { throw 'Bounded retrieval context failed' }
  $userGraph = Send-API 'graph/query' @{ mode = 'symbols'; scope = $scope; name = 'User' }
  if (@($userGraph.nodes | Where-Object { $_.kind -eq 'symbol' -and $_.name -eq 'User' }).Count -eq 0) { throw 'Go graph symbol query failed' }
  $nameGraph = Send-API 'graph/query' @{ mode = 'symbols'; scope = $scope; name = 'Name' }
  $nameNode = @($nameGraph.nodes | Where-Object { $_.kind -eq 'symbol' -and $_.name -eq 'Name' })[0]
  if ($null -eq $nameNode) { throw 'Go graph field symbol query failed' }
  $reads = Send-API 'graph/query' @{ mode = 'reads'; scope = $scope; targetNodeId = $nameNode.id }
  if (@($reads.edges).Count -eq 0) { throw 'Go graph read query failed' }
  $secret = Send-API 'search' @{ query = 'fixture-secret-must-not-appear'; scope = $scope }
  if ($secret.total -ne 0) { throw 'Configuration policy failed' }
  $again = & $goPath -C src/back-end run ./cmd/valio-agent index --root $fixture --server $Server --repository $id
  if ($LASTEXITCODE -ne 0 -or ($again | ConvertFrom-Json).viewId -ne $receipt.viewId) { throw 'Idempotent upload failed' }
  Write-Output "Smoke passed: project=$id view=$($receipt.viewId); cross-language syntax, retrieval, Go graph, types, config policy and duplicate delivery."
} finally {
  $env:VALIO_API_TOKEN = $previousToken
  Pop-Location
}
