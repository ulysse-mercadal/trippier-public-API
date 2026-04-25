<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth';
	import { register, login, getMe, ApiError } from '$lib/api';

	let mode: 'login' | 'register' = 'register';
	let email    = '';
	let password = '';
	let error    = '';
	let success  = '';
	let loading  = false;

	onMount(async () => {
		const stored = auth.getStoredToken();
		if (!stored) return;
		try {
			const user = await getMe(stored);
			auth.init(stored, user);
		} catch {
			// token expired — stay on landing
		}
	});

	async function submit() {
		error   = '';
		success = '';
		loading = true;
		try {
			if (mode === 'register') {
				await register(email, password);
				success = 'Account created! Check your inbox to verify, then sign in.';
				mode    = 'login';
				password = '';
			} else {
				const token = await login(email, password);
				const user  = await getMe(token);
				auth.init(token, user);
				auth.storeToken(token);
				goto('/dashboard');
			}
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong';
		} finally {
			loading = false;
		}
	}
</script>

<!-- ── Hero ─────────────────────────────────────────────────────────────────── -->
<section class="hero">
	<div class="hero-pills">
		<span class="pill">Public API</span>
		<span class="pill">Beta</span>
		<a
			href="https://github.com/ulysse-mercadal/trippier-public-API"
			class="pill pill-oss"
			target="_blank"
			rel="noopener noreferrer"
		>
			<svg width="13" height="13" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">
				<path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0 0 16 8c0-4.42-3.58-8-8-8z"/>
			</svg>
			Open source
		</a>
	</div>

	<h1>Travel data,<br />for builders.</h1>

	<p class="tagline">
		Search millions of points of interest worldwide and generate intelligent
		itineraries — with a single API key.
	</p>

	<div class="cta-row">
		<a href="#auth" class="btn btn-primary">Get your free key</a>
		<a href="#apis" class="btn btn-ghost">Explore the APIs</a>
	</div>

	<div class="token-pill">
		<span class="dot"></span>
		1 000 free trial tokens on signup
	</div>
</section>

<!-- ── APIs ─────────────────────────────────────────────────────────────────── -->
<section class="apis" id="apis">
	<div class="api-card glass-card">
		<div class="api-header">
			<div class="api-icon">
				<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
					<path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/>
					<circle cx="12" cy="10" r="3"/>
				</svg>
			</div>
			<div class="api-title">
				<h3>POI Search API</h3>
				<p class="api-sub">Points of interest, anywhere on Earth</p>
			</div>
			<span class="badge">1 token / req</span>
		</div>
		<p class="api-desc">
			Aggregate geo-enriched POIs from OpenStreetMap, Wikipedia, Wikivoyage
			and GeoNames in a single call. Filter by type, search by radius,
			polygon or district name.
		</p>
		<div class="endpoint-row">
			<code>GET /pois/search?lat=48.85&amp;lng=2.35&amp;radius=2000</code>
		</div>
		<ul class="features">
			<li>Multi-source aggregation with deduplication</li>
			<li>Type weights &amp; relevance scoring [0 – 100]</li>
			<li>Radius · polygon · district search modes</li>
			<li>Parallel provider fetching, Redis-cached</li>
		</ul>
	</div>

	<div class="api-card glass-card">
		<div class="api-header">
			<div class="api-icon">
				<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
					<circle cx="6" cy="19" r="3"/>
					<circle cx="18" cy="5" r="3"/>
					<path d="M12 19h4.5a3.5 3.5 0 0 0 0-7h-8a3.5 3.5 0 0 1 0-7H12"/>
				</svg>
			</div>
			<div class="api-title">
				<h3>Itinerary API</h3>
				<p class="api-sub">Smart trip planning, zero boilerplate</p>
			</div>
			<span class="badge badge-amber">50 tokens / req</span>
		</div>
		<p class="api-desc">
			Turn a list of POIs and constraints (days, pace, transport) into an
			optimised day-by-day itinerary. Built on the POI Search API — tokens
			cover the planning computation.
		</p>
		<div class="endpoint-row">
			<code>POST /itinerary/generate</code>
		</div>
		<ul class="features">
			<li>Day-by-day schedule generation</li>
			<li>Configurable pace &amp; transport mode</li>
			<li>Opening hours &amp; proximity awareness</li>
			<li>JSON output ready for your UI</li>
		</ul>
	</div>
</section>

<!-- ── Token model ───────────────────────────────────────────────────────────── -->
<section class="token-section">
	<div class="token-card glass-card">
		<h2>Simple token model</h2>
		<p class="token-sub">
			Every account starts with <strong>1 000 tokens</strong> per month, shared across all your keys.
		</p>
		<div class="token-table">
			<div class="token-row header">
				<span>Endpoint</span><span>Cost</span><span>1 000 tokens =</span>
			</div>
			<div class="token-row">
				<code>GET /pois/search</code><span>1 token</span><span>1 000 searches</span>
			</div>
			<div class="token-row">
				<code>POST /itinerary/generate</code><span>50 tokens</span><span>20 itineraries</span>
			</div>
		</div>
	</div>
</section>

<!-- ── Open source callout ───────────────────────────────────────────────────── -->
<section class="oss-section">
	<div class="oss-card glass-card">
		<div class="oss-body">
			<div>
				<p class="oss-title">Fully open source</p>
				<p class="oss-desc">
					MIT licensed. Deploy the entire stack yourself with Docker Compose,
					or use the hosted API with free trial tokens.
				</p>
			</div>
			<a
				href="https://github.com/ulysse-mercadal/trippier-public-API"
				class="btn btn-ghost"
				target="_blank"
				rel="noopener noreferrer"
			>
				View on GitHub
			</a>
		</div>
	</div>
</section>

<!-- ── Auth ──────────────────────────────────────────────────────────────────── -->
<section class="auth-section" id="auth">
	<div class="auth-card glass-card">
		<h2>{mode === 'register' ? 'Start for free' : 'Welcome back'}</h2>
		<p class="auth-sub">
			{mode === 'register'
				? '1 000 trial tokens. Create your key in 30 seconds.'
				: 'Sign in to manage your API keys.'}
		</p>

		<div class="tabs">
			<button class="tab" class:active={mode === 'register'} on:click={() => { mode = 'register'; error = ''; }}>Create account</button>
			<button class="tab" class:active={mode === 'login'}    on:click={() => { mode = 'login';    error = ''; }}>Sign in</button>
		</div>

		{#if error}  <p class="alert alert-error">{error}</p>   {/if}
		{#if success}<p class="alert alert-success">{success}</p>{/if}

		<form on:submit|preventDefault={submit}>
			<div class="field">
				<label for="email">Email</label>
				<input id="email" type="email" bind:value={email} placeholder="you@example.com" required autocomplete="email" />
			</div>
			<div class="field">
				<label for="password">Password</label>
				<input id="password" type="password" bind:value={password} placeholder="••••••••" required autocomplete={mode === 'login' ? 'current-password' : 'new-password'} />
			</div>
			<button class="btn btn-primary submit-btn" type="submit" disabled={loading}>
				{loading ? '…' : mode === 'register' ? 'Create account & get key' : 'Sign in'}
			</button>
		</form>
	</div>
</section>

<style>
	/* ── Hero ── */
	.hero {
		text-align: center;
		padding: 9rem 2rem 6rem;
		max-width: 760px;
		margin: 0 auto;
	}

	.hero-pills {
		display: flex;
		justify-content: center;
		gap: 0.5rem;
		flex-wrap: wrap;
		margin-bottom: 2rem;
	}

	.pill {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
		font-size: 0.72rem;
		font-weight: 600;
		letter-spacing: 0.07em;
		text-transform: uppercase;
		color: var(--muted);
		border: 1px solid var(--border);
		border-radius: 999px;
		padding: 0.3rem 0.8rem;
	}

	.pill-oss {
		color: var(--accent);
		border-color: rgba(16, 185, 129, 0.3);
		background: var(--accent-glow);
		transition: opacity 0.15s;
	}
	.pill-oss:hover { opacity: 0.8; }

	h1 {
		font-size: clamp(2.8rem, 6.5vw, 4.4rem);
		font-weight: 800;
		letter-spacing: -0.04em;
		line-height: 1.05;
		color: var(--text);
		margin-bottom: 1.5rem;
	}

	.tagline {
		color: var(--muted);
		font-size: 1.1rem;
		line-height: 1.8;
		max-width: 480px;
		margin: 0 auto 2.75rem;
	}

	.cta-row {
		display: flex;
		gap: 0.75rem;
		justify-content: center;
		flex-wrap: wrap;
		margin-bottom: 2rem;
	}

	.token-pill {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.8rem;
		color: var(--muted);
		border: 1px solid var(--border);
		border-radius: 999px;
		padding: 0.35rem 1rem;
	}

	.dot {
		width: 7px; height: 7px;
		border-radius: 50%;
		background: var(--accent);
		flex-shrink: 0;
	}

	/* ── APIs ── */
	.apis {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
		gap: 1rem;
		max-width: 960px;
		margin: 0 auto;
		padding: 0 2rem 2rem;
	}

	.api-card {
		padding: 2rem 2.25rem;
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
	}

	.api-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
	}

	.api-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 34px;
		height: 34px;
		border-radius: 8px;
		background: var(--accent-glow);
		border: 1px solid rgba(16, 185, 129, 0.2);
		color: var(--accent);
		flex-shrink: 0;
		margin-top: 1px;
	}

	.api-title h3    { font-size: 1.05rem; font-weight: 700; margin-bottom: 0.2rem; }
	.api-sub         { font-size: 0.8rem; color: var(--muted); }

	.badge {
		background: var(--accent-glow);
		color: var(--accent);
		border: 1px solid rgba(16, 185, 129, 0.3);
		border-radius: 999px;
		padding: 0.2rem 0.7rem;
		font-size: 0.72rem;
		font-weight: 600;
		white-space: nowrap;
		flex-shrink: 0;
	}
	.badge-amber {
		background: rgba(245, 158, 11, 0.1);
		color: #f59e0b;
		border-color: rgba(245, 158, 11, 0.3);
	}

	.api-desc {
		font-size: 0.875rem;
		color: var(--muted);
		line-height: 1.7;
	}

	.endpoint-row {
		background: rgba(0, 0, 0, 0.35);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 0.65rem 1rem;
		overflow-x: auto;
	}
	.endpoint-row code { font-size: 0.78rem; color: var(--accent); }

	.features {
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 0.45rem;
	}
	.features li {
		font-size: 0.83rem;
		color: var(--muted);
		padding-left: 1rem;
		position: relative;
		line-height: 1.5;
	}
	.features li::before {
		content: '—';
		position: absolute;
		left: 0;
		color: rgba(255, 255, 255, 0.15);
	}

	/* ── Token model ── */
	.token-section {
		max-width: 700px;
		margin: 2rem auto 0;
		padding: 0 2rem 2rem;
	}

	.token-card {
		padding: 2rem 2.25rem;
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.token-card h2   { font-size: 1.2rem; font-weight: 700; }
	.token-sub       { font-size: 0.875rem; color: var(--muted); line-height: 1.6; margin-top: -0.5rem; }

	.token-table {
		border: 1px solid rgba(255, 255, 255, 0.06);
		border-radius: var(--radius);
		overflow: hidden;
	}

	.token-row {
		display: grid;
		grid-template-columns: 1fr auto auto;
		gap: 1.5rem;
		padding: 0.8rem 1.25rem;
		font-size: 0.85rem;
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
		align-items: center;
	}
	.token-row:last-child { border-bottom: none; }
	.token-row.header {
		font-size: 0.7rem;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: var(--muted);
		background: rgba(0, 0, 0, 0.2);
	}
	.token-row code                  { color: var(--accent); }
	.token-row span:nth-child(2)     { color: var(--text); font-weight: 600; text-align: right; }
	.token-row span:nth-child(3)     { color: var(--muted); white-space: nowrap; text-align: right; }

	/* ── OSS callout ── */
	.oss-section {
		max-width: 700px;
		margin: 2rem auto 0;
		padding: 0 2rem 2rem;
	}

	.oss-card {
		padding: 1.75rem 2.25rem;
	}

	.oss-body {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 2rem;
		flex-wrap: wrap;
	}

	.oss-title {
		font-size: 0.95rem;
		font-weight: 700;
		color: var(--text);
		margin-bottom: 0.35rem;
	}

	.oss-desc {
		font-size: 0.83rem;
		color: var(--muted);
		line-height: 1.6;
		max-width: 380px;
	}

	/* ── Auth ── */
	.auth-section {
		display: flex;
		justify-content: center;
		padding: 2rem 2rem 10rem;
	}

	.auth-card {
		padding: 2.5rem;
		width: 100%;
		max-width: 420px;
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
	}

	.auth-card h2 { font-size: 1.35rem; font-weight: 700; }
	.auth-sub     { font-size: 0.85rem; color: var(--muted); line-height: 1.5; margin-top: -0.5rem; }

	.tabs {
		display: flex;
		background: rgba(0, 0, 0, 0.3);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 3px;
	}

	.tab {
		flex: 1;
		background: none;
		border: none;
		border-radius: calc(var(--radius) - 2px);
		color: var(--muted);
		font-size: 0.85rem;
		font-weight: 500;
		padding: 0.4rem 0;
		cursor: pointer;
		transition: background 0.15s, color 0.15s;
	}
	.tab.active { background: rgba(255, 255, 255, 0.06); color: var(--text); }

	.submit-btn { width: 100%; justify-content: center; padding: 0.7rem; }
</style>
