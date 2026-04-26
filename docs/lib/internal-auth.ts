import { createHmac } from 'crypto';
import { existsSync, readFileSync } from 'fs';
import { resolve } from 'path';

function loadEnv(): Record<string, string> {
  const envPath = resolve(process.cwd(), '../.env');
  if (!existsSync(envPath)) return {};
  const env: Record<string, string> = {};
  for (const line of readFileSync(envPath, 'utf-8').split('\n')) {
    const m = line.match(/^([A-Z_][A-Z0-9_]*)=([^#]*)/);
    if (m) env[m[1]] = m[2].trim();
  }
  return env;
}

const _env = loadEnv();

const _secret = process.env.INTERNAL_SECRET ?? _env.INTERNAL_SECRET;
if (!_secret) {
  throw new Error('INTERNAL_SECRET is not set — refusing to start without a configured secret');
}
export const INTERNAL_SECRET: string = _secret;

export const POI_URL =
  process.env.FRONTEND_POI_API_URL ?? _env.FRONTEND_POI_API_URL ?? 'http://localhost:8080';

export const ITIN_URL =
  process.env.FRONTEND_ITINERARY_API_URL ?? _env.FRONTEND_ITINERARY_API_URL ?? 'http://localhost:8000';

export function internalAuth(): string {
  const ts = String(Math.floor(Date.now() / 1000));
  const sig = createHmac('sha256', INTERNAL_SECRET).update(ts).digest('hex');
  return `${ts}.${sig}`;
}
