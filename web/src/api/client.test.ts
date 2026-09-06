import { expect, it, vi } from "vitest";
import { ApiError, request } from "./client";
it("surfaces structured API errors and triggers session gate on 401", async () => {
  vi.spyOn(globalThis, "fetch").mockResolvedValue(
    new Response(
      JSON.stringify({
        error: { message: "Session expired" },
        requestId: "request-1",
      }),
      { status: 401 },
    ),
  );
  const listener = vi.fn();
  window.addEventListener("valio:unauthorized", listener);
  await expect(request("/workspace")).rejects.toMatchObject({
    status: 401,
    message: "Session expired",
    requestId: "request-1",
  });
  expect(listener).toHaveBeenCalledOnce();
  window.removeEventListener("valio:unauthorized", listener);
});
it("returns an actionable network error instead of fabricated data", async () => {
  vi.spyOn(globalThis, "fetch").mockRejectedValue(
    new TypeError("Failed to fetch"),
  );
  await expect(request("/workspace")).rejects.toBeInstanceOf(ApiError);
});
