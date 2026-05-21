<script lang="ts">
	import { browser } from '$app/environment';
	import { env } from '$env/dynamic/public';
	import type { DocPage } from '$lib/data/docs';
	import { buildCurl, buildJs, buildPy, sampleValue } from '$lib/utils/docs';

	export let page: DocPage | null;

	type Lang = 'curl' | 'js' | 'py';

	let lang: Lang = 'curl';
	let running    = false;
	let copied     = false;

	// If PUBLIC_POI_API_URL is set (dev), always use it — ignore localStorage to avoid stale prod URL.
	const envBaseUrl          = env.PUBLIC_POI_API_URL ?? '';
	const envItineraryBaseUrl = env.PUBLIC_ITINERARY_API_URL ?? '';
	const defaultBaseUrl      = envBaseUrl || 'https://api.poi.trippier.dev';

	let apiKey  = browser ? (localStorage.getItem('docs_api_key') ?? '') : '';
	let baseUrl = envBaseUrl || (browser ? (localStorage.getItem('docs_base_url') ?? defaultBaseUrl) : defaultBaseUrl);
	let tmKey   = browser ? (localStorage.getItem('docs_tm_key') ?? '') : '';
	let ebToken = browser ? (localStorage.getItem('docs_eb_token') ?? '') : '';

	$: isEventsRoute = page?.path?.includes('/events') ?? false;

	let responseBody: string | null = null;
	let responseStatus: number | null = null;
	let responseMs: number | null = null;
	let responseError: string | null = null;

	$: snippets = page ? { curl: buildCurl(page, baseUrl), js: buildJs(page, baseUrl), py: buildPy(page, baseUrl) } : null;
	$: code     = snippets?.[lang] ?? '';

	$: { page; responseBody = null; responseStatus = null; responseMs = null; responseError = null; running = false; }

	function saveSettings() {
		if (!browser) return;
		localStorage.setItem('docs_api_key',  apiKey);
		localStorage.setItem('docs_base_url', baseUrl);
		localStorage.setItem('docs_tm_key',   tmKey);
		localStorage.setItem('docs_eb_token', ebToken);
	}

	function copySnippet() {
		if (browser) navigator.clipboard?.writeText(code);
		copied = true;
		setTimeout(() => (copied = false), 1200);
	}

	function buildUrl(): string {
		const path = page?.path ?? '';
		const base = (envItineraryBaseUrl && path.startsWith('/itinerary/'))
			? envItineraryBaseUrl
			: baseUrl;
		if (page?.method !== 'GET') return base + path;
		const SNIPPET_KEYS = new Set(['lat', 'lng', 'radius', 'date']);
		const params = (page?.params ?? [])
			.filter(p => p.in !== 'path' && SNIPPET_KEYS.has(p.name))
			.map(p => `${encodeURIComponent(p.name)}=${encodeURIComponent(sampleValue(p))}`)
			.join('&');
		return base + path + (params ? `?${params}` : '');
	}

	async function tryIt() {
		if (!page) return;
		running = true;
		responseBody = null; responseStatus = null; responseMs = null; responseError = null;

		const url = buildUrl();
		const headers: Record<string, string> = {};
		if (apiKey) headers['X-API-Key'] = apiKey;
		if (page.method !== 'GET') headers['Content-Type'] = 'application/json';
		if (isEventsRoute && tmKey)   headers['X-Ticketmaster-Key']   = tmKey;
		if (isEventsRoute && ebToken) headers['X-Eventbrite-Token'] = ebToken;

		const t0 = performance.now();
		try {
			const res = await fetch(url, {
				method:  page.method ?? 'GET',
				headers,
				body: page.method !== 'GET' && page.body ? page.body : undefined,
			});
			responseMs     = Math.round(performance.now() - t0);
			responseStatus = res.status;
			const text = await res.text();
			try {
				responseBody = JSON.stringify(JSON.parse(text), null, 2);
			} catch {
				responseBody = text;
			}
		} catch (e) {
			responseMs    = Math.round(performance.now() - t0);
			responseError = e instanceof Error ? e.message : String(e);
		} finally {
			running = false;
		}
	}
</script>

<aside class="d-panel">
	<!-- Settings block -->
	<div class="d-panel-block d-settings">
		<div class="d-panel-head">
			<span class="d-resp-label">Configuration</span>
		</div>
		<div class="d-settings-body">
			<label class="d-field">
				<span>Base URL</span>
				<input
					type="text"
					bind:value={baseUrl}
					on:blur={saveSettings}
					placeholder="https://api.poi.trippier.dev"
					spellcheck="false"
				/>
			</label>
			<label class="d-field">
				<span>X-API-Key</span>
				<input
					type="password"
					bind:value={apiKey}
					on:blur={saveSettings}
					placeholder="Votre clé API"
					spellcheck="false"
				/>
			</label>
			{#if isEventsRoute}
				<div class="d-byok-sep">Clés providers (BYOK)</div>
				<label class="d-field">
					<span>X-Ticketmaster-Key</span>
					<input
						type="password"
						bind:value={tmKey}
						on:blur={saveSettings}
						placeholder="Votre clé Ticketmaster"
						spellcheck="false"
					/>
				</label>
				<label class="d-field">
					<span>X-Eventbrite-Token</span>
					<input
						type="password"
						bind:value={ebToken}
						on:blur={saveSettings}
						placeholder="Votre token Eventbrite"
						spellcheck="false"
					/>
				</label>
			{/if}
		</div>
	</div>

	{#if !page}
		<div class="d-panel-empty">
			<svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z"/></svg>
			<p>Sélectionnez une route pour voir le snippet et lancer un essai.</p>
		</div>
	{:else}
		<div class="d-panel-block">
			<div class="d-panel-head">
				<div class="d-tabs">
					<button class:active={lang === 'curl'} on:click={() => lang = 'curl'}>cURL</button>
					<button class:active={lang === 'js'}   on:click={() => lang = 'js'}>JavaScript</button>
					<button class:active={lang === 'py'}   on:click={() => lang = 'py'}>Python</button>
				</div>
				<button class="d-panel-copy" on:click={copySnippet}>
					{#if copied}
						<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12l5 5 9-11"/></svg> copié
					{:else}
						<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="9" y="9" width="11" height="11" rx="2"/><path d="M5 15V5a2 2 0 0 1 2-2h10"/></svg> copier
					{/if}
				</button>
			</div>
			<pre class="d-panel-code">{code}</pre>
			<div class="d-panel-actions">
				<button class="d-btn-primary d-try" on:click={tryIt} disabled={running}>
					{#if running}
						<span class="d-spin">⟳</span> exécution…
					{:else}
						<svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z"/></svg>
						Try it
					{/if}
				</button>
				<span class="d-panel-base">{(() => { try { return new URL(baseUrl).hostname; } catch { return baseUrl; } })()}</span>
			</div>
		</div>

		<div class="d-panel-block d-panel-resp">
			<div class="d-panel-head">
				<span class="d-resp-label">
					Réponse
					{#if responseStatus !== null}
						<span class="d-status-pill" class:ok={responseStatus < 300} class:warn={responseStatus >= 400 && responseStatus < 500} class:err={responseStatus >= 500}>
							{responseStatus}
						</span>
					{/if}
					{#if responseMs !== null}
						<span class="d-status-pill timing">{responseMs} ms</span>
					{/if}
					{#if running}
						<span class="d-status-pill pending">pending…</span>
					{/if}
				</span>
			</div>
			<pre class="d-panel-code resp">{#if responseError}<span class="d-err-text">{responseError}</span>{:else}{responseBody ?? (running ? '…' : '// cliquez sur "Try it" pour exécuter l\'appel')}{/if}</pre>
		</div>
	{/if}
</aside>

<style>
	.d-panel {
		border-left: 1px solid var(--border);
		background: color-mix(in oklch, var(--bg) 85%, transparent);
		height: calc(100vh - 56px);
		position: sticky;
		top: 56px;
		overflow-y: auto;
		padding: 24px;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.d-panel-empty {
		display: flex;
		flex-direction: column;
		gap: 12px;
		padding: 40px 20px;
		text-align: center;
		color: var(--text-3);
		font-size: 13px;
		align-items: center;
		margin-top: 20px;
	}
	.d-panel-empty svg { color: var(--accent); opacity: 0.5; }
	.d-panel-empty p   { color: var(--text-3); margin: 0; }

	.d-settings-body {
		padding: 12px;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.d-field {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.d-field span {
		font-family: var(--font-mono);
		font-size: 10px;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--text-3);
	}
	.d-field input {
		background: var(--bg);
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		padding: 7px 10px;
		font-family: var(--font-mono);
		font-size: 12px;
		color: var(--text);
		outline: none;
		transition: border-color .12s ease;
	}
	.d-field input:focus { border-color: var(--accent); }
	.d-field input::placeholder { color: var(--text-3); }
	.d-byok-sep {
		font-family: var(--font-mono);
		font-size: 10px;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--text-3);
		padding-top: 6px;
		border-top: 1px solid var(--border);
	}

	.d-panel-block {
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		background: oklch(9% 0.01 175);
		overflow: hidden;
	}
	.d-panel-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
		padding: 8px 12px;
		background: var(--surface);
		border-bottom: 1px solid var(--border);
	}
	.d-tabs { display: flex; gap: 2px; }
	.d-tabs button {
		font-family: var(--font-mono);
		font-size: 11.5px;
		color: var(--text-3);
		padding: 4px 10px;
		border-radius: 4px;
		border: none;
		background: none;
		cursor: pointer;
		transition: color .12s ease, background .12s ease;
	}
	.d-tabs button:hover { color: var(--text); }
	.d-tabs button.active {
		color: var(--accent);
		background: color-mix(in oklch, var(--accent) 10%, transparent);
	}
	.d-panel-copy {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--text-3);
		padding: 3px 8px;
		border-radius: 4px;
		border: none;
		background: none;
		cursor: pointer;
		transition: color .12s ease, background .12s ease;
	}
	.d-panel-copy:hover { color: var(--text); background: var(--bg); }
	.d-panel-code {
		margin: 0;
		padding: 14px 16px;
		font-family: var(--font-mono);
		font-size: 12px;
		color: var(--text-2);
		line-height: 1.6;
		white-space: pre-wrap;
		word-break: break-word;
		overflow-x: auto;
		max-height: 360px;
		overflow-y: auto;
	}
	.d-panel-code.resp { color: var(--text-3); }
	.d-err-text { color: oklch(72% 0.16 25); }
	.d-panel-actions {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 10px 12px;
		border-top: 1px solid var(--border);
		background: var(--surface);
	}
	.d-btn-primary {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 8px 14px;
		background: var(--accent);
		color: #08120e;
		border-radius: var(--r-md);
		font-weight: 600;
		font-size: 13px;
		border: none;
		cursor: pointer;
		transition: filter .15s ease;
	}
	.d-btn-primary:hover { filter: brightness(1.08); }
	.d-btn-primary:disabled { opacity: 0.5; cursor: not-allowed; filter: none; }
	.d-try { padding: 7px 14px; font-size: 12.5px; }
	.d-spin { display: inline-block; animation: spin .7s linear infinite; }
	@keyframes spin { to { transform: rotate(360deg); } }
	.d-panel-base { font-family: var(--font-mono); font-size: 11px; color: var(--text-3); }
	.d-resp-label {
		font-family: var(--font-mono);
		font-size: 11px;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--text-3);
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.d-status-pill {
		font-family: var(--font-mono);
		font-size: 10.5px;
		padding: 2px 7px;
		border-radius: 3px;
		letter-spacing: 0.04em;
		text-transform: none;
		color: var(--text-3);
		background: var(--surface);
	}
	.d-status-pill.ok      { color: var(--accent); background: color-mix(in oklch, var(--accent) 14%, transparent); }
	.d-status-pill.warn    { color: oklch(82% 0.12 80); background: color-mix(in oklch, oklch(82% 0.12 80) 12%, transparent); }
	.d-status-pill.err     { color: oklch(72% 0.16 25); background: color-mix(in oklch, oklch(72% 0.16 25) 12%, transparent); }
	.d-status-pill.pending { color: oklch(82% 0.12 80); background: color-mix(in oklch, oklch(82% 0.12 80) 12%, transparent); }
	.d-status-pill.timing  { color: var(--text-3); }
</style>
