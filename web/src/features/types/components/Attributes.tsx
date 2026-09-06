import type { AttributeUse, AttributeValue } from "../../../models/types";
import { FactValue, Reference, TypeName } from "./FactValue";
function Argument({ value }: { value: AttributeValue }) {
  return (
    <div className="attribute-argument">
      <span className="mono">{value.kind}</span> <TypeName type={value.type} />
      {value.kind === "redacted" ? (
        <span>Redacted · {value.redactionReason}</span>
      ) : (
        <FactValue
          fact={value.kind === "literal" ? value.literal : value.expression}
        />
      )}
    </div>
  );
}
export function Attributes({ items }: { items?: AttributeUse[] }) {
  if (!items?.length) return null;
  return (
    <details className="member-detail">
      <summary>
        Attributes / annotations <span>{items.length} reported</span>
      </summary>
      <ul className="attribute-tree">
        {items.map((a) => (
          <li key={a.id}>
            <strong>
              <FactValue fact={a.name} />
            </strong>
            <Reference reference={a.attributeClass} />
            {a.positionalArguments?.map((v, i) => (
              <div key={i}>
                <span className="subtle">Argument {i + 1}</span>
                <Argument value={v} />
              </div>
            ))}
            {a.namedArguments?.map((v, i) => (
              <div key={i}>
                <FactValue fact={v.name} />
                <Argument value={v.value} />
              </div>
            ))}
          </li>
        ))}
      </ul>
    </details>
  );
}
