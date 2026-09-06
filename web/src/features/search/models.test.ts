import { expect, it } from "vitest";
import { searchRequest } from "./models";
it("serializes explicit workspace, project and immutable view scope", () => {
  expect(
    searchRequest("w", "p", "view-old", 'lang:go AND "é"', "regex", true, 50),
  ).toEqual({
    query: 'lang:go AND "é"',
    mode: "regex",
    scope: { workspaceId: "w", projectIds: ["p"], viewId: "view-old" },
    limit: 50,
    offset: 50,
    caseSensitive: true,
  });
});
it("all projects is an empty project list, never an invented identity", () => {
  expect(searchRequest("w", "", "v", "hello").scope).toEqual({
    workspaceId: "w",
    projectIds: [],
    viewId: "v",
  });
});
