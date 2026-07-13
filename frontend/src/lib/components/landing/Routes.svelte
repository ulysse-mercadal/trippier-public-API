<script lang="ts">
	import { t } from '$lib/i18n';
	import { PAGES } from '$lib/data/docs';

	// Single source of truth — derived from docs data
	const ROUTES = PAGES
		.filter(p => p.kind === 'route' && p.path && p.method)
		.map(p => {
			const subpath = p.path!.startsWith('/v1/pois/')
				? p.path!.replace('/v1/pois/', '')
				: null;
			return { m: p.method!, path: p.path!, cost: p.cost ?? 0, subpath };
		});

	let activeIdx = 4;
	let running   = false;
	let response: string | null = null;
	let httpStatus: number | null = null;
	let elapsed: number | null = null;

	/**
	 * Picks the badge color for a given HTTP method.
	 * @param m - HTTP method (e.g. GET, POST)
	 * @returns CSS color value
	 */
	function methodColor(m: string): string {
		return ({ GET: 'var(--accent)', POST: '#e4b07a' } as Record<string, string>)[m] ?? 'var(--text)';
	}

	/**
	 * Sets the active route and clears any previous try-it result.
	 * @param i - index of the route to select
	 */
	function selectRoute(i: number) {
		activeIdx = i;
		response = null;
		httpStatus = null;
		elapsed = null;
	}

	/**
	 * Executes a live request against the currently selected route's proxy endpoint and records the outcome.
	 */
	async function tryRoute() {
		running = true;
		response = null;
		httpStatus = null;
		elapsed = null;

		const route = ROUTES[activeIdx];
		const t0 = Date.now();
		try {
			let res: Response;

			if (route.path === '/health') {
				res = await fetch('/api/proxy/health');
			} else if (route.path === '/v1/itinerary/generate') {
				res = await fetch('/api/proxy/itinerary', {
					method: 'POST',
					headers: { 'content-type': 'application/json' },
					body: JSON.stringify({
						days: 1,
						poi_query: { lat: 40.7549, lng: -73.9840, radius: 5000 },
						preferences: { pace: 'moderate', start_time: '09:00' },
					}),
				});
			} else {
				const qs = new URLSearchParams({ subpath: route.subpath!, lat: '40.7549', lng: '-73.9840', radius: '5000' });
				res = await fetch(`/api/proxy/pois?${qs}`);
			}

			elapsed = Date.now() - t0;
			httpStatus = res.status;
			const data = await res.json();
			response = JSON.stringify(data, null, 2);
		} catch {
			elapsed = Date.now() - t0;
			response = $t('routes_error');
		}

		running = false;
	}

	$: active    = ROUTES[activeIdx];
	$: activePage = PAGES.find(p => p.path === active?.path && p.method === active?.m);
	$: requestBody = active?.m === 'POST'
		? `\nContent-Type: application/json\n\n${activePage?.body ?? '{}'}`
		: '?lat=40.7549&lng=-73.9840&radius=5000';
</script>

<section class="routes-section" id="routes">
	<div class="wrap">
		<div class="section-head">
			<span class="eyebrow">{$t('routes_eyebrow')}</span>
			<h2>{$t('routes_title')}</h2>
			<p class="section-sub">{$t('routes_sub')}</p>
		</div>
		<div class="routes-layout">
			<ul class="routes-list">
				{#each ROUTES as rt, i}
					<li
						class:active={i === activeIdx}
						on:click={() => selectRoute(i)}
						on:keypress={() => selectRoute(i)}
						role="button"
						tabindex="0"
					>
						<span class="method" style="color:{methodColor(rt.m)}">{rt.m}</span>
						<span class="rt-path">{rt.path}</span>
						<span class="rt-cost">{rt.cost === 0 ? 'free' : `${rt.cost} tk`}</span>
					</li>
				{/each}
			</ul>
			<div class="routes-detail">
				<div class="rd-head">
					<span class="method" style="color:{methodColor(active.m)}">{active.m}</span>
					<code class="rd-path">{active.path}</code>
					<button class="btn-primary btn-try" on:click={tryRoute} disabled={running}>
						{#if running}
							{$t('routes_running')}
						{:else}
							<svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z"/></svg>
							Try it
						{/if}
					</button>
				</div>
				<p class="rd-desc">{activePage?.summary ?? ''}</p>
				<div class="rd-meta">
					<span>{$t('routes_cost_label')} : <strong>{active.cost === 0 ? $t('routes_cost_free') : `${active.cost} ${active.cost > 1 ? $t('routes_cost_tokens') : $t('routes_cost_token')}`}</strong></span>
					<span>{$t('routes_auth_label')} : <strong>{active.cost === 0 ? $t('routes_auth_optional') : 'X-API-Key'}</strong></span>
					<span>Rate-limit : <strong>{$t('routes_ratelimit')}</strong></span>
				</div>
				<div class="rd-block">
					<div class="rd-block-head">{$t('routes_request')}</div>
					<pre class="rd-code">{active.m} {active.path}{requestBody}</pre>
				</div>
				<div class="rd-block">
					<div class="rd-block-head">
						{$t('routes_response')}
						{#if httpStatus !== null}
							<span class="ok-pill" class:err-pill={httpStatus >= 400}>{httpStatus}</span>
						{/if}
						{#if elapsed !== null}
							<span class="elapsed">{elapsed} ms</span>
						{/if}
					</div>
					<pre class="rd-code rd-code--response">{response ?? (running ? '…' : $t('routes_waiting'))}</pre>
				</div>
			</div>
		</div>
		<div class="routes-foot">
			<a href="/docs" class="btn-primary">
				{$t('routes_doc_cta')}
				<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M5 12h14M13 6l6 6-6 6"/></svg>
			</a>
			<span>{$t('routes_doc_desc')}</span>
		</div>
	</div>
</section>

<style>
	.routes-section { padding-block: 96px; }
	.routes-layout {
		display: grid;
		grid-template-columns: 360px 1fr;
		gap: 1px;
		border: 1px solid var(--border);
		border-radius: var(--r-lg);
		background: var(--border);
		overflow: hidden;
	}
	.routes-list {
		list-style: none;
		padding: 8px 0;
		margin: 0;
		background: var(--bg-2);
		max-height: 600px;
		overflow-y: auto;
	}
	.routes-list li {
		display: grid;
		grid-template-columns: 60px 1fr auto;
		align-items: center;
		gap: 12px;
		padding: 12px 18px;
		cursor: pointer;
		transition: background .12s ease;
		border-left: 2px solid transparent;
	}
	.routes-list li:hover { background: var(--surface); }
	.routes-list li.active { background: var(--surface); border-left-color: var(--accent); }
	.method { font-family: var(--font-mono); font-size: 11px; font-weight: 600; letter-spacing: 0.04em; }
	.rt-path { font-family: var(--font-mono); font-size: 13px; color: var(--text); }
	.rt-cost {
		font-family: var(--font-mono);
		font-size: 10.5px;
		color: var(--text-3);
		padding: 2px 6px;
		border-radius: 4px;
		background: var(--bg);
		border: 1px solid var(--border);
	}
	.routes-detail {
		background: var(--bg-2);
		padding: 24px 28px 28px;
		display: flex;
		flex-direction: column;
		gap: 18px;
		min-width: 0;
		overflow: hidden;
	}
	.rd-head { display: flex; align-items: center; gap: 14px; }
	.rd-path { font-family: var(--font-mono); font-size: 16px; color: var(--text); flex: 1; }
	.btn-try { padding: 8px 14px; font-size: 13px; }
	.rd-desc { color: var(--text-2); font-size: 14.5px; line-height: 1.5; margin: 0; }
	.rd-meta {
		display: flex;
		gap: 24px;
		flex-wrap: wrap;
		padding: 12px 0;
		border-block: 1px solid var(--border);
		font-family: var(--font-mono);
		font-size: 12px;
		color: var(--text-3);
	}
	.rd-meta strong { color: var(--text); font-weight: 500; margin-left: 6px; }
	.rd-block { display: flex; flex-direction: column; gap: 8px; }
	.rd-block-head {
		display: flex;
		align-items: center;
		gap: 10px;
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--text-3);
		letter-spacing: 0.06em;
		text-transform: uppercase;
	}
	.ok-pill {
		font-size: 10px;
		padding: 1px 6px;
		background: color-mix(in oklch, var(--accent) 18%, transparent);
		color: var(--accent);
		border-radius: 3px;
		letter-spacing: 0.04em;
	}
	.err-pill {
		background: color-mix(in oklch, #e78f8f 18%, transparent);
		color: #e78f8f;
	}
	.elapsed { font-family: var(--font-mono); font-size: 10px; color: var(--text-3); }
	.rd-code {
		background: var(--code-bg);
		border: 1px solid var(--code-border);
		border-radius: var(--r-md);
		padding: 14px 16px;
		font-family: var(--font-mono);
		font-size: 12.5px;
		color: var(--code-text-2);
		line-height: 1.6;
		margin: 0;
		white-space: pre-wrap;
		word-break: break-word;
		min-height: 50px;
	}
	.rd-code--response {
		max-height: 280px;
		overflow-y: auto;
		overflow-x: hidden;
	}
	.routes-foot {
		display: flex;
		align-items: center;
		gap: 16px;
		margin-top: 22px;
		padding: 18px 22px;
		border: 1px dashed color-mix(in oklch, var(--accent) 25%, var(--border));
		border-radius: var(--r-md);
		background: color-mix(in oklch, var(--accent) 3%, transparent);
		flex-wrap: wrap;
	}
	.routes-foot span { font-size: 13px; color: var(--text-2); line-height: 1.5; flex: 1; min-width: 240px; }

	@media (max-width: 980px) {
		.routes-layout { grid-template-columns: 1fr; }
		.routes-list   { max-height: 280px; }
	}
</style>
