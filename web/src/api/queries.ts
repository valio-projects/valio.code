import { useQuery } from "@tanstack/react-query";
import { request } from "./client";
import type {
  AnalysisView,
  Capabilities,
  Definition,
  Repository,
  Workspace,
} from "../models/workspace";
export const useWorkspace = () =>
  useQuery({
    queryKey: ["workspace"],
    queryFn: ({ signal }) => request<Workspace>("/workspace", { signal }),
  });
export const useProjects = () =>
  useQuery({
    queryKey: ["projects"],
    queryFn: ({ signal }) =>
      request<{ items: Definition[] }>("/projects", { signal }),
  });
export const useRepositories = () =>
  useQuery({
    queryKey: ["repositories"],
    queryFn: ({ signal }) =>
      request<{ items: Repository[] }>("/repositories", { signal }),
  });
export const useCapabilities = () =>
  useQuery({
    queryKey: ["capabilities"],
    queryFn: ({ signal }) => request<Capabilities>("/capabilities", { signal }),
  });
export const useView = (id: string) =>
  useQuery({
    queryKey: ["view", id],
    queryFn: ({ signal }) =>
      request<AnalysisView>(
        id ? `/views/${encodeURIComponent(id)}` : "/views",
        { signal },
      ),
  });
