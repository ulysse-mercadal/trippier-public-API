<svelte:head><title>trippier/api · Dashboard</title></svelte:head>

<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth';
	import { listKeys, createKey, revokeKey, getMe, ApiError } from '$lib/api';
	import { t, initLocale } from '$lib/i18n';
	import type { ApiKeyWithUsage } from '$lib/types';

	let keys: ApiKeyWithUsage[] = [];
	let newKeyName  = '';
	let revealedKey = '';
	let error       = '';
	let creating    = false;
	let loaded      = false;
	let revokeTarget: ApiKeyWithUsage | null = null;
	let revoking = false;

	$: activeKeys  = keys.filter(k => !k.revoked);
	$: revokedKeys = keys.filter(k => k.revoked);
	$: globalKey   = activeKeys[0] ?? null;

	onMount(async () => {
		initLocale();
		const stored = auth.getStoredToken();
		if (!stored) { goto('/login'); return; }
		if (!$auth.user) {
			try { auth.init(stored, await getMe(stored)); }
			catch { goto('/login'); return; }
		}
		await loadKeys();
		loaded = true;
	});

	/**
	 * Fetches the current API keys and updates local state.
	 */
	async function loadKeys() {
		const token = $auth.token || auth.getStoredToken();
		try { keys = await listKeys(token); }
		catch (e) { error = e instanceof ApiError ? e.message : 'Loading error'; }
	}

	/**
	 * Creates a new API key from the entered name and reveals it.
	 */
	async function handleCreate() {
		if (!newKeyName.trim()) return;
		creating = true; error = '';
		const token = $auth.token || auth.getStoredToken();
		try {
			const result = await createKey(token, newKeyName.trim());
			revealedKey = result.key;
			newKeyName  = '';
			await loadKeys();
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Creation error';
		} finally { creating = false; }
	}

	/**
	 * Revokes the key currently targeted for revocation and refreshes the list.
	 */
	async function confirmRevoke() {
		if (!revokeTarget) return;
		revoking = true;
		const token = $auth.token || auth.getStoredToken();
		try {
			await revokeKey(token, revokeTarget.id);
			revokeTarget = null;
			await loadKeys();
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Revoke error';
		} finally { revoking = false; }
	}

	/**
	 * Computes the percentage of tokens used by a key.
	 * @param k the API key with usage data
	 * @returns percentage of tokens used, clamped to 0-100
	 */
	function usedPct(k: ApiKeyWithUsage): number {
		if (k.tokens_limit <= 0) return 0;
		return Math.max(0, Math.min(100, ((k.tokens_limit - k.tokens_remaining) / k.tokens_limit) * 100));
	}

	/**
	 * Formats a duration in seconds as a short reset label.
	 * @param secs seconds until reset
	 * @returns formatted label in minutes or hours
	 */
	function resetLabel(secs: number): string {
		const m = Math.ceil(secs / 60);
		return m > 60 ? `${Math.ceil(m / 60)}h` : `${m}m`;
	}

	/**
	 * Copies the revealed API key to the clipboard.
	 */
	function copyKey() { navigator.clipboard.writeText(revealedKey); }
</script>

<div class="db">

	<div class="db-head wrap">
		<div>
			<p class="eyebrow">dashboard</p>
			<h1>{$t('db_title')}</h1>
			{#if $auth.user}<p class="db-email">{$auth.user.email}</p>{/if}
		</div>
		<a href="/" class="btn-ghost db-back">{$t('db_back')}</a>
	</div>

	<div class="db-body wrap">

		{#if error}<p class="db-error">{error}</p>{/if}

		{#if revealedKey}
			<div class="db-reveal">
				<div class="db-reveal-top">
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2 4 6v6c0 5 3.5 8.5 8 10 4.5-1.5 8-5 8-10V6l-8-4z"/><path d="m9 12 2 2 4-4"/></svg>
					<strong>{$t('db_copy_warning')}</strong>
					<button class="db-dismiss" on:click={() => (revealedKey = '')}>✕</button>
				</div>
				<div class="db-reveal-key">
					<code>{revealedKey}</code>
					<button class="db-copy-btn" on:click={copyKey}>{$t('db_copy_btn')}</button>
				</div>
			</div>
		{/if}

		{#if globalKey}
			{@const pct = globalKey.tokens_limit > 0 ? Math.max(0, Math.min(100, (globalKey.tokens_remaining / globalKey.tokens_limit) * 100)) : 0}
			{@const variant = pct > 25 ? 'ok' : pct > 10 ? 'warn' : 'err'}
			<div class="db-token-card">
				<div class="db-token-head">
					<div>
						<p class="db-token-label">{$t('db_tokens_remaining')}</p>
						<p class="db-token-count">
							<strong>{globalKey.tokens_remaining}</strong>
							<span>/ {globalKey.tokens_limit}</span>
						</p>
					</div>
					<div class="db-token-reset">
						<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
						{$t('db_reset_in')} {resetLabel(globalKey.resets_in_secs)}
					</div>
				</div>
				<div class="db-track">
					<div class="db-fill db-fill-{variant}" style="width:{pct}%"></div>
				</div>
			</div>
		{/if}

		<div class="db-section">
			<h2>{$t('db_section_title')}</h2>
			<p class="db-hint">1 000 {$t('db_hint')} <code>Authorization: Bearer</code>.</p>
			<div class="db-create">
				<input
					class="db-input"
					placeholder={$t('db_placeholder')}
					bind:value={newKeyName}
					on:keydown={(e) => e.key === 'Enter' && handleCreate()}
				/>
				<button class="db-btn-primary" disabled={creating || !newKeyName.trim()} on:click={handleCreate}>
					{creating ? '…' : $t('db_new_key')}
				</button>
			</div>
		</div>

		{#if !loaded}
			<p class="db-loading">{$t('db_loading')}</p>
		{:else if activeKeys.length === 0 && revokedKeys.length === 0}
			<div class="db-empty">
				<p>{$t('db_empty1')}</p>
				<p>{$t('db_empty2')}</p>
			</div>
		{:else}
			<ul class="db-keys">
				{#each activeKeys as k (k.id)}
					{@const used = k.tokens_limit - k.tokens_remaining}
					{@const pct  = usedPct(k)}
					{@const variant = pct < 75 ? 'ok' : pct < 90 ? 'warn' : 'err'}
					<li class="db-key">
						<div class="db-key-top">
							<div class="db-key-info">
								<span class="db-key-name">{k.name}</span>
								<code class="db-key-prefix">{k.key_prefix}…</code>
							</div>
							<button class="db-revoke-btn" on:click={() => (revokeTarget = k)}>{$t('db_revoke_btn')}</button>
						</div>
						<div class="db-key-usage">
							<div class="db-key-usage-meta">
								<span>{used} {$t('db_tokens_used')}</span>
								<span>{k.tokens_remaining} {$t('db_remaining')}</span>
							</div>
							<div class="db-track">
								<div class="db-fill db-fill-{variant}" style="width:{pct}%"></div>
							</div>
						</div>
					</li>
				{/each}

				{#each revokedKeys as k (k.id)}
					<li class="db-key db-key-revoked">
						<div class="db-key-top">
							<div class="db-key-info">
								<span class="db-key-name">{k.name}</span>
								<code class="db-key-prefix">{k.key_prefix}…</code>
							</div>
							<span class="db-revoked-badge">{$t('db_revoked_badge')}</span>
						</div>
					</li>
				{/each}
			</ul>
		{/if}

	</div>
</div>

{#if revokeTarget}
	<div class="db-modal-backdrop" on:click={() => !revoking && (revokeTarget = null)} role="presentation">
		<div class="db-modal" on:click|stopPropagation role="dialog" aria-modal="true">
			<div class="db-modal-head">
				<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
				<h3>{$t('db_modal_title')}</h3>
			</div>
			<p class="db-modal-body">
				{$t('db_modal_body_pre')} <strong>{revokeTarget.name}</strong> (<code>{revokeTarget.key_prefix}…</code>) {$t('db_modal_body_post')}
			</p>
			<div class="db-modal-foot">
				<button class="db-modal-cancel" on:click={() => (revokeTarget = null)} disabled={revoking}>{$t('db_modal_cancel')}</button>
				<button class="db-modal-confirm" on:click={confirmRevoke} disabled={revoking}>
					{revoking ? '…' : $t('db_modal_confirm')}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.db { min-height: 100vh; }
	.db-head {
		padding-top: 48px;
		padding-bottom: 32px;
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 16px;
		border-bottom: 1px solid var(--border);
	}
	.db-head h1 { font-size: 28px; font-weight: 600; letter-spacing: -0.02em; margin: 4px 0 0; }
	.eyebrow { font-family: var(--font-mono); font-size: 11px; letter-spacing: 0.07em; color: var(--accent); text-transform: lowercase; }
	.db-email { font-family: var(--font-mono); font-size: 12px; color: var(--text-3); margin-top: 4px; }
	.db-back {
		font-family: var(--font-mono);
		font-size: 12px;
		padding: 7px 14px;
		border-radius: var(--r-md);
		border: 1px solid var(--border);
		background: none;
		color: var(--text-3);
		text-decoration: none;
		transition: color .12s ease, border-color .12s ease;
		margin-top: 8px;
		white-space: nowrap;
	}
	.db-back:hover { color: var(--text); border-color: var(--text-3); }

	.db-body { padding-top: 32px; padding-bottom: 80px; display: flex; flex-direction: column; gap: 24px; }

	.db-error {
		font-size: 13px;
		color: oklch(72% 0.16 25);
		background: oklch(72% 0.16 25 / 0.08);
		border: 1px solid oklch(72% 0.16 25 / 0.25);
		border-radius: var(--r-md);
		padding: 10px 14px;
	}

	.db-reveal {
		border: 1px solid color-mix(in oklch, var(--accent) 30%, var(--border));
		border-radius: var(--r-md);
		background: color-mix(in oklch, var(--accent) 5%, transparent);
		padding: 14px 16px;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.db-reveal-top {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 13px;
	}
	.db-reveal-top svg { color: var(--accent); flex-shrink: 0; }
	.db-reveal-top strong { color: var(--accent); flex: 1; font-weight: 500; }
	.db-dismiss { background: none; border: none; color: var(--text-3); cursor: pointer; font-size: 12px; padding: 0; }
	.db-dismiss:hover { color: var(--text); }
	.db-reveal-key {
		display: flex;
		align-items: center;
		gap: 10px;
		background: var(--bg);
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		padding: 10px 14px;
	}
	.db-reveal-key code { flex: 1; font-family: var(--font-mono); font-size: 12px; color: var(--text); word-break: break-all; }
	.db-copy-btn {
		font-family: var(--font-mono);
		font-size: 11px;
		padding: 4px 10px;
		border-radius: 5px;
		border: 1px solid var(--border);
		background: var(--surface);
		color: var(--text-3);
		cursor: pointer;
		flex-shrink: 0;
		transition: color .12s, border-color .12s;
	}
	.db-copy-btn:hover { color: var(--accent); border-color: var(--accent); }

	.db-token-card {
		background: var(--bg-2);
		border: 1px solid var(--border);
		border-radius: var(--r-lg);
		padding: 20px 22px;
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.db-token-head { display: flex; align-items: flex-end; justify-content: space-between; }
	.db-token-label { font-family: var(--font-mono); font-size: 10.5px; text-transform: uppercase; letter-spacing: 0.07em; color: var(--text-3); margin-bottom: 4px; }
	.db-token-count { font-size: 28px; font-weight: 700; letter-spacing: -0.02em; line-height: 1; }
	.db-token-count strong { color: var(--text); }
	.db-token-count span { font-size: 16px; font-weight: 400; color: var(--text-3); margin-left: 4px; }
	.db-token-reset {
		display: inline-flex;
		align-items: center;
		gap: 5px;
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--text-3);
		margin-bottom: 4px;
	}

	.db-track { background: var(--surface); border-radius: 999px; height: 6px; overflow: hidden; }
	.db-fill { height: 100%; border-radius: 999px; transition: width .4s ease; }
	.db-fill-ok  { background: var(--accent); }
	.db-fill-warn { background: oklch(82% 0.12 80); }
	.db-fill-err  { background: oklch(72% 0.16 25); }

	.db-section { display: flex; flex-direction: column; gap: 10px; }
	.db-section h2 { font-size: 16px; font-weight: 600; letter-spacing: -0.01em; }
	.db-hint { font-size: 13px; color: var(--text-3); line-height: 1.55; }
	.db-hint code { font-family: var(--font-mono); font-size: 11.5px; color: var(--accent); }
	.db-create { display: flex; gap: 8px; }
	.db-input {
		flex: 1;
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		color: var(--text);
		padding: 9px 12px;
		font-family: var(--font-sans);
		font-size: 13.5px;
		outline: none;
		transition: border-color .12s;
	}
	.db-input:focus { border-color: var(--accent); }
	.db-btn-primary {
		padding: 9px 16px;
		background: var(--accent);
		color: #08120e;
		border: none;
		border-radius: var(--r-md);
		font-family: var(--font-sans);
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
		white-space: nowrap;
		transition: filter .12s;
	}
	.db-btn-primary:hover:not(:disabled) { filter: brightness(1.08); }
	.db-btn-primary:disabled { opacity: 0.45; cursor: not-allowed; }

	.db-keys { list-style: none; display: flex; flex-direction: column; gap: 10px; }
	.db-key {
		background: var(--bg-2);
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		padding: 16px 18px;
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.db-key-revoked { opacity: 0.4; }
	.db-key-top { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
	.db-key-info { display: flex; align-items: center; gap: 10px; }
	.db-key-name { font-size: 14px; font-weight: 600; color: var(--text); }
	.db-key-prefix { font-family: var(--font-mono); font-size: 11.5px; color: var(--text-3); }
	.db-revoke-btn {
		font-family: var(--font-mono);
		font-size: 11px;
		padding: 4px 10px;
		border-radius: 5px;
		border: 1px solid oklch(72% 0.16 25 / 0.3);
		background: none;
		color: oklch(72% 0.16 25);
		cursor: pointer;
		transition: background .12s, border-color .12s;
		white-space: nowrap;
	}
	.db-revoke-btn:hover { background: oklch(72% 0.16 25 / 0.08); border-color: oklch(72% 0.16 25 / 0.6); }
	.db-key-usage { display: flex; flex-direction: column; gap: 6px; }
	.db-key-usage-meta { display: flex; justify-content: space-between; font-family: var(--font-mono); font-size: 11px; color: var(--text-3); }
	.db-revoked-badge {
		font-family: var(--font-mono);
		font-size: 10.5px;
		color: var(--text-3);
		border: 1px solid var(--border);
		border-radius: 20px;
		padding: 2px 8px;
	}
	.db-loading { font-family: var(--font-mono); font-size: 12px; color: var(--text-3); }
	.db-empty {
		text-align: center;
		padding: 40px 24px;
		border: 1px dashed var(--border);
		border-radius: var(--r-md);
		font-size: 13.5px;
		color: var(--text-3);
		line-height: 1.8;
	}

	.db-modal-backdrop {
		position: fixed;
		inset: 0;
		background: oklch(0% 0 0 / 0.6);
		z-index: 200;
		display: flex;
		align-items: center;
		justify-content: center;
		backdrop-filter: blur(4px);
	}
	.db-modal {
		background: var(--bg);
		border: 1px solid var(--border);
		border-radius: var(--r-lg);
		width: min(440px, calc(100vw - 32px));
		padding: 24px;
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.db-modal-head {
		display: flex;
		align-items: center;
		gap: 10px;
		color: oklch(82% 0.12 80);
	}
	.db-modal-head h3 { font-size: 15px; font-weight: 600; color: var(--text); margin: 0; }
	.db-modal-body { font-size: 13.5px; color: var(--text-2); line-height: 1.6; margin: 0; }
	.db-modal-body strong { color: var(--text); font-weight: 500; }
	.db-modal-body code { font-family: var(--font-mono); font-size: 12px; color: var(--text-3); }
	.db-modal-foot { display: flex; justify-content: flex-end; gap: 8px; padding-top: 4px; }
	.db-modal-cancel {
		font-family: var(--font-mono);
		font-size: 12.5px;
		padding: 7px 14px;
		border-radius: var(--r-md);
		border: 1px solid var(--border);
		background: none;
		color: var(--text-3);
		cursor: pointer;
	}
	.db-modal-cancel:hover { color: var(--text); }
	.db-modal-confirm {
		font-family: var(--font-mono);
		font-size: 12.5px;
		font-weight: 600;
		padding: 7px 16px;
		border-radius: var(--r-md);
		border: none;
		background: oklch(72% 0.16 25);
		color: #fff;
		cursor: pointer;
		transition: filter .12s;
	}
	.db-modal-confirm:hover:not(:disabled) { filter: brightness(1.12); }
	.db-modal-confirm:disabled { opacity: 0.45; cursor: not-allowed; }
</style>
