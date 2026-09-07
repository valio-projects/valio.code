param(
  [Parameter(Mandatory)][string]$ViewId,
  [Parameter(Mandatory)][string]$ProjectId,
  [string]$Server = 'http://127.0.0.1:8080',
  [string]$ModelProfile = 'qwen3-8b-lmstudio-docker',
  [int]$ExpectedDimension = 4096
)
$ErrorActionPreference = 'Stop'
# Use the synthetic project and pinned view printed by smoke.ps1. This check
# creates embeddings only for that explicit scope, never the whole workspace.
$headers = @{ Authorization = 'Bearer valio-local-development-token-0001' }
function Send-API([string]$Path, $Body) {
  try {
    Invoke-RestMethod -Uri "$Server/api/v1/$Path" -Method Post -Headers $headers `
      -ContentType application/json -Body ($Body | ConvertTo-Json -Depth 12) -TimeoutSec 180
  } catch {
    throw "AI smoke operation '$Path' failed with HTTP status $([int]$_.Exception.Response.StatusCode)"
  }
}
$scope = @{ workspaceId = 'workspace-main'; viewId = $ViewId; projectIds = @($ProjectId) }
$probe = Send-API 'ai/probe' @{ modelProfile = $ModelProfile }
if ($probe.status -ne 'ready' -or $probe.dimension -ne $ExpectedDimension) { throw 'Provider dimension probe failed' }
$offset = 0
$finished = $false
for ($page = 0; $page -lt 32; $page++) {
  $indexed = Send-API 'embeddings/index' @{ scope = $scope; modelProfile = $ModelProfile; representation = 'code'; offset = $offset; limit = 8 }
  if ($indexed.eligible -le 0) { throw 'Fixture has no code representations' }
  if ($indexed.complete) { $finished = $true; break }
  if ($indexed.nextOffset -le $offset) { throw 'Embedding indexing made no progress' }
  $offset = $indexed.nextOffset
}
if (!$finished) { throw 'Fixture embedding budget exceeded' }
foreach ($mode in @('semantic', 'hybrid')) {
  $result = Send-API 'retrieval/search' @{ scope = $scope; modelProfile = $ModelProfile; mode = $mode; query = 'user identifier and name'; limit = 5 }
  if ($result.viewId -ne $ViewId -or @($result.hits).Count -eq 0) { throw "No pinned $mode retrieval results" }
  foreach ($hit in $result.hits) {
    if ($ProjectId -notin $hit.chunk.projectIds) { throw 'Retrieval escaped the selected fixture project' }
  }
  $semantic = @($result.hits | Where-Object { 'semantic' -in $_.provenance.channel })
  if ($semantic.Count -eq 0) { throw "$mode retrieval did not use the embedding channel" }
}
Write-Output "AI smoke passed: profile=$ModelProfile dimension=$($probe.dimension) representations=$($indexed.eligible); persisted embeddings, semantic and hybrid retrieval in project=$ProjectId view=$ViewId."
