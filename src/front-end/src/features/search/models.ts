export interface ByteRange {
  start: number;
  end: number;
}
export interface SearchMatch {
  fileId: string;
  path: string;
  repositoryId: string;
  snapshotId: string;
  projectIds: string[];
  ranges: ByteRange[];
  rangesTruncated?: boolean;
}
export interface SearchResult {
  viewId: string;
  matches: SearchMatch[];
  total: number;
  complete: boolean;
  truncated: boolean;
  scannedFiles: number;
  scannedBytes: number;
}
export interface SearchRequest {
  query: string;
  mode: string;
  scope: { workspaceId: string; projectIds: string[]; viewId: string };
  limit: number;
  offset: number;
  caseSensitive: boolean;
}
export function searchRequest(
  workspaceId: string,
  projectId: string,
  viewId: string,
  query: string,
  mode = "substring",
  caseSensitive = false,
  offset = 0,
): SearchRequest {
  return {
    query,
    mode,
    scope: { workspaceId, projectIds: projectId ? [projectId] : [], viewId },
    limit: 50,
    offset,
    caseSensitive,
  };
}
