import { Button } from "@heroui/react";
import {
  ArrowRight,
  Braces,
  FileCode2,
  FolderCode,
  GitBranch,
  Layers3,
} from "lucide-react";
import type { AnalysisView, Workspace } from "../../models/workspace";
import { useProjects, useRepositories } from "../../api/queries";
import { Badge, ErrorState } from "../../components/Feedback";
import { useNavigation } from "../../providers/NavigationProvider";
import { Onboarding } from "./Onboarding";
export function WorkspacePage({
  workspace,
  view,
}: {
  workspace: Workspace;
  view?: AnalysisView;
}) {
  const projects = useProjects();
  const repositories = useRepositories();
  const { navigate } = useNavigation();
  const stats = [
    {
      label: "Projects",
      value: projects.data?.items?.length,
      icon: FolderCode,
      note: "Defined boundaries",
    },
    {
      label: "Repositories",
      value: repositories.data?.items?.length,
      icon: GitBranch,
      note: "Registered sources",
    },
    {
      label: "Source files",
      value: view?.files?.length,
      icon: FileCode2,
      note: view ? "In this view" : "No published view",
    },
    {
      label: "Analysis view",
      value: view ? view.status : undefined,
      icon: Layers3,
      note: view ? "Snapshot pinned" : "Awaiting first upload",
    },
  ];
  return (
    <>
      <div className="page-heading">
        <div>
          <div className="eyebrow">WORKSPACE / OVERVIEW</div>
          <h1>
            Your code, connected<span className="title-dot">.</span>
          </h1>
          <p>
            {workspace.name} <span className="inline-divider">/</span> A shared
            place to explore source and inspect its structure.
          </p>
        </div>
        {view && (
          <Button onPress={() => navigate({ page: "search" })}>
            Explore source <ArrowRight size={16} />
          </Button>
        )}
      </div>
      <div className="stats-grid">
        {stats.map((s) => (
          <article className="stat-card" key={s.label}>
            <div>
              <span>{s.label}</span>
              <s.icon size={17} />
            </div>
            <strong>{s.value ?? "—"}</strong>
            <small>{s.note}</small>
          </article>
        ))}
      </div>
      {projects.error && (
        <ErrorState
          error={projects.error}
          retry={() => void projects.refetch()}
        />
      )}{" "}
      {repositories.error && (
        <ErrorState
          error={repositories.error}
          retry={() => void repositories.refetch()}
        />
      )}{" "}
      {!view ? (
        <Onboarding workspaceId={workspace.id} />
      ) : (
        <>
          <section className="view-panel">
            <div className="section-heading">
              <div>
                <div className="eyebrow">CURRENT ANALYSIS</div>
                <h2>One view. Concrete evidence.</h2>
              </div>
              <Badge status={view.status}>{view.status}</Badge>
            </div>
            <div className="view-details">
              <div>
                <label>View ID</label>
                <code>{view.id}</code>
              </div>
              <div>
                <label>Snapshot</label>
                <code>{view.snapshotId}</code>
              </div>
              <div>
                <label>Analysis profile</label>
                <code>{view.profile}</code>
              </div>
              <div>
                <label>Composition</label>
                <span>
                  {view.mixed
                    ? "Mixed repository snapshots"
                    : "Single snapshot"}
                </span>
              </div>
            </div>
            <div className="projection-grid">
              {Object.entries(view.projections || {}).map(([name, status]) => (
                <div key={name}>
                  <span>{name.replaceAll("_", " ")}</span>
                  <Badge status={status}>{status}</Badge>
                </div>
              ))}
            </div>
            <p className="section-description">
              Published views preserve project definitions and source
              membership. Project edits appear in a future view.
            </p>
          </section>
          <div className="workspace-actions">
            <button onClick={() => navigate({ page: "search" })}>
              <FileCode2 size={23} />
              <h3>Search the source</h3>
              <p>Query code, paths, and symbol names within a pinned view.</p>
              <ArrowRight size={18} />
            </button>
            <button onClick={() => navigate({ page: "types" })}>
              <Braces size={23} />
              <h3>Inspect a type</h3>
              <p>
                Read member signatures, annotations, and profile-specific facts.
              </p>
              <ArrowRight size={18} />
            </button>
          </div>
        </>
      )}
      <div className="workspace-footnote">
        <span className="tiny-dot" />
        Evidence is scoped. Unknown facts stay unknown.
      </div>
    </>
  );
}
