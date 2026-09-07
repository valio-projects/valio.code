import type {
  Constant,
  Field,
  GenericParameter,
  Method,
  Parameter,
  Property,
} from "../../../models/types";
import { Attributes } from "./Attributes";
import { FactValue, TypeName } from "./FactValue";
export function Parameters({
  items,
  label,
}: {
  items?: Parameter[];
  label: string;
}) {
  return (
    <div className="parameter-group">
      <h5>{label}</h5>
      {items?.length ? (
        items.map((p) => (
          <div className="parameter" key={p.id}>
            <span>
              <FactValue fact={p.name} /> <TypeName type={p.type} />
            </span>
            <span>
              <FactValue fact={p.position} />
              {p.modifiers && <FactValue fact={p.modifiers} />}
            </span>
            {p.defaultValue && (
              <div>
                Default: <FactValue fact={p.defaultValue} />
              </div>
            )}
            <Attributes items={p.attributes} />
            <Attributes items={p.type.attributes} />
          </div>
        ))
      ) : (
        <p className="subtle">No {label.toLowerCase()} reported.</p>
      )}
    </div>
  );
}
export function Generics({ items }: { items?: GenericParameter[] }) {
  if (!items?.length) return null;
  return (
    <details className="member-detail">
      <summary>Generic parameters</summary>
      {items.map((g) => (
        <div key={g.id}>
          <FactValue fact={g.name} />
          <span> Constraints: </span>
          {g.constraints?.map((c, i) => (
            <TypeName key={i} type={c} />
          ))}
          <Attributes items={g.attributes} />
        </div>
      ))}
    </details>
  );
}
export function MethodRow({ method }: { method: Method }) {
  return (
    <details className="member-row">
      <summary>
        <span>
          <FactValue fact={method.name} />
        </span>
        <span className="member-signature">
          {method.signature?.status === "known"
            ? method.signature.value
            : "Signature unresolved"}
        </span>
      </summary>
      <div className="member-body">
        <div className="fact-strip">
          <span>
            Visibility <FactValue fact={method.visibility} />
          </span>
          <span>
            Modifiers <FactValue fact={method.modifiers} />
          </span>
          <span>
            Signature <FactValue fact={method.signature} />
          </span>
        </div>
        {method.receiver && (
          <Parameters items={[method.receiver]} label="Receiver" />
        )}
        <Parameters items={method.parameters} label="Parameters" />
        <Parameters items={method.returns} label="Returns" />
        <Generics items={method.genericParameters} />
        <Attributes items={method.attributes} />
      </div>
    </details>
  );
}
export function Fields({
  items,
  title = "Fields",
}: {
  items?: (Field | Property)[];
  title?: string;
}) {
  return (
    <section className="type-section">
      <h3>
        {title} <span>{items?.length || 0} reported</span>
      </h3>
      {items?.length ? (
        items.map((f) => (
          <details className="member-row" key={f.id}>
            <summary>
              <FactValue fact={f.name} />
              <TypeName type={f.type} />
            </summary>
            <div className="member-body">
              <div className="fact-strip">
                <span>
                  Visibility <FactValue fact={f.visibility} />
                </span>
                <span>
                  Modifiers <FactValue fact={f.modifiers} />
                </span>
                {f.tag && (
                  <span>
                    Struct tag <FactValue fact={f.tag} />
                  </span>
                )}
              </div>
              <Attributes items={f.attributes} />
              <Attributes items={f.type.attributes} />
              {"parameters" in f && (
                <Parameters items={f.parameters} label="Indexer parameters" />
              )}
              {"getter" in f && f.getter && <MethodRow method={f.getter} />}{" "}
              {"setter" in f && f.setter && <MethodRow method={f.setter} />}
            </div>
          </details>
        ))
      ) : (
        <p className="section-empty">
          No {title.toLowerCase()} reported. This is not a completeness claim.
        </p>
      )}
    </section>
  );
}
export function Methods({ items, title }: { items?: Method[]; title: string }) {
  return (
    <section className="type-section">
      <h3>
        {title} <span>{items?.length || 0} reported</span>
      </h3>
      {items?.length ? (
        items.map((m) => <MethodRow key={m.id} method={m} />)
      ) : (
        <p className="section-empty">
          No {title.toLowerCase()} reported. This is not a completeness claim.
        </p>
      )}
    </section>
  );
}
export function Constants({
  items,
  title,
}: {
  items?: Constant[];
  title: string;
}) {
  if (!items?.length) return null;
  return (
    <section className="type-section">
      <h3>{title}</h3>
      {items.map((c) => (
        <details className="member-row" key={c.id}>
          <summary>
            <FactValue fact={c.name} />
            <FactValue fact={c.value?.value} />
          </summary>
          <div className="member-body">
            <div className="fact-strip">
              <span>
                Expression <FactValue fact={c.expression} />
              </span>
              <span>
                Type <TypeName type={c.value?.type} />
              </span>
            </div>
            <Attributes items={c.attributes} />
            <h5>Reported usages</h5>
            {c.occurrences?.length ? (
              c.occurrences.map((o) => (
                <div className="occurrence" key={o.id}>
                  <FactValue fact={o.role} />
                  <code>{JSON.stringify(o.range)}</code>
                </div>
              ))
            ) : (
              <p className="subtle">
                No usages reported; reference completeness is not established.
              </p>
            )}
          </div>
        </details>
      ))}
    </section>
  );
}
