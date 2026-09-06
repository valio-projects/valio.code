import type {
  Fact,
  SymbolReference,
  TypeReference,
} from "../../../models/types";
import { Badge } from "../../../components/Feedback";
export function factText(fact?: Fact<unknown>): string {
  if (
    !fact ||
    fact.status !== "known" ||
    fact.value === undefined ||
    fact.value === null
  )
    return fact?.status === "unsupported" ? "Unsupported" : "Unresolved";
  if (Array.isArray(fact.value))
    return fact.value.length ? fact.value.join(", ") : "None";
  return String(fact.value);
}
export function FactValue({ fact }: { fact?: Fact<unknown> }) {
  const known =
    fact?.status === "known" && fact.value !== undefined && fact.value !== null;
  return (
    <details className={`fact ${known ? "known" : "unknown"}`}>
      <summary title="Show evidence and scope">
        {factText(fact)}
        {known ? (
          <span className="evidence-dot" aria-label="Evidence available" />
        ) : null}
      </summary>
      <div className="fact-detail">
        <strong>{fact?.status || "unresolved"}</strong>
        <p>
          {fact?.reason ||
            (known
              ? "Reported by the analysis producer."
              : "Not collected for this scope.")}
        </p>
        {fact?.scope && (
          <dl className="compact-dl">
            {Object.entries(fact.scope).map(([key, value]) => (
              <div key={key}>
                <dt>{key}</dt>
                <dd>{value}</dd>
              </div>
            ))}
          </dl>
        )}
        {fact?.evidence?.length ? (
          <ul>
            {fact.evidence.map((e) => (
              <li key={e.id}>
                <b>{e.producer}</b>
                <code>{e.id}</code>
                {e.range && <code>{JSON.stringify(e.range)}</code>}
              </li>
            ))}
          </ul>
        ) : (
          <small>No evidence supplied.</small>
        )}
      </div>
    </details>
  );
}
export function Reference({ reference }: { reference?: SymbolReference }) {
  return (
    <details className="reference">
      <summary>
        <Badge status={reference?.resolution}>
          {reference?.resolution || "unresolved"} reference
        </Badge>
      </summary>
      <p>{reference?.reason || "Reference identity in the selected scope."}</p>
      {reference?.exact && <code>{reference.exact.id}</code>}
      {reference?.candidates?.map((c) => (
        <p key={c.id}>
          <code>{c.id}</code> · {c.scope.projectId} · {c.scope.buildProfileId}
        </p>
      ))}
    </details>
  );
}
export function TypeName({ type }: { type?: TypeReference }) {
  return (
    <span className="type-reference">
      <FactValue fact={type?.name} />
      {type?.symbol?.resolution && (
        <span className="subtle"> · {type.symbol.resolution}</span>
      )}
    </span>
  );
}
