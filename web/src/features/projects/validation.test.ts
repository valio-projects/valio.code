import { expect, it } from "vitest";
import { validRootPath, validateRemote } from "./validation";
it("accepts explicit relative roots and rejects traversal and portable path hazards", () => {
  for (const p of [".", "src", "packages/shared"])
    expect(validRootPath(p)).toBe(true);
  for (const p of [
    "../src",
    "/src",
    "C:\\src",
    "src//lib",
    "src/../lib",
    "src/CON",
    "src.",
  ])
    expect(validRootPath(p)).toBe(false);
});
it("rejects remote URLs that embed credentials or secrets", () => {
  expect(validateRemote("https://host/repo.git")).toBe(true);
  expect(validateRemote("")).toBe(true);
  for (const url of [
    "https://user:password@host/repo",
    "https://host/repo?token=secret",
    "file:///root/repo",
  ])
    expect(validateRemote(url)).toBe(false);
});
