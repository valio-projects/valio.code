import { useCapabilities } from "../../api/queries";
import { Badge, ErrorState, Loading } from "../../components/Feedback";
export function CapabilitiesPage() {
  const caps = useCapabilities();
  return (
    <>
      <div className="page-heading">
        <div>
          <div className="eyebrow">SYSTEM / ANALYSIS CONTRACT</div>
          <h1>
            Know what is known<span className="title-dot">.</span>
          </h1>
          <p>
            Capabilities reported by this deployment, including incomplete and
            unsupported analysis.
          </p>
        </div>
      </div>
      {caps.isPending ? (
        <Loading label="Loading capabilities…" />
      ) : caps.error ? (
        <ErrorState error={caps.error} retry={() => void caps.refetch()} />
      ) : (
        caps.data && (
          <>
            <section className="capabilities-panel">
              <div className="section-heading">
                <h2>Available analysis</h2>
                <Badge>{caps.data.buildProfileId}</Badge>
              </div>
              <div className="table-wrap">
                <table>
                  <thead>
                    <tr>
                      <th>Feature</th>
                      <th>Reported status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {caps.data.features?.map((f) => (
                      <tr key={f.name}>
                        <td>{f.name.replaceAll("_", " ")}</td>
                        <td>
                          <Badge status={f.status}>{f.status}</Badge>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </section>
            <div className="capability-notes">
              <section>
                <h3>Analysis boundaries</h3>
                <p>
                  AST facts describe source syntax. Compiler resolution,
                  references, physical layout, and SDK metadata require their
                  own producers and evidence.
                </p>
                <p>
                  A shared type model does not mean every language extractor is
                  available. Check the statuses above before interpreting absent
                  members.
                </p>
              </section>
              <section>
                <h3>Request limits</h3>
                <dl className="compact-dl">
                  {Object.entries(caps.data.limits || {}).map(
                    ([key, value]) => (
                      <div key={key}>
                        <dt>{key.replace(/([A-Z])/g, " $1")}</dt>
                        <dd>{value.toLocaleString()}</dd>
                      </div>
                    ),
                  )}
                  <div>
                    <dt>Authentication</dt>
                    <dd>{caps.data.authentication}</dd>
                  </div>
                </dl>
              </section>
            </div>
          </>
        )
      )}
    </>
  );
}
