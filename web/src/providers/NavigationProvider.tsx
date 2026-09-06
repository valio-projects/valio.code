import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
export type Page =
  "workspace" | "projects" | "search" | "types" | "capabilities";
export interface LocationState {
  page: Page;
  project: string;
  view: string;
  q: string;
  type: string;
  file: string;
  mode: string;
  caseSensitive: boolean;
  profile: string;
}
const pages: Page[] = [
  "workspace",
  "projects",
  "search",
  "types",
  "capabilities",
];
function read(): LocationState {
  const p = new URLSearchParams(window.location.search);
  const page = p.get("page") as Page;
  return {
    page: pages.includes(page) ? page : "workspace",
    project: p.get("project") || "",
    view: p.get("view") || "",
    q: p.get("q") || "",
    type: p.get("type") || "",
    file: p.get("file") || "",
    mode: ["regex", "exact"].includes(p.get("mode") || "")
      ? p.get("mode")!
      : "substring",
    caseSensitive: p.get("case") === "true",
    profile: p.get("profile") || "",
  };
}
const Navigation = createContext<{
  state: LocationState;
  navigate: (update: Partial<LocationState>) => void;
}>(null!);
export function NavigationProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState(read);
  useEffect(() => {
    const listener = () => setState(read());
    window.addEventListener("popstate", listener);
    return () => window.removeEventListener("popstate", listener);
  }, []);
  function navigate(update: Partial<LocationState>) {
    const next = { ...read(), ...update };
    const params = new URLSearchParams();
    for (const [key, value] of Object.entries(next))
      if (value && value !== "workspace")
        params.set(key === "caseSensitive" ? "case" : key, String(value));
    window.history.pushState({}, "", `?${params}`);
    setState(next);
  }
  return (
    <Navigation.Provider value={{ state, navigate }}>
      {children}
    </Navigation.Provider>
  );
}
export const useNavigation = () => useContext(Navigation);
