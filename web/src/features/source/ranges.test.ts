import { expect, it } from "vitest";
import { sourceLines } from "./ranges";
it("maps UTF-8 byte offsets across Unicode and CRLF without JS slicing corruption", () => {
  const lines = sourceLines("🙂é\r\nneedle\nlast", [{ start: 8, end: 14 }]);
  expect(lines.map((l) => l.highlighted)).toEqual([false, true, false]);
  expect(lines[0].text).toBe("🙂é\r");
  expect(lines[1].text).toBe("needle");
});
it("handles multiline and zero-width matches and rejects reversed ranges", () => {
  expect(
    sourceLines("one\ntwo\nthree", [{ start: 2, end: 6 }]).map(
      (l) => l.highlighted,
    ),
  ).toEqual([true, true, false]);
  expect(
    sourceLines("one\ntwo", [{ start: 4, end: 4 }]).map((l) => l.highlighted),
  ).toEqual([false, true]);
  expect(sourceLines("one", [{ start: 3, end: 1 }])[0].highlighted).toBe(false);
});
