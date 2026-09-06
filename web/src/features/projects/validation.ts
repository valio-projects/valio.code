import type { Definition } from "../../models/workspace";
export const validID = (value: string) =>
  /^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$/.test(value);
export const listValues = (value: string) => [
  ...new Set(
    value
      .split(",")
      .map((v) => v.trim())
      .filter(Boolean),
  ),
];
export function validRootPath(path: string): boolean {
  if (path === ".") return true;
  return (
    !!path &&
    !/[\\:\x00-\x1f<>"|?*]/.test(path) &&
    path
      .split("/")
      .every(
        (p) =>
          !!p.trim() &&
          p !== "." &&
          p !== ".." &&
          !/[. ]$/.test(p) &&
          !/^(CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])(?:\.|$)/i.test(p),
      )
  );
}
export function validateProject(d: Definition): string | undefined {
  if (!validID(d.project.id))
    return "Use a project ID with letters, digits, underscores, or hyphens (maximum 128 characters).";
  if (!/^[a-z0-9](?:[a-z0-9-]{0,78}[a-z0-9])?$/.test(d.project.key))
    return "Project key must be a lowercase slug, up to 80 characters, with no leading or trailing hyphen.";
  if (!d.project.name.trim()) return "Enter a project name.";
  if (!d.roots.length) return "Add at least one source root.";
  for (const root of d.roots) {
    if (!root.repositoryId) return "Choose a repository for every source root.";
    if (!validRootPath(root.path))
      return "Use a clean repository-relative path, such as src or ., without parent traversal.";
    if (!root.version.trim()) return "Enter a version for every source root.";
  }
}
export function validateRemote(remote: string): boolean {
  if (!remote) return true;
  try {
    const u = new URL(remote);
    return (
      ["https:", "ssh:"].includes(u.protocol) &&
      !!u.hostname &&
      !u.username &&
      !u.password &&
      !u.search &&
      !u.hash
    );
  } catch {
    return false;
  }
}
