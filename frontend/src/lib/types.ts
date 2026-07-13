/** Represents an authenticated account. */
export interface User {
	id: string;
	email: string;
	verified: boolean;
	created_at: string;
}

/** Metadata describing an issued API key. */
export interface ApiKey {
	id: string;
	user_id: string;
	name: string;
	key_prefix: string;
	tokens_limit: number;
	tokens_reset_interval_secs: number;
	revoked: boolean;
	created_at: string;
	last_used_at: string | null;
}

/** API key metadata augmented with current usage/quota state. */
export interface ApiKeyWithUsage extends ApiKey {
	tokens_remaining: number;
	resets_in_secs: number;
}

/** Result returned when a new API key is created, including its secret value. */
export interface CreateKeyResult {
	key: string;
	metadata: ApiKey;
}
