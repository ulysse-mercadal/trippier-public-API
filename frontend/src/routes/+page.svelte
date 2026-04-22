<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth';
	import { register, login, getMe, ApiError } from '$lib/api';

	let mode: 'login' | 'register' = 'register';
	let email = '';
	let password = '';
	let error = '';
	let success = '';
	let loading = false;

	onMount(async () => {
		const stored = auth.getStoredToken();
		if (!stored) return;
		try {
			const user = await getMe(stored);
			auth.init(stored, user);
			goto('/dashboard');
		} catch {
			// token expired — stay on landing
		}
	});

	async function submit() {
		error = '';
		success = '';
		loading = true;
		try {
			if (mode === 'register') {
				await register(email, password);
				success = 'Account created! Check your inbox to verify, then sign in.';
				mode = 'login';
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
	<p class="eyebrow">Public API · Beta</p>
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
		1 000 free tokens on signup · no credit card
	</div>
</section>

<!-- ── APIs ─────────────────────────────────────────────────────────────────── -->
<section class="apis" id="apis">
	<div class="api-card">
		<div class="api-header">
			<span class="api-icon">📍</span>
			<div>
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
			<code>GET /pois/search?lat=48.85&lng=2.35&radius=2000</code>
		</div>
		<ul class="features">
			<li>Multi-source aggregation with deduplication</li>
			<li>Type weights &amp; relevance scoring [0 – 100]</li>
			<li>Radius · polygon · district search modes</li>
			<li>Parallel provider fetching, Redis-cached</li>
		</ul>
	</div>

	<div class="api-card">
		<div class="api-header">
			<span class="api-icon">🗺️</span>
			<div>
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
<section class="token-model">
	<h2>Simple token model</h2>
	<p class="token-sub">Every account starts with <strong>1 000 tokens</strong>. The bucket refills every hour.</p>

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
</section>

<!-- ── Auth ──────────────────────────────────────────────────────────────────── -->
<section class="auth-section" id="auth">
	<div class="auth-card">
		<h2>{mode === 'register' ? 'Start for free' : 'Welcome back'}</h2>
		<p class="auth-sub">
			{mode === 'register'
				? '1 000 tokens, no credit card. Create your key in 30 seconds.'
				: 'Sign in to manage your API keys.'}
		</p>

		<div class="tabs">
			<button class="tab" class:active={mode === 'register'} on:click={() => { mode = 'register'; error = ''; }}>Create account</button>
			<button class="tab" class:active={mode === 'login'} on:click={() => { mode = 'login'; error = ''; }}>Sign in</button>
		</div>

		{#if error}<p class="alert alert-error">{error}</p>{/if}
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
		padding: 8rem 2rem 5rem;
		max-width: 740px;
		margin: 0 auto;
	}
	.eyebrow {
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--accent);
		margin-bottom: 1.25rem;
	}
	h1 {
		font-size: clamp(2.6rem, 6vw, 4.2rem);
		font-weight: 800;
		letter-spacing: -0.04em;
		line-height: 1.05;
		color: var(--text);
		margin-bottom: 1.5rem;
	}
	.tagline {
		color: var(--muted);
		font-size: 1.1rem;
		line-height: 1.75;
		max-width: 500px;
		margin: 0 auto 2.5rem;
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
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: 1.25rem;
		max-width: 900px;
		margin: 0 auto;
		padding: 2rem 2rem 5rem;
	}
	.api-card {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: calc(var(--radius) * 1.5);
		padding: 1.75rem;
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.api-header {
		display: flex;
		align-items: flex-start;
		gap: 0.875rem;
	}
	.api-icon { font-size: 1.4rem; line-height: 1; margin-top: 2px; }
	.api-header > div { flex: 1; }
	.api-header h3 { font-size: 1rem; font-weight: 700; margin-bottom: 0.15rem; }
	.api-sub { font-size: 0.8rem; color: var(--muted); }
	.badge {
		background: var(--accent-glow);
		color: var(--accent);
		border: 1px solid rgba(16,185,129,0.3);
		border-radius: 999px;
		padding: 0.2rem 0.65rem;
		font-size: 0.72rem;
		font-weight: 600;
		white-space: nowrap;
	}
	.badge-amber {
		background: rgba(245,158,11,0.1);
		color: #f59e0b;
		border-color: rgba(245,158,11,0.3);
	}
	.api-desc { font-size: 0.875rem; color: var(--muted); line-height: 1.65; }
	.endpoint-row {
		background: var(--bg);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 0.6rem 0.875rem;
		overflow-x: auto;
	}
	.endpoint-row code { font-size: 0.78rem; color: var(--accent); }
	.features {
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
	}
	.features li {
		font-size: 0.82rem;
		color: var(--muted);
		padding-left: 1rem;
		position: relative;
	}
	.features li::before {
		content: '—';
		position: absolute;
		left: 0;
		color: var(--border);
	}

	/* ── Token model ── */
	.token-model {
		max-width: 600px;
		margin: 0 auto;
		padding: 1rem 2rem 5rem;
		text-align: center;
	}
	.token-model h2 { font-size: 1.5rem; font-weight: 700; margin-bottom: 0.5rem; }
	.token-sub { color: var(--muted); font-size: 0.9rem; margin-bottom: 2rem; }
	.token-table {
		border: 1px solid var(--border);
		border-radius: var(--radius);
		overflow: hidden;
		text-align: left;
	}
	.token-row {
		display: grid;
		grid-template-columns: 1fr auto auto;
		gap: 1rem;
		padding: 0.75rem 1.25rem;
		font-size: 0.85rem;
		border-bottom: 1px solid var(--border);
		align-items: center;
	}
	.token-row:last-child { border-bottom: none; }
	.token-row.header { font-size: 0.72rem; text-transform: uppercase; letter-spacing: 0.07em; color: var(--muted); background: var(--surface); }
	.token-row code { color: var(--accent); }
	.token-row span:nth-child(2) { color: var(--text); font-weight: 600; text-align: right; }
	.token-row span:nth-child(3) { color: var(--muted); white-space: nowrap; text-align: right; }

	/* ── Auth ── */
	.auth-section {
		display: flex;
		justify-content: center;
		padding: 1rem 1rem 8rem;
	}
	.auth-card {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: calc(var(--radius) * 1.5);
		padding: 2.25rem;
		width: 100%;
		max-width: 400px;
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
	}
	.auth-card h2 { font-size: 1.3rem; font-weight: 700; }
	.auth-sub { font-size: 0.85rem; color: var(--muted); line-height: 1.5; margin-top: -0.5rem; }
	.tabs {
		display: flex;
		gap: 0;
		background: var(--bg);
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
	.tab.active { background: var(--surface-2); color: var(--text); }
	.submit-btn { width: 100%; justify-content: center; padding: 0.7rem; }
</style>
