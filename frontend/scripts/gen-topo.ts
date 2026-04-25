/**
 * gen-topo.ts — one-shot script that fetches real SRTM elevation data for the
 * Mont Blanc massif and generates topographic contour polygons with d3-contour.
 *
 * Usage:
 *   bun run scripts/gen-topo.ts
 *
 * Output:
 *   src/lib/topo-data.ts   (static TypeScript module, commit to repo)
 *
 * Data source:
 *   Open Topo Data — SRTM 30m (api.opentopodata.org, rate limit: 1 req/s, 100 pts/req)
 */

import { contours } from 'd3-contour';

const LAT_MAX = 46.18;  // N
const LAT_MIN = 45.50;  // S
const LON_MIN = 6.42;   // W
const LON_MAX = 7.32;   // E

const GRID_W = 40;
const GRID_H = 40;

const THRESHOLDS = Array.from({ length: 15 }, (_, i) => 600 + i * 300);

// ── Build query points (row 0 = north) ───────────────────────────────────────
const pts: { lat: number; lng: number }[] = [];
for (let row = 0; row < GRID_H; row++) {
  const lat = LAT_MAX - (row / (GRID_H - 1)) * (LAT_MAX - LAT_MIN);
  for (let col = 0; col < GRID_W; col++) {
    const lng = LON_MIN + (col / (GRID_W - 1)) * (LON_MAX - LON_MIN);
    pts.push({ lat, lng });
  }
}

// ── Fetch elevations in batches of 100 ───────────────────────────────────────
const BATCH = 100;
const elevations: number[] = new Array(pts.length).fill(0);

console.log(`Fetching SRTM elevation for ${pts.length} grid points (${pts.length / BATCH} requests)…`);

for (let i = 0; i < pts.length; i += BATCH) {
  const batch = pts.slice(i, i + BATCH);
  const locStr = batch.map(p => `${p.lat.toFixed(5)},${p.lng.toFixed(5)}`).join('|');

  const res = await fetch(`https://api.opentopodata.org/v1/srtm30m?locations=${locStr}`);
  if (!res.ok) {
    throw new Error(`API ${res.status}: ${await res.text()}`);
  }
  const json = (await res.json()) as { results: { elevation: number | null }[] };
  json.results.forEach((r, j) => {
    elevations[i + j] = r.elevation ?? 0;
  });

  const done = Math.min(i + BATCH, pts.length);
  process.stdout.write(`  ${done}/${pts.length} points\r`);

  if (done < pts.length) await Bun.sleep(1100); // stay under 1 req/s
}
process.stdout.write('\n');

// ── Generate contour polygons with d3-contour ────────────────────────────────
const gen = contours().size([GRID_W, GRID_H]).thresholds(THRESHOLDS);
const features = gen(elevations as unknown as ArrayLike<number>);

const result = features.map(f => ({
  elevation: f.value,
  polygons: f.coordinates.map(polygon =>
    polygon.map(ring =>
      ring.map(([col, row]) => [col / GRID_W, row / GRID_H] as [number, number])
    )
  ),
}));

console.log(`Generated ${result.length} contour levels:`);
result.forEach(c => {
  const pts = c.polygons.reduce((s, p) => s + p.reduce((s2, r) => s2 + r.length, 0), 0);
  console.log(`  ${c.elevation} m — ${c.polygons.length} polygon(s), ${pts} points`);
});

// ── Write TypeScript module ───────────────────────────────────────────────────
const src = `// Auto-generated — do not edit.
// SRTM 30m elevation data via api.opentopodata.org
// Region: Mont Blanc massif ${LAT_MIN}–${LAT_MAX}°N, ${LON_MIN}–${LON_MAX}°E
// Grid: ${GRID_W}×${GRID_H} | Thresholds: ${THRESHOLDS.join(', ')} m
// Run \`bun run scripts/gen-topo.ts\` to refresh.
// Generated: ${new Date().toISOString()}

export interface ContourLevel {
  elevation: number;
  /** GeoJSON-style polygon rings in normalised [0,1] canvas coords (x=E, y=S). */
  polygons: [number, number][][][];
}

export const MONT_BLANC_CONTOURS: ContourLevel[] = ${JSON.stringify(result, null, 2)};
`;

const outPath = new URL('../src/lib/topo-data.ts', import.meta.url).pathname;
await Bun.write(outPath, src);
console.log(`\nWrote ${outPath}`);
