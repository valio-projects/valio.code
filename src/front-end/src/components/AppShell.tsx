import { useEffect, useState, type ReactNode, type FormEvent } from "react";
import { Button, Input } from "@heroui/react";
import {
  ArrowUpRight,
  Braces,
  ChevronDown,
  Command,
  FolderCode,
  GitBranch,
  Layers3,
  LogOut,
  Network,
  RefreshCw,
  Search,
  ShieldCheck,
} from "lucide-react";
import { useNavigation, type Page } from "../providers/NavigationProvider";
import { useProjects } from "../api/queries";
import type { AnalysisView, Workspace } from "../models/workspace";
import { Badge } from "./Feedback";
const nav: { page: Page; label: string; icon: typeof Layers3 }[] = [
  { page: "workspace", label: "Workspace", icon: Layers3 },
  { page: "projects", label: "Projects", icon: FolderCode },
  { page: "search", label: "Source search", icon: Search },
  { page: "types", label: "Type inspector", icon: Braces },
  { page: "capabilities", label: "Capabilities", icon: ShieldCheck },
];
export function AppShell({
  workspace,
  view,
  refresh,
  onLogout,
  children,
}: {
  workspace: Workspace;
  view?: AnalysisView;
  refresh: () => void;
  onLogout: () => void;
  children: ReactNode;
}) {
  const { state, navigate } = useNavigation();
  const projects = useProjects();
  const [viewInput, setViewInput] = useState(state.view);
  useEffect(() => setViewInput(state.view), [state.view]);
  function pin(e: FormEvent) {
    e.preventDefault();
    navigate({ view: viewInput.trim(), file: "" });
  }
  return (
    <div className="app-shell">
      <a className="skip-link" href="#main-content">
        Skip to content
      </a>
      <aside className="sidebar">
        <a
          className="brand"
          href="?page=workspace"
          onClick={(e) => {
            e.preventDefault();
            navigate({ page: "workspace" });
          }}
        >
          <span className="brand-symbol">v</span>
          <span>
            valio<span className="brand-code">.code</span>
          </span>
        </a>
        <div className="workspace-switch">
          <span className="workspace-avatar">
            {workspace.name.charAt(0).toUpperCase()}
          </span>
          <div>
            <strong>{workspace.name}</strong>
            <small>Workspace</small>
          </div>
          <ChevronDown size={14} />
        </div>
        <div className="nav-caption">WORKSPACE</div>
        <nav aria-label="Main navigation">
          {nav.slice(0, 2).map((n) => (
            <button
              className={state.page === n.page ? "active" : ""}
              aria-current={state.page === n.page ? "page" : undefined}
              key={n.page}
              onClick={() => navigate({ page: n.page })}
            >
              <n.icon size={18} />
              {n.label}
              {state.page === n.page && <span className="nav-dot" />}
            </button>
          ))}
          <div className="nav-caption">EXPLORE</div>
          {nav.slice(2, 4).map((n) => (
            <button
              className={state.page === n.page ? "active" : ""}
              aria-current={state.page === n.page ? "page" : undefined}
              key={n.page}
              onClick={() => navigate({ page: n.page })}
            >
              <n.icon size={18} />
              {n.label}
              {state.page === n.page && <span className="nav-dot" />}
            </button>
          ))}
          <button
            disabled
            title="A graph projection and graph UI are not available in this deployment"
          >
            <Network size={18} />
            Relationship graph<span className="soon">Later</span>
          </button>
          <div className="nav-caption">SYSTEM</div>
          <button
            className={state.page === "capabilities" ? "active" : ""}
            onClick={() => navigate({ page: "capabilities" })}
          >
            <ShieldCheck size={18} />
            Capabilities
          </button>
        </nav>
        <div className="sidebar-bottom">
          <div className="local-label">
            <span className="tiny-dot" />
            Local deployment
          </div>
          <p>
            Source, scope,
            <br />
            and the facts between.
          </p>
          <Button variant="ghost" size="sm" onPress={onLogout}>
            <LogOut size={15} />
            End session
          </Button>
        </div>
      </aside>
      <div className="main-shell">
        <header className="topbar">
          <div className="breadcrumb">
            <Layers3 size={15} />
            <span>{workspace.name}</span>
            <span>/</span>
            <strong>{nav.find((n) => n.page === state.page)?.label}</strong>
          </div>
          <button
            className="quick-search"
            onClick={() => navigate({ page: "search" })}
          >
            <Search size={15} />
            <span>Search workspace</span>
            <Command size={12} />
            <ArrowUpRight size={12} />
          </button>
        </header>
        <div className="scopebar">
          <label>
            <GitBranch size={14} />
            <select
              aria-label="Project scope"
              value={state.project}
              onChange={(e) => navigate({ project: e.target.value, file: "" })}
            >
              <option value="">All projects</option>
              {(view?.projects || projects.data?.items || []).map((d) => (
                <option key={d.project.id} value={d.project.id}>
                  {d.project.name}
                </option>
              ))}
              {state.project &&
                !(view?.projects || projects.data?.items || []).some(
                  (d) => d.project.id === state.project,
                ) && (
                  <option value={state.project}>
                    {state.project} · unavailable
                  </option>
                )}
            </select>
          </label>
          <span className="scope-divider" />
          <details className="view-picker">
            <summary>
              <Layers3 size={14} />
              <span>
                {state.view
                  ? `Pinned · ${state.view.slice(0, 12)}`
                  : "Latest analysis view"}
              </span>
              <ChevronDown size={12} />
            </summary>
            <form onSubmit={pin}>
              <label htmlFor="view-id">Pin an immutable view</label>
              <Input
                id="view-id"
                placeholder="View ID"
                value={viewInput}
                onChange={(e) => setViewInput(e.target.value)}
              />
              <div>
                <Button size="sm" type="submit">
                  Load view
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  onPress={() => {
                    setViewInput("");
                    navigate({ view: "", file: "" });
                  }}
                >
                  Use latest
                </Button>
              </div>
            </form>
          </details>
          {view && <Badge status={view.status}>{view.status}</Badge>}
          <Button
            className="refresh-view"
            variant="ghost"
            size="sm"
            onPress={refresh}
          >
            <RefreshCw size={13} />
            Refresh
          </Button>
        </div>
        <main id="main-content" tabIndex={-1}>
          {children}
        </main>
        <footer className="app-footer">
          <span>valio.code</span>
          <span>
            {view ? `view ${view.id.slice(0, 20)}` : "No analysis view loaded"}
          </span>
          <span>AGPL-3.0</span>
        </footer>
      </div>
    </div>
  );
}
