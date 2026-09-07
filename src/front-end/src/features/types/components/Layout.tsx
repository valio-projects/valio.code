import type { ArrayShape, TypeDescriptor } from "../../../models/types";
import { FactValue, TypeName } from "./FactValue";
export function ArrayDetails({ array }: { array?: ArrayShape }) {
  if (!array) return null;
  return (
    <section className="type-section">
      <h3>Array shape</h3>
      <div className="fact-strip">
        <span>
          Element <TypeName type={array.elementType} />
        </span>
        <span>
          Rank <FactValue fact={array.rank} />
        </span>
      </div>
      {array.dimensions?.map((d, i) => (
        <div className="dimension" key={i}>
          <span>Dimension {i + 1}</span>
          <span>
            Length <FactValue fact={d.length} />
          </span>
          <span>
            Lower bound <FactValue fact={d.lowerBound} />
          </span>
        </div>
      ))}
    </section>
  );
}
export function Layout({ type }: { type: TypeDescriptor }) {
  const layout = type.layout;
  const facts = [
    ["Architecture", layout?.targetArchitecture],
    ["ABI", layout?.abi],
    ["Compiler", layout?.compiler],
    ["Compiler version", layout?.compilerVersion],
    ["Layout kind", layout?.kind],
    ["Size · bytes", layout?.sizeBytes],
    ["Alignment · bytes", layout?.alignmentBytes],
    ["Packing · bytes", layout?.packingBytes],
  ] as const;
  return (
    <section className="type-section">
      <h3>
        Memory layout <span>{type.scope.buildProfileId}</span>
      </h3>
      <p className="section-description">
        Physical facts belong to this target and build profile. Missing values
        remain unresolved.
      </p>
      <div className="layout-grid">
        {facts.map(([label, fact]) => (
          <div key={label}>
            <label>{label}</label>
            <FactValue fact={fact} />
          </div>
        ))}
      </div>
      {layout?.unknownReasons?.length ? (
        <p className="notice">{layout.unknownReasons.join(" · ")}</p>
      ) : null}
      <h4>Field offsets</h4>
      {type.fields?.length ? (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Field</th>
                <th>Offset · bytes</th>
                <th>Bit offset</th>
                <th>Bit width</th>
              </tr>
            </thead>
            <tbody>
              {type.fields.map((f) => {
                const l = layout?.fields?.find((x) => x.fieldId === f.id);
                return (
                  <tr key={f.id}>
                    <td>
                      <FactValue fact={f.name} />
                    </td>
                    <td>
                      <FactValue fact={l?.offsetBytes} />
                    </td>
                    <td>
                      <FactValue fact={l?.bitOffset} />
                    </td>
                    <td>
                      <FactValue fact={l?.bitWidth} />
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      ) : (
        <p className="section-empty">No field layout entries reported.</p>
      )}
    </section>
  );
}
