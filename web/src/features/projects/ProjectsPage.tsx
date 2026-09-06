import { useState } from "react";
import { Button } from "@heroui/react";
import { ArrowUpRight, FolderCode, GitBranch, Plus } from "lucide-react";
import { useProjects, useRepositories } from "../../api/queries";
import { Badge, Empty, ErrorState, Loading } from "../../components/Feedback";
import type { Definition, Workspace } from "../../models/workspace";
import { useNavigation } from "../../providers/NavigationProvider";
import { ProjectForm } from "./ProjectForm";
import { RepositoryForm } from "./RepositoryForm";
export function ProjectsPage({ workspace }: { workspace: Workspace }) {
  const projects = useProjects();
  const repositories = useRepositories();
  const { navigate } = useNavigation();
  const [editor, setEditor] = useState<"project" | "repository" | Definition>();
  return (
    <>
      <div className="page-heading">
        <div>
          <div className="eyebrow">WORKSPACE / CATALOG</div>
          <h1>
            Projects<span className="title-dot">.</span>
          </h1>
          <p>
            Logical boundaries for your code, independent of where it lives.
          </p>
        </div>
        <Button
          onPress={() => setEditor("project")}
          isDisabled={!repositories.data?.items?.length}
        >
          <Plus size={17} />
          Create project
        </Button>
      </div>
      {editor === "repository" && (
        <RepositoryForm
          workspaceId={workspace.id}
          onClose={() => setEditor(undefined)}
        />
      )}{" "}
      {editor && editor !== "repository" && (
        <ProjectForm
          workspaceId={workspace.id}
          repositories={repositories.data?.items || []}
          initial={typeof editor === "object" ? editor : undefined}
          onClose={() => setEditor(undefined)}
        />
      )}
      <div className="section-heading">
        <h2>Project catalog</h2>
        <span className="subtle">
          {projects.data
            ? `${projects.data.items?.length || 0} defined`
            : "Loading"}
        </span>
      </div>
      {projects.isPending ? (
        <Loading label="Loading projects…" />
      ) : projects.error ? (
        <ErrorState
          error={projects.error}
          retry={() => void projects.refetch()}
        />
      ) : !projects.data?.items?.length ? (
        <Empty title="Give your code a boundary">
          <p>
            Register a repository, then create a project with explicit source
            roots.
          </p>
          <Button variant="outline" onPress={() => setEditor("repository")}>
            Register repository
          </Button>
        </Empty>
      ) : (
        <div className="project-grid">
          {projects.data.items.map((d) => (
            <article className="project-card" key={d.project.id}>
              <div className="project-card-top">
                <span className="project-icon">
                  <FolderCode size={22} />
                </span>
                <Badge status={d.project.status}>{d.project.kind}</Badge>
              </div>
              <h3>{d.project.name}</h3>
              <code className="subtle">{d.project.key}</code>
              <p>{d.project.description || "No description added."}</p>
              <div className="project-roots">
                {d.roots.map((r, i) => (
                  <div key={i}>
                    <GitBranch size={13} />
                    <code>
                      {r.repositoryId} / {r.path}
                    </code>
                    <span>{r.role}</span>
                  </div>
                ))}
              </div>
              <footer>
                <Button variant="ghost" size="sm" onPress={() => setEditor(d)}>
                  Edit definition
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onPress={() =>
                    navigate({ page: "search", project: d.project.id })
                  }
                >
                  Explore <ArrowUpRight size={14} />
                </Button>
              </footer>
            </article>
          ))}
        </div>
      )}
      <div className="section-heading repositories-heading">
        <div>
          <h2>Repositories</h2>
          <p className="subtle">Registered source identities</p>
        </div>
        <Button
          variant="outline"
          size="sm"
          onPress={() => setEditor("repository")}
        >
          <Plus size={15} />
          Register repository
        </Button>
      </div>
      {repositories.isPending ? (
        <Loading label="Loading repositories…" />
      ) : repositories.error ? (
        <ErrorState
          error={repositories.error}
          retry={() => void repositories.refetch()}
        />
      ) : (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Repository ID</th>
                <th>Remote</th>
                <th>Workspace</th>
              </tr>
            </thead>
            <tbody>
              {repositories.data?.items?.length ? (
                repositories.data.items.map((r) => (
                  <tr key={r.id}>
                    <td>
                      <GitBranch size={14} /> <code>{r.id}</code>
                    </td>
                    <td>
                      {r.remoteUrl || (
                        <span className="subtle">
                          Local source · no remote URL
                        </span>
                      )}
                    </td>
                    <td className="mono">{r.workspaceId}</td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={3} className="subtle">
                    No repositories registered.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </>
  );
}
