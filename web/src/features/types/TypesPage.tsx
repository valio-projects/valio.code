import { useEffect, useState, type FormEvent } from "react";
import { Button, Input } from "@heroui/react";
import { useQuery } from "@tanstack/react-query";
import { ScanSearch } from "lucide-react";
import { request } from "../../api/client";
import type { AnalysisView } from "../../models/workspace";
import type { TypeResult } from "../../models/types";
import { useNavigation } from "../../providers/NavigationProvider";
import { Empty, ErrorState, Loading, Badge } from "../../components/Feedback";
import { TypeCard } from "./components/TypeCard";
export function TypesPage({ view }: { view?: AnalysisView }) {
  const { state, navigate } = useNavigation();
  const [name, setName] = useState(state.type);
  const [profile, setProfile] = useState(state.profile);
  useEffect(() => {
    setName(state.type);
    setProfile(state.profile);
  }, [state.type, state.profile]);
  const params = new URLSearchParams({
    name: state.type,
    viewId: view?.id || "",
    ...(state.project ? { projectId: state.project } : {}),
    ...(state.profile ? { buildProfileId: state.profile } : {}),
  });
  const result = useQuery({
    queryKey: ["types", params.toString()],
    queryFn: ({ signal }) =>
      request<TypeResult>(`/types?${params}`, { signal }),
    enabled: !!state.type && !!view,
  });
  function submit(e: FormEvent) {
    e.preventDefault();
    navigate({ type: name.trim(), profile: profile.trim() });
  }
  return (
    <>
      <div className="page-heading">
        <div>
          <div className="eyebrow">ANALYSIS / TYPE METADATA</div>
          <h1>
            Type inspector<span className="title-dot">.</span>
          </h1>
          <p>
            Inspect structure, signatures, annotations, and the evidence behind
            each fact.
          </p>
        </div>
      </div>
      <form className="query-bar" onSubmit={submit}>
        <label className="search-input">
          <ScanSearch size={20} />
          <Input
            aria-label="Type name"
            placeholder="Type name or fully qualified name"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
        </label>
        <Input
          className="profile-input"
          aria-label="Build profile ID"
          placeholder="All build profiles"
          value={profile}
          onChange={(e) => setProfile(e.target.value)}
        />
        <Button
          type="submit"
          isDisabled={!view || !name.trim()}
          isPending={result.isFetching}
        >
          Inspect type
        </Button>
      </form>
      <p className="query-help">
        Names can resolve to several candidates across projects and build
        profiles. Each candidate keeps its own scope.
      </p>
      {!view ? (
        <Empty title="An analysis view is required">
          <p>Index and upload source before inspecting types.</p>
        </Empty>
      ) : !state.type ? (
        <Empty title="Start with a type name">
          <p>
            Search for a class, struct, interface, alias, or another named type.
            <br />
            Available facts depend on the language producer and selected view.
          </p>
        </Empty>
      ) : result.isPending ? (
        <Loading label="Resolving type candidates…" />
      ) : result.error ? (
        <ErrorState error={result.error} retry={() => void result.refetch()} />
      ) : (
        result.data && (
          <>
            <div className="results-heading">
              <strong>{result.data.candidates?.length || 0} candidates</strong>
              <Badge status={result.data.status}>
                {result.data.status.replace("_", " ")}
              </Badge>
              <span className="mono">
                view {result.data.viewId.slice(0, 16)}
              </span>
            </div>
            {result.data.status === "ambiguous" && (
              <p className="notice">
                Multiple candidates match. Select a project or build profile to
                narrow the scope; no candidate is silently chosen.
              </p>
            )}
            {result.data.candidates?.length ? (
              <div className="type-cards">
                {result.data.candidates.map((c) => (
                  <TypeCard
                    key={`${c.type.id}:${c.type.scope.projectId}:${c.type.scope.buildProfileId}`}
                    candidate={c}
                  />
                ))}
              </div>
            ) : (
              <Empty title="No type matches this scope">
                <p>
                  Check the name, project, build profile, and analysis
                  capabilities.
                </p>
              </Empty>
            )}
          </>
        )
      )}
    </>
  );
}
