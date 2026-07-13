<script lang="ts">
	import { browser } from '$app/environment';
	import { env } from '$env/dynamic/public';
	import type { DocPage } from '$lib/data/docs';
	import { buildCurl, buildJs, buildPy, sampleValue, initialValue, activeQueryParams, resolvePath } from '$lib/utils/docs';

	export let page: DocPage | null;

	type Lang = 'curl' | 'js' | 'py';
	type RouteType = 'events' | 'pois' | 'itinerary' | 'other';

	// BYOK provider registry; add new providers here.
	const BYOK_PROVIDERS = [
		// Event providers
		{ id: 'ticketmaster', label: 'Ticketmaster', header: 'X-Ticketmaster-Key',  storageKey: 'byok_tm', placeholder: 'API Key',              routeTypes: ['events'] as RouteType[] },
		{ id: 'eventbrite',   label: 'Eventbrite',   header: 'X-Eventbrite-Token',  storageKey: 'byok_eb', placeholder: 'Private Token',         routeTypes: ['events'] as RouteType[] },
		{ id: 'meetup',       label: 'Meetup',        header: 'X-Meetup-Token',      storageKey: 'byok_mu', placeholder: 'Personal Access Token', routeTypes: ['events'] as RouteType[] },
		{ id: 'openagenda',   label: 'OpenAgenda',    header: 'X-OpenAgenda-Key',    storageKey: 'byok_oa', placeholder: 'API Key',              routeTypes: ['events'] as RouteType[] },
		// POI providers (custom routes)
		{ id: 'foursquare', label: 'Foursquare',   header: 'X-Foursquare-Key', storageKey: 'byok_fs',  placeholder: 'API Key',    routeTypes: ['pois'] as RouteType[] },
		{ id: 'baidu',      label: 'Baidu Maps',   header: 'X-Baidu-Key',      storageKey: 'byok_bd',  placeholder: 'API Key',    routeTypes: ['pois'] as RouteType[] },
		{ id: 'kakao',      label: 'Kakao Maps',   header: 'X-Kakao-Key',      storageKey: 'byok_kk',  placeholder: 'REST API Key',routeTypes: ['pois'] as RouteType[] },
		{ id: 'navitime',   label: 'Navitime',     header: 'X-Navitime-Key',   storageKey: 'byok_nv',  placeholder: 'API Key',    routeTypes: ['pois'] as RouteType[] },
		{ id: 'mappls',     label: 'Mappls (IN)',  header: 'X-Mappls-Key',     storageKey: 'byok_mp',  placeholder: 'API Key',    routeTypes: ['pois'] as RouteType[] },
		{ id: 'grabmaps',   label: 'GrabMaps (SEA)',header: 'X-Grabmaps-Key',  storageKey: 'byok_gm',  placeholder: 'API Key',    routeTypes: ['pois'] as RouteType[] },
	];

	let lang: Lang = 'curl';
	let running    = false;
	let copied     = false;

	const envBaseUrl          = env.PUBLIC_POI_API_URL ?? '';
	const envItineraryBaseUrl = env.PUBLIC_ITINERARY_API_URL ?? '';
	const defaultBaseUrl      = envBaseUrl || 'https://api.poi.trippier.dev';

	let apiKey  = browser ? (localStorage.getItem('docs_api_key') ?? '') : '';
	let baseUrl = envBaseUrl || (browser ? (localStorage.getItem('docs_base_url') ?? defaultBaseUrl) : defaultBaseUrl);

	// Per-provider: selected (chip toggled) + key value
	type ProviderState = { selected: boolean; key: string };
	let byok: Record<string, ProviderState> = Object.fromEntries(
		BYOK_PROVIDERS.map(p => [p.id, {
			selected: browser ? localStorage.getItem(p.storageKey + '_on') === '1' : false,
			key:      browser ? (localStorage.getItem(p.storageKey) ?? '') : '',
		}])
	);

	let responseBody: string | null = null;
	let responseStatus: number | null = null;
	let responseMs: number | null = null;
	let responseError: string | null = null;

	// One entry per route param, plus the JSON body for non-GET routes; reset on route change.
	let paramValues: Record<string, string> = {};
	let bodyValue = '';
	/**
	 * Resets param and body inputs for the currently selected doc page.
	 * @param p - selected doc page, or null when none is selected
	 */
	function resetInputs(p: DocPage | null) {
		paramValues = Object.fromEntries((p?.params ?? []).map(pr => [pr.name, initialValue(pr)]));
		bodyValue = p?.body ?? '';
	}
	$: resetInputs(page);

	$: routeType = ((): RouteType => {
		const path = page?.path ?? '';
		if (path.includes('/events'))    return 'events';
		if (path.startsWith('/pois'))    return 'pois';
		if (path.startsWith('/itinerar')) return 'itinerary';
		return 'other';
	})();

	$: visibleProviders = BYOK_PROVIDERS.filter(p => p.routeTypes.includes(routeType));

	$: snippets = page ? { curl: buildCurl(page, baseUrl, paramValues, bodyValue), js: buildJs(page, baseUrl, paramValues, bodyValue), py: buildPy(page, baseUrl, paramValues, bodyValue) } : null;
	$: code     = snippets?.[lang] ?? '';

	$: { page; responseBody = null; responseStatus = null; responseMs = null; responseError = null; running = false; }

	/**
	 * Persists the API key, base URL, and BYOK provider settings to localStorage.
	 */
	function saveSettings() {
		if (!browser) return;
		localStorage.setItem('docs_api_key',  apiKey);
		localStorage.setItem('docs_base_url', baseUrl);
		BYOK_PROVIDERS.forEach(p => {
			localStorage.setItem(p.storageKey,         byok[p.id].key);
			localStorage.setItem(p.storageKey + '_on', byok[p.id].selected ? '1' : '0');
		});
	}

	/**
	 * Toggles a BYOK provider's selected state and saves the settings.
	 * @param id - identifier of the provider to toggle
	 */
	function toggleProvider(id: string) {
		byok[id] = { ...byok[id], selected: !byok[id].selected };
		saveSettings();
	}

	/**
	 * Copies the current code snippet to the clipboard and briefly flags success.
	 */
	function copySnippet() {
		if (browser) navigator.clipboard?.writeText(code);
		copied = true;
		setTimeout(() => (copied = false), 1200);
	}

	/**
	 * Builds the request URL for the current page from the base URL and param values.
	 * @returns fully resolved request URL
	 */
	function buildUrl(): string {
		if (!page) return baseUrl;
		const path = resolvePath(page, paramValues);
		const base = (envItineraryBaseUrl && path.startsWith('/v1/itinerary/'))
			? envItineraryBaseUrl
			: baseUrl;
		if (page.method !== 'GET') return base + path;
		const params = activeQueryParams(page, paramValues)
			.map(p => `${encodeURIComponent(p.name)}=${encodeURIComponent(paramValues[p.name])}`)
			.join('&');
		return base + path + (params ? `?${params}` : '');
	}

	/**
	 * Executes the current request against the API and records the response.
	 */
	async function tryIt() {
		if (!page) return;
		running = true;
		responseBody = null; responseStatus = null; responseMs = null; responseError = null;

		const url = buildUrl();
		const headers: Record<string, string> = {};
		if (apiKey) headers['X-API-Key'] = apiKey;
		if (page.method !== 'GET') headers['Content-Type'] = 'application/json';

		for (const p of visibleProviders) {
			const state = byok[p.id];
			if (state.selected && state.key) headers[p.header] = state.key;
		}

		const t0 = performance.now();
		try {
			const res = await fetch(url, {
				method:  page.method ?? 'GET',
				headers,
				body: page.method !== 'GET' && bodyValue.trim() ? bodyValue : undefined,
			});
			responseMs     = Math.round(performance.now() - t0);
			responseStatus = res.status;
			const text = await res.text();
			try { responseBody = JSON.stringify(JSON.parse(text), null, 2); }
			catch { responseBody = text; }
		} catch (e) {
			responseMs    = Math.round(performance.now() - t0);
			responseError = e instanceof Error ? e.message : String(e);
		} finally {
			running = false;
		}
	}
</script>

<aside class="d-panel">
	<div class="d-panel-block d-settings">
		<div class="d-panel-head">
			<span class="d-resp-label">Configuration</span>
		</div>
		<div class="d-settings-body">
			<label class="d-field">
				<span>Base URL</span>
				<input type="text" bind:value={baseUrl} on:blur={saveSettings}
					placeholder="https://api.poi.trippier.dev" spellcheck="false" />
			</label>
			<label class="d-field">
				<span>X-API-Key</span>
				<input type="password" bind:value={apiKey} on:blur={saveSettings}
					placeholder="Votre clé API" spellcheck="false" />
			</label>

			{#if visibleProviders.length > 0}
				<div class="d-byok-sep">Providers (BYOK)</div>

				<div class="d-chips">
					{#each visibleProviders as p}
						<button
							class="d-chip"
							class:active={byok[p.id].selected}
							on:click={() => toggleProvider(p.id)}
						>{p.label}</button>
					{/each}
				</div>

				{#each visibleProviders.filter(p => byok[p.id].selected) as p}
					<label class="d-field">
						<span>{p.header}</span>
						<input
							type="password"
							bind:value={byok[p.id].key}
							on:blur={saveSettings}
							placeholder={p.placeholder}
							spellcheck="false"
						/>
					</label>
				{/each}
			{/if}
		</div>
	</div>

	{#if !page}
		<div class="d-panel-empty">
			<svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z"/></svg>
			<p>Sélectionnez une route pour voir le snippet et lancer un essai.</p>
		</div>
	{:else}
		{#if (page.params && page.params.length) || page.method !== 'GET'}
			<details class="d-panel-block d-acc">
				<summary class="d-panel-head d-acc-summary">
					<span class="d-resp-label">
						Paramètres
						{#if page.method === 'GET' && page.params?.length}
							<span class="d-acc-count">{page.params.length}</span>
						{:else if page.method !== 'GET'}
							<span class="d-acc-count">body</span>
						{/if}
					</span>
					<svg class="d-acc-chevron" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M6 9l6 6 6-6"/></svg>
				</summary>
				<div class="d-settings-body">
					{#if page.method === 'GET'}
						{#each page.params ?? [] as p}
							<label class="d-field">
								<span>{p.name}{#if p.required}<em class="d-req">*</em>{/if}<em class="d-ptype">{p.type}{#if p.in === 'path'} · path{/if}</em></span>
								<input type="text" bind:value={paramValues[p.name]}
									placeholder={sampleValue(p)} spellcheck="false" />
							</label>
						{/each}
					{:else}
						<label class="d-field">
							<span>body<em class="d-ptype">json</em></span>
							<textarea class="d-body" bind:value={bodyValue} rows="10" spellcheck="false"></textarea>
						</label>
					{/if}
				</div>
			</details>
		{/if}

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
		overscroll-behavior: contain;
		min-height: 0;
		padding: 24px;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	/* Blocks must not be squeezed by the flex column — keep their natural height
	   so the panel (not the blocks) is what scrolls. */
	.d-panel > * { flex-shrink: 0; }

	.d-acc { padding: 0; }
	.d-acc-summary { cursor: pointer; list-style: none; user-select: none; }
	.d-acc-summary::-webkit-details-marker { display: none; }
	.d-acc-summary .d-resp-label { display: flex; align-items: center; gap: 8px; }
	.d-acc-count {
		font-family: var(--font-mono);
		font-size: 10px;
		padding: 1px 6px;
		border-radius: 20px;
		color: var(--code-text-3);
		background: var(--code-bg);
		text-transform: none;
		letter-spacing: 0;
	}
	.d-acc-chevron { color: var(--code-text-3); transition: transform .15s ease; flex-shrink: 0; }
	.d-acc[open] .d-acc-chevron { transform: rotate(180deg); }
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
	.d-field { display: flex; flex-direction: column; gap: 4px; }
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
	.d-field span { display: flex; align-items: center; gap: 6px; }
	.d-req { color: var(--accent); font-style: normal; }
	.d-ptype {
		font-style: normal;
		text-transform: none;
		letter-spacing: 0;
		color: var(--text-3);
		opacity: 0.7;
	}
	.d-body {
		background: var(--bg);
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		padding: 8px 10px;
		font-family: var(--font-mono);
		font-size: 12px;
		line-height: 1.5;
		color: var(--text);
		outline: none;
		resize: vertical;
		transition: border-color .12s ease;
	}
	.d-body:focus { border-color: var(--accent); }

	.d-byok-sep {
		font-family: var(--font-mono);
		font-size: 10px;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--text-3);
		padding-top: 6px;
		border-top: 1px solid var(--border);
	}

	.d-chips { display: flex; flex-wrap: wrap; gap: 6px; }
	.d-chip {
		font-family: var(--font-mono);
		font-size: 11px;
		padding: 4px 10px;
		border-radius: 20px;
		border: 1px solid var(--border);
		background: var(--bg);
		color: var(--text-3);
		cursor: pointer;
		transition: all .12s ease;
	}
	.d-chip:hover { border-color: var(--accent); color: var(--text); }
	.d-chip.active {
		background: color-mix(in oklch, var(--accent) 12%, transparent);
		border-color: var(--accent);
		color: var(--accent);
	}

	.d-panel-block {
		border: 1px solid var(--code-border);
		border-radius: var(--r-md);
		background: var(--code-bg-2);
		overflow: hidden;
	}
	.d-panel-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
		padding: 8px 12px;
		background: var(--code-surface);
		border-bottom: 1px solid var(--code-border);
	}
	.d-tabs { display: flex; gap: 2px; }
	.d-tabs button {
		font-family: var(--font-mono);
		font-size: 11.5px;
		color: var(--code-text-3);
		padding: 4px 10px;
		border-radius: 4px;
		border: none;
		background: none;
		cursor: pointer;
		transition: color .12s ease, background .12s ease;
	}
	.d-tabs button:hover { color: var(--code-text); }
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
		color: var(--code-text-3);
		padding: 3px 8px;
		border-radius: 4px;
		border: none;
		background: none;
		cursor: pointer;
		transition: color .12s ease, background .12s ease;
	}
	.d-panel-copy:hover { color: var(--code-text); background: var(--code-bg); }
	.d-panel-code {
		margin: 0;
		padding: 14px 16px;
		font-family: var(--font-mono);
		font-size: 12px;
		color: var(--code-text-2);
		line-height: 1.6;
		white-space: pre-wrap;
		word-break: break-word;
		overflow-x: auto;
		max-height: 360px;
		overflow-y: auto;
	}
	.d-panel-code.resp { color: var(--code-text-3); }
	.d-err-text { color: oklch(72% 0.16 25); }
	.d-panel-actions {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 10px 12px;
		border-top: 1px solid var(--code-border);
		background: var(--code-surface);
	}
	.d-btn-primary {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 8px 14px;
		background: var(--accent);
		color: var(--accent-on);
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
	.d-panel-base { font-family: var(--font-mono); font-size: 11px; color: var(--code-text-3); }
	.d-resp-label {
		font-family: var(--font-mono);
		font-size: 11px;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--code-text-3);
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
		color: var(--code-text-3);
		background: var(--code-bg);
	}
	.d-status-pill.ok      { color: var(--accent); background: color-mix(in oklch, var(--accent) 14%, transparent); }
	.d-status-pill.warn    { color: oklch(82% 0.12 80); background: color-mix(in oklch, oklch(82% 0.12 80) 12%, transparent); }
	.d-status-pill.err     { color: oklch(72% 0.16 25); background: color-mix(in oklch, oklch(72% 0.16 25) 12%, transparent); }
	.d-status-pill.pending { color: oklch(82% 0.12 80); background: color-mix(in oklch, oklch(82% 0.12 80) 12%, transparent); }
	.d-status-pill.timing  { color: var(--code-text-3); }
</style>
