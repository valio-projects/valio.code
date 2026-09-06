import type { TypeCandidate } from "../../../models/types";
import { Badge } from "../../../components/Feedback";
import { Attributes } from "./Attributes";
import { factText, FactValue, Reference, TypeName } from "./FactValue";
import { Constants, Fields, Generics, Methods } from "./Members";
import { ArrayDetails, Layout } from "./Layout";
export function TypeCard({ candidate }: { candidate: TypeCandidate }) {
  const t = candidate.type;
  return (
    <article className="type-card">
      <header className="type-card-header">
        <div className="type-monogram">T</div>
        <div>
          <div className="eyebrow">
            {factText(t.language)} / {factText(t.kind)}
          </div>
          <h2>{factText(t.name)}</h2>
          <div className="qualified-name">
            <FactValue fact={t.fullyQualifiedName} />
          </div>
        </div>
        <Badge status={candidate.reference?.resolution}>
          {candidate.reference?.resolution ||
            t.symbol?.resolution ||
            "unresolved"}
        </Badge>
      </header>
      <div className="type-scope">
        <span>
          Project <code>{t.scope.projectId}</code>
        </span>
        <span>
          Profile <code>{t.scope.buildProfileId}</code>
        </span>
        <span>
          Version <code>{t.scope.versionId}</code>
        </span>
      </div>
      <div className="type-card-content">
        <div className="fact-strip">
          <span>
            Kind <FactValue fact={t.kind} />
          </span>
          <span>
            Visibility <FactValue fact={t.visibility} />
          </span>
          <span>
            Modifiers <FactValue fact={t.modifiers} />
          </span>
        </div>
        <Reference reference={candidate.reference || t.symbol} />
        <div className="count-grid">
          {Object.entries(candidate.counts || {}).map(([label, fact]) => (
            <div key={label}>
              <span>{label.replace(/([A-Z])/g, " $1")}</span>
              <strong>{Number.isFinite(fact) ? fact : "Unresolved"}</strong>
            </div>
          ))}
        </div>
        <p className="evidence-hint">
          Select any fact to inspect its evidence and scope. “Reported” counts
          describe the returned metadata only.
        </p>
        <Attributes items={t.attributes} />
        <Generics items={t.genericParameters} />
        {t.underlyingType && (
          <section className="type-section">
            <h3>Underlying type</h3>
            <TypeName type={t.underlyingType} />
            <Attributes items={t.underlyingType.attributes} />
          </section>
        )}
        {t.embeddedTypes?.length ? (
          <section className="type-section">
            <h3>Embedded types</h3>
            {t.embeddedTypes.map((e, i) => (
              <div key={i}>
                <TypeName type={e} />
                <Attributes items={e.attributes} />
              </div>
            ))}
          </section>
        ) : null}
        <Fields items={t.fields} />
        <Fields items={t.properties} title="Properties" />
        <Methods items={t.methods} title="Methods" />
        <Methods items={t.constructors} title="Constructors" />
        {t.enum && (
          <>
            <section className="type-section">
              <h3>Enum underlying type</h3>
              <TypeName type={t.enum.underlyingType} />
            </section>
            <Constants items={t.enum.members} title="Enum members" />
          </>
        )}
        <Constants items={t.constants} title="Typed constants" />
        <ArrayDetails array={t.array} />
        <Layout type={t} />
      </div>
    </article>
  );
}
