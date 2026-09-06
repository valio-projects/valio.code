import { useEffect, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { ApiError, logout } from "./api/client";
import { useView, useWorkspace } from "./api/queries";
import { AppShell } from "./components/AppShell";
import { ErrorState, Loading } from "./components/Feedback";
import { useNavigation } from "./providers/NavigationProvider";
import { LoginPage } from "./features/auth/LoginPage";
import { WorkspacePage } from "./features/workspace/WorkspacePage";
import { ProjectsPage } from "./features/projects/ProjectsPage";
import { SearchPage } from "./features/search/SearchPage";
import { TypesPage } from "./features/types/TypesPage";
import { CapabilitiesPage } from "./features/capabilities/CapabilitiesPage";
function Workbench({
  workspace,
}: {
  workspace: NonNullable<ReturnType<typeof useWorkspace>["data"]>;
}) {
  const { state } = useNavigation();
  const client = useQueryClient();
  const view = useView(state.view);
  const [sessionError, setSessionError] = useState<unknown>();
  const noView = view.error instanceof ApiError && view.error.status === 404;
  async function endSession() {
    try {
      await logout();
      client.clear();
      window.dispatchEvent(new Event("valio:unauthorized"));
    } catch (e) {
      setSessionError(e);
    }
  }
  return (
    <AppShell
      workspace={workspace}
      view={view.data}
      refresh={() => void client.invalidateQueries()}
      onLogout={() => void endSession()}
    >
      {sessionError ? <ErrorState error={sessionError} /> : null}
      {view.error && !noView && (
        <ErrorState error={view.error} retry={() => void view.refetch()} />
      )}{" "}
      {noView && state.view && (
        <p className="notice" role="status">
          The pinned view could not be found. Load another view or select “Use
          latest”.
        </p>
      )}
      {view.isPending && <Loading label="Loading analysis view…" />}
      {state.page === "workspace" && (
        <WorkspacePage workspace={workspace} view={view.data} />
      )}{" "}
      {state.page === "projects" && <ProjectsPage workspace={workspace} />}{" "}
      {state.page === "search" && (
        <SearchPage workspace={workspace} view={view.data} />
      )}{" "}
      {state.page === "types" && <TypesPage view={view.data} />}{" "}
      {state.page === "capabilities" && <CapabilitiesPage />}
    </AppShell>
  );
}
export function App() {
  const workspace = useWorkspace();
  const client = useQueryClient();
  const [unauthorized, setUnauthorized] = useState(false);
  useEffect(() => {
    const listener = () => setUnauthorized(true);
    window.addEventListener("valio:unauthorized", listener);
    return () => window.removeEventListener("valio:unauthorized", listener);
  }, []);
  if (
    unauthorized ||
    (workspace.error instanceof ApiError && workspace.error.status === 401)
  )
    return (
      <LoginPage
        onSuccess={() => {
          client.clear();
          setUnauthorized(false);
          void client.invalidateQueries();
        }}
      />
    );
  if (workspace.isPending)
    return (
      <main className="boot-screen">
        <div className="brand dark-brand">
          v <span>valio.code</span>
        </div>
        <Loading />
      </main>
    );
  if (workspace.error || !workspace.data)
    return (
      <main className="boot-screen">
        <div className="brand dark-brand">
          v <span>valio.code</span>
        </div>
        <h1>Workspace unavailable</h1>
        <ErrorState
          error={workspace.error}
          retry={() => void workspace.refetch()}
        />
      </main>
    );
  return <Workbench workspace={workspace.data} />;
}
