import { useState, type FormEvent } from "react";
import { Button, Input } from "@heroui/react";
import { Plus, X } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import { request } from "../../api/client";
import type {
  Definition,
  Repository,
  SourceRoot,
} from "../../models/workspace";
import { ErrorState } from "../../components/Feedback";
import { listValues, validateProject } from "./validation";
const emptyRoot = (): SourceRoot => ({
  repositoryId: "",
  path: ".",
  role: "code",
  version: "v1",
});
export function ProjectForm({
  workspaceId,
  repositories,
  initial,
  onClose,
}: {
  workspaceId: string;
  repositories: Repository[];
  initial?: Definition;
  onClose: () => void;
}) {
  const client = useQueryClient();
  const [draft, setDraft] = useState<Definition>(() =>
    initial
      ? structuredClone(initial)
      : {
          project: {
            id: "",
            workspaceId,
            key: "",
            name: "",
            description: "",
            kind: "service",
            tags: [],
            buildProfiles: [],
            environmentProfiles: [],
            status: "active",
          },
          roots: [emptyRoot()],
        },
  );
  const [error, setError] = useState<unknown>();
  const [busy, setBusy] = useState(false);
  function project(key: string, value: string | string[]) {
    setDraft((d) => ({ ...d, project: { ...d.project, [key]: value } }));
  }
  function root(index: number, key: string, value: string | string[]) {
    setDraft((d) => ({
      ...d,
      roots: d.roots.map((r, i) => (i === index ? { ...r, [key]: value } : r)),
    }));
  }
  async function submit(e: FormEvent) {
    e.preventDefault();
    const problem = validateProject(draft);
    if (problem) {
      setError(new Error(problem));
      return;
    }
    setBusy(true);
    setError(undefined);
    try {
      await request(
        initial
          ? `/projects/${encodeURIComponent(initial.project.id)}`
          : "/projects",
        { method: initial ? "PUT" : "POST", body: JSON.stringify(draft) },
      );
      await client.invalidateQueries({ queryKey: ["projects"] });
      onClose();
    } catch (e) {
      setError(e);
    } finally {
      setBusy(false);
    }
  }
  return (
    <section className="editor-panel">
      <div className="section-heading">
        <h2>{initial ? "Edit project" : "Create project"}</h2>
        <Button variant="ghost" size="sm" onPress={onClose}>
          Cancel
        </Button>
      </div>
      <p>
        Define a logical boundary. Source roots can overlap and can span
        repositories.
      </p>
      <form onSubmit={submit}>
        <div className="form-grid">
          <label>
            Project ID
            <Input
              required
              value={draft.project.id}
              onChange={(e) => project("id", e.target.value)}
              disabled={!!initial}
              maxLength={128}
            />
          </label>
          <label>
            Stable key
            <Input
              required
              value={draft.project.key}
              onChange={(e) => project("key", e.target.value)}
              disabled={!!initial}
              placeholder="billing-service"
              maxLength={80}
            />
          </label>
          <label>
            Display name
            <Input
              required
              value={draft.project.name}
              onChange={(e) => project("name", e.target.value)}
              maxLength={256}
            />
          </label>
          <label>
            Kind
            <select
              value={draft.project.kind}
              onChange={(e) => project("kind", e.target.value)}
            >
              {["service", "library", "application", "tool"].map((k) => (
                <option key={k}>{k}</option>
              ))}
            </select>
          </label>
          <label className="span-two">
            Description
            <textarea
              value={draft.project.description}
              onChange={(e) => project("description", e.target.value)}
              rows={2}
              maxLength={8192}
            />
          </label>
          <label>
            Build profiles <span className="subtle">comma separated IDs</span>
            <Input
              defaultValue={draft.project.buildProfiles.join(", ")}
              onBlur={(e) =>
                project("buildProfiles", listValues(e.target.value))
              }
            />
          </label>
          <label>
            Environment profiles{" "}
            <span className="subtle">comma separated IDs</span>
            <Input
              defaultValue={draft.project.environmentProfiles.join(", ")}
              onBlur={(e) =>
                project("environmentProfiles", listValues(e.target.value))
              }
            />
          </label>
          <label>
            Tags
            <Input
              defaultValue={draft.project.tags.join(", ")}
              onBlur={(e) => project("tags", listValues(e.target.value))}
            />
          </label>
          <label>
            Status
            <select
              value={draft.project.status}
              onChange={(e) => project("status", e.target.value)}
            >
              <option>active</option>
              <option>archived</option>
            </select>
          </label>
        </div>
        <div className="section-heading">
          <h3>Source roots</h3>
          <Button
            variant="outline"
            size="sm"
            onPress={() =>
              setDraft((d) => ({ ...d, roots: [...d.roots, emptyRoot()] }))
            }
          >
            <Plus size={15} />
            Add root
          </Button>
        </div>
        {draft.roots.map((r, i) => (
          <fieldset className="root-editor" key={i}>
            <legend>Root {i + 1}</legend>
            <div className="form-grid">
              <label>
                Repository
                <select
                  required
                  value={r.repositoryId}
                  onChange={(e) => root(i, "repositoryId", e.target.value)}
                >
                  <option value="">Choose a registered repository</option>
                  {repositories.map((r) => (
                    <option key={r.id} value={r.id}>
                      {r.id}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Relative path
                <Input
                  required
                  value={r.path}
                  onChange={(e) => root(i, "path", e.target.value)}
                />
              </label>
              <label>
                Role
                <select
                  value={r.role}
                  onChange={(e) => root(i, "role", e.target.value)}
                >
                  {[
                    "code",
                    "tests",
                    "contracts",
                    "configuration",
                    "deployment",
                    "documentation",
                    "generated",
                  ].map((role) => (
                    <option key={role}>{role}</option>
                  ))}
                </select>
              </label>
              <label>
                Root definition version
                <Input
                  required
                  value={r.version}
                  onChange={(e) => root(i, "version", e.target.value)}
                />
              </label>
            </div>
            <details className="advanced-root">
              <summary>Include / exclude patterns and build unit</summary>
              <div className="form-grid">
                <label>
                  Include globs <span className="subtle">comma separated</span>
                  <Input
                    defaultValue={r.include?.join(", ")}
                    onBlur={(e) =>
                      root(i, "include", listValues(e.target.value))
                    }
                  />
                </label>
                <label>
                  Exclude globs <span className="subtle">comma separated</span>
                  <Input
                    defaultValue={r.exclude?.join(", ")}
                    onBlur={(e) =>
                      root(i, "exclude", listValues(e.target.value))
                    }
                  />
                </label>
                <label>
                  Build unit
                  <Input
                    value={r.buildUnit || ""}
                    onChange={(e) => root(i, "buildUnit", e.target.value)}
                  />
                </label>
              </div>
            </details>
            <Button
              variant="ghost"
              size="sm"
              isDisabled={draft.roots.length === 1}
              onPress={() =>
                setDraft((d) => ({
                  ...d,
                  roots: d.roots.filter((_, n) => n !== i),
                }))
              }
            >
              <X size={14} />
              Remove root
            </Button>
          </fieldset>
        ))}
        {error ? <ErrorState error={error} /> : null}
        <div className="form-actions">
          <small>
            New analysis views capture the saved definition. Existing views stay
            pinned.
          </small>
          <Button type="submit" isPending={busy} isDisabled={busy}>
            {initial ? "Save project" : "Create project"}
          </Button>
        </div>
      </form>
    </section>
  );
}
