import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import { LoginPage } from "./LoginPage";
it("exchanges token in POST body, clears field, and never persists browser credentials", async () => {
  const fetchMock = vi
    .spyOn(globalThis, "fetch")
    .mockResolvedValue(new Response(null, { status: 204 }));
  const store = vi.spyOn(Storage.prototype, "setItem");
  const onSuccess = vi.fn();
  render(<LoginPage onSuccess={onSuccess} />);
  const input = screen.getByLabelText("Access token") as HTMLInputElement;
  fireEvent.change(input, { target: { value: "test-secret-123" } });
  fireEvent.submit(input.closest("form")!);
  expect(input.value).toBe("");
  await waitFor(() => expect(onSuccess).toHaveBeenCalledOnce());
  expect(store).not.toHaveBeenCalled();
  expect(fetchMock).toHaveBeenCalledWith(
    "/api/v1/session",
    expect.objectContaining({
      credentials: "include",
      method: "POST",
      body: JSON.stringify({ token: "test-secret-123" }),
    }),
  );
  expect(window.location.href).not.toContain("test-secret");
});
