import { useEffect, useState, type FormEvent } from "react";
import { Button, Input } from "@heroui/react";
import { useQuery } from "@tanstack/react-query";
import { ArrowRight, FileCode2, Search } from "lucide-react";
import { request } from "../../api/client";
import { Badge, Empty, ErrorState, Loading } from "../../components/Feedback";
import type { AnalysisView, Workspace } from "../../models/workspace";
import { useNavigation } from "../../providers/NavigationProvider";
import { SourceViewer } from "../source/SourceViewer";
import { searchRequest, type SearchResult } from "./models";
export function SearchPage({
  workspace,
  view,
}: {
  workspace: Workspace;
  view?: AnalysisView;
}) {
  const { state, navigate } = useNavigation();
  const [query, setQuery] = useState(state.q);
  const [offset, setOffset] = useState(0);
  useEffect(() => {
    setQuery(state.q);
    setOffset(0);
  }, [state.q, state.project, state.view, state.mode, state.caseSensitive]);
  const payload = searchRequest(
    workspace.id,
    state.project,
    view?.id || "",
    state.q,
    state.mode,
    state.caseSensitive,
    offset,
  );
  const result = useQuery({
    queryKey: ["search", payload],
    queryFn: ({ signal }) =>
      request<SearchResult>("/search", {
        method: "POST",
        body: JSON.stringify(payload),
        signal,
      }),
    enabled: !!view && !!state.q,
  });
  function submit(e: FormEvent) {
    e.preventDefault();
    setOffset(0);
    navigate({ q: query.trim(), file: "" });
  }
  const selected = result.data?.matches?.find((m) => m.fileId === state.file);
  return (
    <>
      <div className="page-heading">
        <div>
          <div className="eyebrow">EXPLORE / SOURCE SEARCH</div>
          <h1>
            Find your way through code<span className="title-dot">.</span>
          </h1>
          <p>
            Search the selected snapshot with text, filters, and boolean
            expressions.
          </p>
        </div>
      </div>
      <form className="query-bar" onSubmit={submit}>
        <label className="search-input">
          <Search size={20} />
          <Input
            aria-label="Search query"
            placeholder="Search code, symbols, paths…"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </label>
        <Button
          type="submit"
          isDisabled={!view || !query.trim()}
          isPending={result.isFetching}
        >
          Search <ArrowRight size={16} />
        </Button>
      </form>
      <div className="search-options">
        <label>
          Match mode
          <select
            value={state.mode}
            onChange={(e) => navigate({ mode: e.target.value, file: "" })}
          >
            <option value="substring">Substring</option>
            <option value="exact">Exact field</option>
            <option value="regex">Regular expression</option>
          </select>
        </label>
        <label className="check-label">
          <input
            type="checkbox"
            checked={state.caseSensitive}
            onChange={(e) =>
              navigate({ caseSensitive: e.target.checked, file: "" })
            }
          />
          Case sensitive
        </label>
        <details className="syntax-help">
          <summary>Query syntax</summary>
          <p>
            <code>lang:go AND content:"NewClient"</code>
            <br />
            <code>path:internal NOT test:true</code>
            <br />
            Use AND, OR, NOT, parentheses, quotes, and /RE2 expressions/.
            Filters: project, repo, file, lang, path, content, symbol, kind,
            test, generated.
          </p>
        </details>
      </div>
      {!view ? (
        <Empty title="Source has not been indexed">
          <p>
            Open Workspace for the indexing steps. Search becomes available
            after a view is published.
          </p>
        </Empty>
      ) : !state.q ? (
        <Empty title="A precise question. A traceable answer.">
          <p>
            Search content, narrow by project, and open the matching source.
            <br />
            Every result stays tied to the selected analysis view.
          </p>
        </Empty>
      ) : result.isPending ? (
        <Loading label="Searching the selected view…" />
      ) : result.error ? (
        <ErrorState error={result.error} retry={() => void result.refetch()} />
      ) : (
        result.data && (
          <>
            <div className="results-heading">
              <strong>{result.data.total} matching files</strong>
              <Badge status={result.data.complete ? "known" : "partial"}>
                {result.data.complete ? "Complete scan" : "Partial scan"}
              </Badge>
              {result.data.truncated && (
                <Badge status="partial">Truncated results</Badge>
              )}
              <span>
                {result.data.scannedFiles} files ·{" "}
                {result.data.scannedBytes.toLocaleString()} bytes scanned
              </span>
            </div>
            {!result.data.complete && (
              <p className="notice">
                The scan reached its budget. The reported count is a lower bound
                for the selected scope.
              </p>
            )}
            {result.data.complete && result.data.truncated && (
              <p className="notice">
                Some source highlights were omitted. The matching-file count is
                exact for the selected view and scope.
              </p>
            )}
            <div
              className={`search-results ${state.file ? "with-source" : ""}`}
            >
              <div className="result-list">
                {result.data.matches?.length ? (
                  result.data.matches.map((m) => (
                    <button
                      className={`search-result ${state.file === m.fileId ? "selected" : ""}`}
                      key={m.fileId}
                      onClick={() => navigate({ file: m.fileId })}
                    >
                      <FileCode2 size={18} />
                      <span>
                        <strong>{m.path}</strong>
                        <small>
                          {m.repositoryId} ·{" "}
                          {m.projectIds?.join(", ") || "Outside project roots"}
                        </small>
                        <span className="result-ranges">
                          {m.ranges?.length
                            ? `${m.ranges.length} reported source ranges`
                            : "Matched file metadata"}
                          {m.rangesTruncated && " · highlights truncated"}
                        </span>
                      </span>
                      <ArrowRight size={15} />
                    </button>
                  ))
                ) : (
                  <Empty title="No matching files">
                    <p>Try a broader query or another project scope.</p>
                  </Empty>
                )}
                <div className="pagination">
                  <Button
                    variant="outline"
                    size="sm"
                    isDisabled={offset === 0}
                    onPress={() => setOffset((v) => Math.max(0, v - 50))}
                  >
                    Previous
                  </Button>
                  <span>
                    {result.data.matches?.length ? offset + 1 : 0}–
                    {offset + (result.data.matches?.length || 0)}
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    isDisabled={offset + 50 >= result.data.total}
                    onPress={() => setOffset((v) => v + 50)}
                  >
                    Next
                  </Button>
                </div>
              </div>
              {state.file && (
                <SourceViewer
                  fileId={state.file}
                  viewId={result.data.viewId}
                  ranges={selected?.ranges}
                  onClose={() => navigate({ file: "" })}
                />
              )}
            </div>
          </>
        )
      )}
    </>
  );
}
