<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth';
	import { listKeys, createKey, revokeKey, getMe, ApiError } from '$lib/api';
	import type { ApiKeyWithUsage } from '$lib/types';
	import TokenBar from '$lib/components/TokenBar.svelte';

	let keys: ApiKeyWithUsage[] = [];
	let newKeyName = '';
	let revealedKey = '';
	let error = '';
	let creating = false;
	let loaded = false;

	onMount(async () => {
		const stored = auth.getStoredToken();
		if (!stored) { goto('/login'); return; }

		if (!$auth.user) {
			try {
				const user = await getMe(stored);
				auth.init(stored, user);
			} catch {
				goto('/login');
				return;
			}
		}
		await loadKeys();
		loaded = true;
	});

	async function loadKeys() {
		const token = $auth.token || auth.getStoredToken();
		try {
			keys = await listKeys(token);
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Failed to load keys';
		}
	}

	async function handleCreate() {
		if (!newKeyName.trim()) return;
		creating = true;
		error = '';
		const token = $auth.token || auth.getStoredToken();
		try {
			const result = await createKey(token, newKeyName.trim());
			revealedKey = result.key;
			newKeyName = '';
			await loadKeys();
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Failed to create key';
		} finally {
			creating = false;
		}
	}

	async function handleRevoke(id: string) {
		if (!confirm('Revoke this key? Any running apps using it will stop working.')) return;
		const token = $auth.token || auth.getStoredToken();
		try {
			await revokeKey(token, id);
			await loadKeys();
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Failed to revoke key';
		}
	}

	function copyKey() {
		navigator.clipboard.writeText(revealedKey);
	}
</script>

<svelte:head><title>Dashboard · Trippier API</title></svelte:head>

<div class="dash">

	<!-- Header -->
	<div class="dash-head">
		<div>
			<h1>Dashboard</h1>
			{#if $auth.user}
				<p class="user-email">{$auth.user.email}</p>
			{/if}
		</div>
		<a href="/" class="btn btn-ghost btn-sm">← Back to home</a>
	</div>

	<!-- New key revealed -->
	{#if revealedKey}
		<div class="reveal-banner">
			<div class="reveal-top">
				<strong>Your new API key — copy it now, it won't be shown again.</strong>
				<button class="btn btn-sm btn-ghost" on:click={() => (revealedKey = '')}>Dismiss</button>
			</div>
			<div class="reveal-key">
				<code>{revealedKey}</code>
				<button class="btn btn-sm btn-ghost" on:click={copyKey}>Copy</button>
			</div>
		</div>
	{/if}

	{#if error}<p class="alert alert-error">{error}</p>{/if}

	<!-- Create section -->
	<div class="section">
		<h2>API Keys</h2>
		<p class="section-hint">Each account gets <strong>1 000 tokens / month</strong>, shared across all keys. Use the key in the <code>X-API-Key</code> header.</p>

		<div class="create-row">
			<input
				class="name-input"
				placeholder="Key name (e.g. my-app, production)"
				bind:value={newKeyName}
				on:keydown={(e) => e.key === 'Enter' && handleCreate()}
			/>
			<button class="btn btn-primary btn-sm" disabled={creating || !newKeyName.trim()} on:click={handleCreate}>
				{creating ? '…' : '+ New key'}
			</button>
		</div>
	</div>

	<!-- Keys list -->
	{#if !loaded}
		<p class="loading">Loading…</p>
	{:else if keys.length === 0}
		<div class="empty-state">
			<p class="empty-title">No API keys yet</p>
			<p class="empty-sub">Create a key above to start calling the POI &amp; Itinerary APIs.</p>
		</div>
	{:else}
		<ul class="key-list">
			{#each keys.filter(k => !k.revoked) as k (k.id)}
				<li class="key-item">
					<div class="key-meta">
						<div class="key-name-row">
							<span class="key-name">{k.name}</span>
							<code class="key-prefix">{k.key_prefix}…</code>
						</div>
						<button class="btn btn-sm btn-danger" on:click={() => handleRevoke(k.id)}>Revoke</button>
					</div>
					<TokenBar remaining={k.tokens_remaining} limit={k.tokens_limit} resetsInSecs={k.resets_in_secs} />
				</li>
			{/each}

			{#each keys.filter(k => k.revoked) as k (k.id)}
				<li class="key-item revoked">
					<div class="key-meta">
						<div class="key-name-row">
							<span class="key-name">{k.name}</span>
							<code class="key-prefix">{k.key_prefix}…</code>
						</div>
						<span class="revoked-badge">Revoked</span>
					</div>
				</li>
			{/each}
		</ul>
	{/if}

	<!-- Quick-start -->
	<div class="quickstart">
		<h2>Quick start</h2>
		<p class="section-hint">Replace <code>YOUR_KEY</code> with one of your API keys.</p>

		<div class="code-block">
			<p class="code-label">Search POIs near the Eiffel Tower</p>
			<pre><code>curl "https://api.trippier.dev/pois/search?lat=48.858&lng=2.294&radius=500" \
  -H "X-API-Key: YOUR_KEY"</code></pre>
		</div>

		<div class="code-block">
			<p class="code-label">Generate an itinerary</p>
			<pre><code>curl -X POST "https://api.trippier.dev/itinerary/generate" \
  -H "X-API-Key: YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{"{"}\"location\":\"Paris\",\"days\":2{"}"}'</code></pre>
		</div>
	</div>

</div>

<style>
	.dash {
		max-width: 700px;
		margin: 0 auto;
		padding: 3rem 2rem 6rem;
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.dash-head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
	}
	.dash-head h1 { font-size: 1.6rem; font-weight: 800; letter-spacing: -0.02em; }
	.user-email { font-size: 0.82rem; color: var(--muted); margin-top: 0.2rem; }

	/* Reveal banner */
	.reveal-banner {
		position: relative;
		isolation: isolate;
		background: rgba(10, 10, 10, 0.62);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		border: 1px solid rgba(16,185,129,0.25);
		border-radius: 14px;
		padding: 1rem 1.25rem;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}
	.reveal-top {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 1rem;
	}
	.reveal-top strong { font-size: 0.85rem; color: var(--accent); }
	.reveal-key {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		background: rgba(0,0,0,0.3);
		border-radius: calc(var(--radius) - 2px);
		padding: 0.5rem 0.875rem;
	}
	.reveal-key code {
		flex: 1;
		font-size: 0.82rem;
		color: var(--text);
		word-break: break-all;
	}

	/* Section */
	.section { display: flex; flex-direction: column; gap: 0.75rem; }
	.section h2 { font-size: 1.1rem; font-weight: 700; }
	.section-hint { font-size: 0.83rem; color: var(--muted); line-height: 1.5; }

	.create-row { display: flex; gap: 0.75rem; }
	.name-input {
		flex: 1;
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		color: var(--text);
		padding: 0.5rem 0.875rem;
		font-size: 0.875rem;
		font-family: var(--font);
		outline: none;
		transition: border-color 0.15s;
	}
	.name-input:focus { border-color: var(--accent); }

	/* Key list */
	.key-list { list-style: none; display: flex; flex-direction: column; gap: 0.75rem; }
	.key-item {
		position: relative;
		isolation: isolate;
		background: rgba(10, 10, 10, 0.62);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		border: 1px solid rgba(255, 255, 255, 0.07);
		border-radius: 14px;
		padding: 1rem 1.25rem;
		display: flex;
		flex-direction: column;
		gap: 0.875rem;
	}
	.key-item.revoked { opacity: 0.35; }

	.key-meta { display: flex; align-items: center; justify-content: space-between; gap: 1rem; }
	.key-name-row { display: flex; align-items: center; gap: 0.75rem; }
	.key-name { font-weight: 600; font-size: 0.9rem; }
	.key-prefix { font-size: 0.75rem; color: var(--muted); }
	.revoked-badge {
		font-size: 0.72rem;
		color: var(--muted);
		border: 1px solid var(--border);
		border-radius: 999px;
		padding: 0.15rem 0.6rem;
	}

	.loading { color: var(--muted); font-size: 0.9rem; }

	.empty-state {
		text-align: center;
		padding: 3rem 2rem;
		background: var(--surface);
		border: 1px dashed var(--border);
		border-radius: var(--radius);
	}
	.empty-title { font-weight: 600; margin-bottom: 0.4rem; }
	.empty-sub { font-size: 0.85rem; color: var(--muted); }

	/* Quickstart */
	.quickstart { display: flex; flex-direction: column; gap: 1rem; }
	.quickstart h2 { font-size: 1.1rem; font-weight: 700; }
	.code-block {
		position: relative;
		isolation: isolate;
		background: rgba(10, 10, 10, 0.62);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		border: 1px solid rgba(255, 255, 255, 0.07);
		border-radius: 14px;
		overflow: hidden;
	}
	.code-label {
		font-size: 0.75rem;
		color: var(--muted);
		padding: 0.5rem 1rem;
		border-bottom: 1px solid var(--border);
	}
	.code-block pre {
		padding: 0.875rem 1rem;
		overflow-x: auto;
	}
	.code-block code {
		font-size: 0.78rem;
		color: var(--text);
		line-height: 1.7;
	}
</style>
