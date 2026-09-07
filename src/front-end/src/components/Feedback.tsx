import { Button } from "@heroui/react";
import { AlertCircle, LoaderCircle } from "lucide-react";
import { ApiError } from "../api/client";
export function Loading({ label = "Loading workspace…" }: { label?: string }) {
  return (
    <div className="feedback" role="status">
      <LoaderCircle className="spin" size={20} />
      {label}
    </div>
  );
}
export function ErrorState({
  error,
  retry,
}: {
  error: unknown;
  retry?: () => void;
}) {
  return (
    <div className="error-state" role="alert">
      <AlertCircle size={20} />
      <div>
        <strong>Unable to load this data</strong>
        <p>
          {error instanceof Error
            ? error.message
            : "An unexpected error occurred."}
        </p>
        {error instanceof ApiError && error.requestId && (
          <small>Request {error.requestId}</small>
        )}
      </div>
      {retry && (
        <Button variant="outline" size="sm" onPress={retry}>
          Retry
        </Button>
      )}
    </div>
  );
}
export function Badge({
  children,
  status,
}: {
  children: React.ReactNode;
  status?: string;
}) {
  return <span className={`badge ${status || ""}`}>{children}</span>;
}
export function Empty({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div className="empty-state">
      <span className="empty-mark">[ ]</span>
      <h2>{title}</h2>
      <div>{children}</div>
    </div>
  );
}
