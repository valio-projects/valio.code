import { useState, type FormEvent } from "react";
import { Button, Input } from "@heroui/react";
import { useQueryClient } from "@tanstack/react-query";
import { post } from "../../api/client";
import { ErrorState } from "../../components/Feedback";
import { validID, validateRemote } from "./validation";
export function RepositoryForm({
  workspaceId,
  onClose,
}: {
  workspaceId: string;
  onClose: () => void;
}) {
  const client = useQueryClient();
  const [id, setID] = useState("");
  const [remote, setRemote] = useState("");
  const [error, setError] = useState<unknown>();
  const [busy, setBusy] = useState(false);
  async function submit(e: FormEvent) {
    e.preventDefault();
    if (!validID(id)) {
      setError(
        new Error(
          "Repository ID must use letters, digits, underscores, or hyphens (maximum 128 characters).",
        ),
      );
      return;
    }
    if (!validateRemote(remote)) {
      setError(
        new Error(
          "Use an HTTPS or SSH URL without credentials, query parameters, or fragments.",
        ),
      );
      return;
    }
    setBusy(true);
    setError(undefined);
    try {
      await post("/repositories", { id, workspaceId, remoteUrl: remote });
      await client.invalidateQueries({ queryKey: ["repositories"] });
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
        <h2>Register repository</h2>
        <Button variant="ghost" size="sm" onPress={onClose}>
          Cancel
        </Button>
      </div>
      <p>
        A repository identifies the source. Projects select roots within one or
        more repositories.
      </p>
      <form onSubmit={submit} className="form-grid">
        <label>
          Repository ID
          <Input
            value={id}
            onChange={(e) => setID(e.target.value)}
            placeholder="repo-main"
            required
            maxLength={128}
          />
        </label>
        <label>
          Remote URL <span className="subtle">optional</span>
          <Input
            value={remote}
            onChange={(e) => setRemote(e.target.value)}
            placeholder="https://host/organization/repository.git"
            maxLength={2048}
          />
        </label>
        {error ? <ErrorState error={error} /> : null}
        <div className="form-actions">
          <Button type="submit" isPending={busy} isDisabled={busy}>
            Register repository
          </Button>
        </div>
      </form>
    </section>
  );
}
