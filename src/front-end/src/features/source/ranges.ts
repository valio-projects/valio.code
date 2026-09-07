import type { ByteRange } from "../search/models";
export interface SourceLine {
  number: number;
  text: string;
  highlighted: boolean;
}
/** API offsets are half-open UTF-8 bytes. Never apply byte positions as JS string offsets. */
export function sourceLines(
  content: string,
  ranges: ByteRange[] = [],
): SourceLine[] {
  const encoder = new TextEncoder();
  let start = 0;
  return content.split("\n").map((text, index) => {
    const end = start + encoder.encode(text).length;
    const highlighted = ranges.some(
      (r) =>
        Number.isFinite(r.start) &&
        Number.isFinite(r.end) &&
        r.start >= 0 &&
        r.end >= r.start &&
        ((r.start < end + 1 && r.end > start) ||
          (r.start === r.end && r.start >= start && r.start <= end)),
    );
    start = end + 1;
    return { number: index + 1, text, highlighted };
  });
}
