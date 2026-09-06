import { useQuery } from "@tanstack/react-query";
import { Button } from "@heroui/react";
import { FileCode2, X } from "lucide-react";
import { request } from "../../api/client";
import type { SourceFile } from "../../models/workspace";
import type { ByteRange } from "../search/models";
import { Badge, ErrorState, Loading } from "../../components/Feedback";
import { sourceLines } from "./ranges";
export function SourceViewer({
  fileId,
  viewId,
  ranges,
  onClose,
}: {
  fileId: string;
  viewId: string;
  ranges?: ByteRange[];
  onClose: () => void;
}) {
  const file = useQuery({
    queryKey: ["file", fileId, viewId],
    queryFn: ({ signal }) =>
      request<SourceFile>(
        `/files/${encodeURIComponent(fileId)}?viewId=${encodeURIComponent(viewId)}`,
        { signal },
      ),
    enabled: !!fileId && !!viewId,
  });
  return (
    <section className="source-viewer" aria-label="Source viewer">
      <header>
        <FileCode2 size={17} />
        <strong>{file.data?.path || "Source"}</strong>
        {file.data?.language && <Badge>{file.data.language}</Badge>}
        <Button
          isIconOnly
          aria-label="Close source viewer"
          variant="ghost"
          size="sm"
          onPress={onClose}
        >
          <X size={17} />
        </Button>
      </header>
      {file.isPending ? (
        <Loading label="Loading source…" />
      ) : file.error ? (
        <ErrorState error={file.error} retry={() => void file.refetch()} />
      ) : (
        file.data && (
          <>
            <div className="source-meta">
              <span>{file.data.repositoryId}</span>
              <span>snapshot {file.data.snapshotId}</span>
              <span>projects {file.data.projectIds?.join(", ") || "none"}</span>
            </div>
            <div
              className="source-scroll"
              tabIndex={0}
              aria-label="Source code, scrollable"
            >
              <pre>
                <code>
                  {sourceLines(file.data.content, ranges).map((line) => (
                    <span
                      className={`code-line ${line.highlighted ? "highlighted" : ""}`}
                      key={line.number}
                    >
                      <span className="line-number" aria-hidden="true">
                        {line.number}
                      </span>
                      <span>{line.text || "\u200b"}</span>
                    </span>
                  ))}
                </code>
              </pre>
            </div>
            <footer>
              Original source text · UTF-8 match ranges highlight complete lines
            </footer>
          </>
        )
      )}
    </section>
  );
}
