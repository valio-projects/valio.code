import { Button } from "@heroui/react";
import { ArrowUpRight, Terminal } from "lucide-react";
import { useNavigation } from "../../providers/NavigationProvider";
export function Onboarding({ workspaceId }: { workspaceId: string }) {
  const { navigate } = useNavigation();
  return (
    <section className="onboarding">
      <div className="onboarding-copy">
        <span className="outlined-icon">
          <Terminal size={24} />
        </span>
        <div className="eyebrow">YOUR FIRST ANALYSIS VIEW</div>
        <h2>
          Bring your code
          <br />
          into focus.
        </h2>
        <p>
          Register your sources, define project boundaries, and upload a
          snapshot from your local worktree.
        </p>
        <Button onPress={() => navigate({ page: "projects" })}>
          Set up projects <ArrowUpRight size={16} />
        </Button>
      </div>
      <div className="onboarding-steps">
        <div>
          <span className="step-number">01</span>
          <section>
            <h3>Register a repository</h3>
            <p>
              Choose a stable source identity in Projects. The remote URL is
              optional for local code.
            </p>
          </section>
        </div>
        <div>
          <span className="step-number">02</span>
          <section>
            <h3>Define your projects</h3>
            <p>
              Add explicit source roots, roles, and build profiles. One
              repository can belong to several projects.
            </p>
          </section>
        </div>
        <div>
          <span className="step-number">03</span>
          <section>
            <h3>Index and upload a snapshot</h3>
            <p>
              Set <code>VALIO_API_TOKEN</code> in your terminal environment,
              then run the agent from this checkout. Replace the root and
              repository values.
            </p>
            <pre>
              <code>{`go run ./cmd/valio-agent index \
  --root /path/to/worktree \
  --server http://localhost:8080 \
  --workspace ${workspaceId} \
  --repository YOUR_REPOSITORY_ID`}</code>
            </pre>
            <small>
              Shell line continuations vary. Source filtering happens before
              upload. Refresh the view after the command completes.
            </small>
          </section>
        </div>
      </div>
    </section>
  );
}
