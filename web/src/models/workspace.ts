export interface Workspace {
  id: string;
  name: string;
}
export interface Repository {
  id: string;
  workspaceId: string;
  remoteUrl: string;
}
export interface Project {
  id: string;
  workspaceId: string;
  key: string;
  name: string;
  description: string;
  kind: string;
  tags: string[];
  buildProfiles: string[];
  environmentProfiles: string[];
  status: string;
}
export interface SourceRoot {
  repositoryId: string;
  path: string;
  role: string;
  version: string;
  include?: string[];
  exclude?: string[];
  buildUnit?: string;
}
export interface Definition {
  project: Project;
  roots: SourceRoot[];
}
export interface FileRef {
  id: string;
  path: string;
  repositoryId: string;
  snapshotId: string;
  language: string;
  size: number;
  projectIds: string[];
}
export interface SourceFile extends FileRef {
  content: string;
}
export interface AnalysisView {
  id: string;
  workspaceId: string;
  snapshotId: string;
  files: FileRef[];
  repositories: {
    repositoryId: string;
    snapshotId: string;
    commit: string;
    manifestFingerprint: string;
  }[];
  projects: Definition[];
  mixed: boolean;
  status: string;
  projections: Record<string, string>;
  profile: string;
}
export interface Capabilities {
  features: { name: string; status: string }[];
  limits: { maxBodyBytes: number; maxFiles: number; maxScanBytes: number };
  authentication: string;
  buildProfileId: string;
}
