import { useState, type FormEvent } from "react";
import { Button, Input } from "@heroui/react";
import { ArrowRight, LockKeyhole } from "lucide-react";
import { login } from "../../api/client";
import { ErrorState } from "../../components/Feedback";
export function LoginPage({ onSuccess }: { onSuccess: () => void }) {
  const [token, setToken] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<unknown>();
  async function submit(event: FormEvent) {
    event.preventDefault();
    setBusy(true);
    setError(undefined);
    const submitted = token;
    setToken("");
    try {
      await login(submitted);
      onSuccess();
    } catch (error) {
      setError(error);
    } finally {
      setBusy(false);
    }
  }
  return (
    <main className="auth-page">
      <div className="auth-story">
        <a className="brand" href="/">
          v<span>valio.code</span>
        </a>
        <div>
          <div className="eyebrow">SOURCE INTELLIGENCE WORKSPACE</div>
          <h1>
            Follow the code.
            <br />
            Know the evidence.
          </h1>
          <p>
            Explore projects, search immutable source snapshots, and inspect the
            facts behind your types.
          </p>
        </div>
        <span className="auth-caption">
          Local deployment · Workspace access
        </span>
      </div>
      <section className="auth-form">
        <div className="auth-box">
          <LockKeyhole size={28} />
          <h2>Open your workspace</h2>
          <p>
            Enter the bootstrap access token configured for this deployment.
          </p>
          <form onSubmit={submit}>
            <label htmlFor="access-token">Access token</label>
            <Input
              id="access-token"
              type="password"
              autoComplete="off"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              required
              autoFocus
              aria-describedby="token-help"
            />
            <small id="token-help">
              Your token is exchanged for a browser session and cleared from
              this form.
            </small>
            {error ? <ErrorState error={error} /> : null}
            <Button
              type="submit"
              isDisabled={!token.trim() || busy}
              isPending={busy}
              className="full"
            >
              Open workspace <ArrowRight size={16} />
            </Button>
          </form>
        </div>
      </section>
    </main>
  );
}
