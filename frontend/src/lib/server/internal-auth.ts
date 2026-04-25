import { createHmac } from 'crypto';

/**
 * Build an X-Internal-Auth header value: "<ts>.<hmac-sha256(secret, ts)>".
 * The timestamp-bound HMAC prevents replaying a captured header.
 * Receiving services reject tokens older than ±30 s.
 */
export function buildInternalAuth(secret: string): string {
	const ts = String(Math.floor(Date.now() / 1000));
	const sig = createHmac('sha256', secret).update(ts).digest('hex');
	return `${ts}.${sig}`;
}
