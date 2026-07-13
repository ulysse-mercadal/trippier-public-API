import { createHmac } from 'crypto';

/**
 * Builds a replay-resistant X-Internal-Auth header value: "<ts>.<hmac-sha256(secret, ts)>".
 * @param secret shared secret used to sign the timestamp
 * @returns the header value; receiving services reject tokens older than ±30s
 */
export function buildInternalAuth(secret: string): string {
	const ts = String(Math.floor(Date.now() / 1000));
	const sig = createHmac('sha256', secret).update(ts).digest('hex');
	return `${ts}.${sig}`;
}
