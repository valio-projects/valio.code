export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
    public requestId?: string,
  ) {
    super(message);
  }
}
export async function request<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`/api/v1${path}`, {
      ...init,
      credentials: "include",
      headers: { "Content-Type": "application/json", ...init.headers },
    });
  } catch (error) {
    if (error instanceof DOMException && error.name === "AbortError")
      throw error;
    throw new ApiError(
      0,
      "Cannot reach the API. Check that the API and database services are running, then retry.",
    );
  }
  if (!response.ok) {
    if (response.status === 401 && path !== "/session")
      window.dispatchEvent(new Event("valio:unauthorized"));
    const body = await response.json().catch(() => ({}));
    throw new ApiError(
      response.status,
      body.error?.message || `The API returned HTTP ${response.status}.`,
      body.requestId,
    );
  }
  return response.status === 204 ? (undefined as T) : response.json();
}
export const post = <T>(path: string, value: unknown) =>
  request<T>(path, { method: "POST", body: JSON.stringify(value) });
export const login = (token: string) => post<void>("/session", { token });
export const logout = () => request<void>("/session", { method: "DELETE" });
