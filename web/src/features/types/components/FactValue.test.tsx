import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { FactValue, factText } from "./FactValue";
import { Layout } from "./Layout";
import type { TypeDescriptor } from "../../../models/types";
describe("evidence-backed type facts", () => {
  it("preserves known zero and does not coerce absent facts to zero", () => {
    expect(factText({ status: "known", value: 0 })).toBe("0");
    expect(factText()).toBe("Unresolved");
    expect(factText({ status: "unresolved", reason: "No compiler" })).toBe(
      "Unresolved",
    );
    expect(factText({ status: "unsupported" })).toBe("Unsupported");
    expect(factText({ status: "known", value: [] })).toBe("None");
  });
  it("shows reason, producer and scope for inspection", () => {
    render(
      <FactValue
        fact={{
          status: "known",
          value: 42,
          evidence: [{ id: "fact-a", producer: "go/types" }],
          scope: {
            workspaceId: "w",
            projectId: "p",
            buildProfileId: "arm64",
            versionId: "snapshot-1",
          },
        }}
      />,
    );
    expect(screen.getByText("42")).toBeTruthy();
    expect(screen.getByText("go/types")).toBeTruthy();
    expect(screen.getByText("arm64")).toBeTruthy();
  });
  it("renders missing ABI and field offsets as unresolved", () => {
    const type = {
      scope: { buildProfileId: "syntax-default" },
      fields: [{ id: "f", name: { status: "known", value: "Field" } }],
    } as TypeDescriptor;
    render(<Layout type={type} />);
    expect(screen.getAllByText("Unresolved").length).toBe(11);
    expect(screen.queryByText("0")).toBeNull();
  });
});
