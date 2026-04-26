<svelte:head>
	<title>Trippier API — Travel data for builders</title>
</svelte:head>

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
			Open source · MIT
		</a>
	</div>

	<h1>Travel data,<br />for builders.</h1>

	<p class="tagline">
		Search millions of points of interest and generate itineraries with a single API call.
		Try it free, then plug it straight into your stack.
	</p>

	<div class="cta-row">
		<a href="/login" class="btn btn-primary">Get 1 000 free tokens</a>
		<a href="https://github.com/ulysse-mercadal/trippier-public-API" class="btn btn-ghost" target="_blank" rel="noopener noreferrer">View on GitHub →</a>
	</div>
</section>

<!-- ── Below fold ────────────────────────────────────────────────────────────── -->
<div class="below-fold">

	<!-- ── APIs ─────────────────────────────────────────────────────────────── -->
	<section class="apis" id="apis">
		<div class="section-intro">
			<h2>Two APIs. Every destination.</h2>
			<p>One key, two endpoints. Everything you need to build travel features into your product.</p>
		</div>

		<div class="api-grid">
			<div class="api-card">
				<div class="api-header">
					<div class="api-icon">
						<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
							<path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/>
							<circle cx="12" cy="10" r="3"/>
						</svg>
					</div>
					<div class="api-title">
						<h3>POI Search</h3>
						<p class="api-sub">Find anything, anywhere on Earth</p>
					</div>
				</div>
				<p class="api-desc">
					Hotels, restaurants, museums, parks — aggregated, deduplicated and scored
					from multiple open data sources. One call, clean results.
				</p>
				<div class="endpoint-row">
					<code>GET /pois/search?lat=48.85&amp;lng=2.35&amp;radius=2000</code>
				</div>
				<ul class="features">
					<li>Radius, polygon or district search modes</li>
					<li>Relevance scores from 0 to 100</li>
					<li>Parallel fetch + Redis cache</li>
					<li>Full JSON, ready to display</li>
				</ul>
			</div>

			<div class="api-card">
				<div class="api-header">
					<div class="api-icon">
						<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
							<circle cx="6" cy="19" r="3"/>
							<circle cx="18" cy="5" r="3"/>
							<path d="M12 19h4.5a3.5 3.5 0 0 0 0-7h-8a3.5 3.5 0 0 1 0-7H12"/>
						</svg>
					</div>
					<div class="api-title">
						<h3>Itinerary</h3>
						<p class="api-sub">Day-by-day trips, zero boilerplate</p>
					</div>
				</div>
				<p class="api-desc">
					Send a destination and constraints — days, pace, transport. Get back
					a structured day-by-day schedule ready to render.
				</p>
				<div class="endpoint-row">
					<code>POST /itinerary/generate</code>
				</div>
				<ul class="features">
					<li>Configurable pace & transport mode</li>
					<li>Opening hours & proximity aware</li>
					<li>Powered by the POI Search API</li>
					<li>JSON output ready for your UI</li>
				</ul>
			</div>
		</div>
	</section>

	<!-- ── Try it ────────────────────────────────────────────────────────────── -->
	<section class="try-section">
		<div class="try-inner">
			<div class="try-text">
				<h2>1 000 tokens to try it.</h2>
				<p>
					Create a free account, get your API key in 30 seconds, and start making
					calls right away. No credit card, no setup — just add the key to your
					existing code.
				</p>
				<div class="try-snippet">
					<span class="snippet-label">Drop it in your existing code</span>
					<pre><code>curl "https://api.trippier.dev/pois/search?lat=48.85&lng=2.35&radius=1000" \
  -H "X-API-Key: YOUR_KEY"</code></pre>
				</div>
			</div>
			<a href="/login" class="btn btn-primary try-cta">Get your free key</a>
		</div>
	</section>

	<!-- ── Self-host ─────────────────────────────────────────────────────────── -->
	<section class="selfhost-section">
		<div class="section-intro">
			<h2>Or run it yourself.</h2>
			<p>The full stack is open source and ships as Docker images. One command and it's running on your infra.</p>
		</div>

		<div class="selfhost-steps">
			<div class="step">
				<div class="step-num">01</div>
				<h3>Clone the repo</h3>
				<div class="step-code">
					<code>git clone https://github.com/ulysse-mercadal/trippier-public-API</code>
				</div>
			</div>
			<div class="step-arrow">→</div>
			<div class="step">
				<div class="step-num">02</div>
				<h3>Pull the images</h3>
				<div class="step-code">
					<code>ghcr.io/ulysse-mercadal/trippier-auth-api</code>
					<code>ghcr.io/ulysse-mercadal/trippier-poi-api</code>
				</div>
			</div>
			<div class="step-arrow">→</div>
			<div class="step">
				<div class="step-num">03</div>
				<h3>Start the stack</h3>
				<div class="step-code">
					<code>docker compose up</code>
				</div>
			</div>
		</div>

		<div class="selfhost-links">
			<a href="https://github.com/ulysse-mercadal/trippier-public-API" class="btn btn-ghost" target="_blank" rel="noopener noreferrer">
				<svg width="15" height="15" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true"><path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0 0 16 8c0-4.42-3.58-8-8-8z"/></svg>
				GitHub repo
			</a>
			<a href="https://github.com/ulysse-mercadal/trippier-public-API/pkgs/container/trippier-poi-api" class="btn btn-ghost" target="_blank" rel="noopener noreferrer">
				<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18M9 21V9"/></svg>
				Docker images
			</a>
		</div>
	</section>

	<!-- ── Providers ─────────────────────────────────────────────────────────── -->
	<section class="providers-section">
		<div class="section-intro">
			<h2>Powered by the best open data</h2>
			<p>
				We aggregate across multiple authoritative sources, merge duplicates, and score results —
				always in full compliance with each provider's Terms of Service.
			</p>
		</div>

		<div class="providers-grid">
			<div class="provider-item">
				<span class="provider-dot active"></span>
				<div class="provider-text">
					<strong>OpenStreetMap</strong>
					<span>ODbL — open, community-maintained</span>
				</div>
			</div>
			<div class="provider-item">
				<span class="provider-dot active"></span>
				<div class="provider-text">
					<strong>Wikipedia</strong>
					<span>CC BY-SA — rich landmark descriptions</span>
				</div>
			</div>
			<div class="provider-item">
				<span class="provider-dot active"></span>
				<div class="provider-text">
					<strong>Wikivoyage</strong>
					<span>CC BY-SA — curated travel guides</span>
				</div>
			</div>
			<div class="provider-item">
				<span class="provider-dot active"></span>
				<div class="provider-text">
					<strong>GeoNames</strong>
					<span>CC BY — geographic name database</span>
				</div>
			</div>
			<div class="provider-item provider-future">
				<span class="provider-dot"></span>
				<div class="provider-text">
					<strong>Google Places</strong>
					<span>Planned — ToS integration in progress</span>
				</div>
			</div>
			<div class="provider-item provider-future">
				<span class="provider-dot"></span>
				<div class="provider-text">
					<strong>Foursquare</strong>
					<span>Planned — ToS integration in progress</span>
				</div>
			</div>
		</div>

		<p class="tos-note">
			<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true" style="flex-shrink:0;margin-top:1px"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
			Each data source is integrated in strict compliance with its Terms of Service. Attributions are preserved and propagated in every API response.
		</p>
	</section>

	<!-- ── Bottom CTA ────────────────────────────────────────────────────────── -->
	<section class="bottom-cta">
		<h2>Ready to build?</h2>
		<p>Get 1 000 tokens to explore the API, then self-host for free with your own stack.</p>
		<div class="cta-row">
			<a href="/login" class="btn btn-primary">Get started — it's free</a>
			<a href="https://github.com/ulysse-mercadal/trippier-public-API" class="btn btn-ghost" target="_blank" rel="noopener noreferrer">View on GitHub</a>
		</div>
	</section>

</div>

<style>
	/* ── Hero ── */
	.hero {
		text-align: center;
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 6rem 2rem 5rem;
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
		color: #4b5563;
		border: 1px solid #d1d5db;
		background: rgba(0, 0, 0, 0.04);
		border-radius: 999px;
		padding: 0.3rem 0.8rem;
	}

	.pill-oss {
		color: var(--accent);
		border-color: rgba(16, 185, 129, 0.4);
		background: rgba(16, 185, 129, 0.08);
		transition: opacity 0.15s;
	}
	.pill-oss:hover { opacity: 0.8; }

	h1 {
		font-size: clamp(2.8rem, 6.5vw, 4.4rem);
		font-weight: 800;
		letter-spacing: -0.04em;
		line-height: 1.05;
		color: #0a0a0a;
		margin-bottom: 1.5rem;
	}

	.tagline {
		color: #4b5563;
		font-size: 1.1rem;
		line-height: 1.8;
		max-width: 520px;
		margin: 0 auto 2.75rem;
	}

	.cta-row {
		display: flex;
		gap: 0.75rem;
		justify-content: center;
		flex-wrap: wrap;
	}

	/* ── Below fold ── */
	.below-fold {
		background: #0a0a0a;
		border-top: 1px solid rgba(0,0,0,0.12);
		position: relative;
		z-index: 1;
	}

	.section-intro {
		text-align: center;
		margin-bottom: 3rem;
	}
	.section-intro h2 {
		font-size: clamp(1.5rem, 3vw, 2rem);
		font-weight: 800;
		letter-spacing: -0.03em;
		margin-bottom: 0.6rem;
	}
	.section-intro p {
		font-size: 0.95rem;
		color: var(--muted);
		max-width: 520px;
		margin: 0 auto;
		line-height: 1.7;
	}

	/* ── APIs ── */
	.apis {
		max-width: 1000px;
		margin: 0 auto;
		padding: 5rem 2rem 3rem;
	}

	.api-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
		gap: 1rem;
	}

	.api-card {
		background: #111;
		border: 1px solid #1e1e1e;
		border-radius: 16px;
		padding: 2rem 2.25rem;
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
	}

	.api-header {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
	}

	.api-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 34px;
		height: 34px;
		border-radius: 8px;
		background: rgba(16, 185, 129, 0.12);
		border: 1px solid rgba(16, 185, 129, 0.2);
		color: var(--accent);
		flex-shrink: 0;
	}

	.api-title h3  { font-size: 1.05rem; font-weight: 700; margin-bottom: 0.2rem; }
	.api-sub       { font-size: 0.8rem; color: var(--muted); }
	.api-desc      { font-size: 0.875rem; color: var(--muted); line-height: 1.7; }

	.endpoint-row {
		background: rgba(0,0,0,0.4);
		border: 1px solid #1a1a1a;
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
		color: rgba(255,255,255,0.12);
	}

	/* ── Try it ── */
	.try-section {
		border-top: 1px solid #141414;
		border-bottom: 1px solid #141414;
		background: rgba(16, 185, 129, 0.03);
		padding: 4rem 2rem;
	}

	.try-inner {
		max-width: 1000px;
		margin: 0 auto;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 3rem;
		flex-wrap: wrap;
	}

	.try-text {
		flex: 1;
		min-width: 260px;
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.try-text h2 {
		font-size: clamp(1.4rem, 2.5vw, 1.9rem);
		font-weight: 800;
		letter-spacing: -0.03em;
	}

	.try-text > p {
		font-size: 0.9rem;
		color: var(--muted);
		line-height: 1.7;
		max-width: 480px;
	}

	.try-snippet {
		background: #0d0d0d;
		border: 1px solid #1e1e1e;
		border-radius: 10px;
		overflow: hidden;
	}

	.snippet-label {
		display: block;
		font-size: 0.72rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--muted);
		padding: 0.5rem 1rem;
		border-bottom: 1px solid #1a1a1a;
	}

	.try-snippet pre {
		padding: 0.875rem 1rem;
		overflow-x: auto;
	}

	.try-snippet code {
		font-size: 0.78rem;
		color: #e5e5e5;
		line-height: 1.7;
	}

	.try-cta { white-space: nowrap; flex-shrink: 0; }

	/* ── Self-host ── */
	.selfhost-section {
		max-width: 1000px;
		margin: 0 auto;
		padding: 4rem 2rem;
	}

	.selfhost-steps {
		display: flex;
		align-items: flex-start;
		gap: 0;
		flex-wrap: wrap;
		margin-bottom: 2rem;
	}

	.step {
		flex: 1;
		min-width: 180px;
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
	}

	.step-num {
		font-size: 0.65rem;
		font-weight: 800;
		letter-spacing: 0.12em;
		color: var(--accent);
		text-transform: uppercase;
	}

	.step h3 {
		font-size: 0.95rem;
		font-weight: 700;
	}

	.step-code {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
	}

	.step-code code {
		font-size: 0.75rem;
		color: var(--muted);
		background: #111;
		border: 1px solid #1e1e1e;
		border-radius: 6px;
		padding: 0.4rem 0.7rem;
		display: block;
		overflow-x: auto;
	}

	.step-arrow {
		color: #333;
		font-size: 1.4rem;
		padding: 1.8rem 1.5rem 0;
		flex-shrink: 0;
	}

	.selfhost-links {
		display: flex;
		gap: 0.75rem;
		flex-wrap: wrap;
	}

	/* ── Providers ── */
	.providers-section {
		max-width: 1000px;
		margin: 0 auto;
		padding: 1rem 2rem 4rem;
		border-top: 1px solid #141414;
	}

	.providers-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
		gap: 1px;
		background: #161616;
		border: 1px solid #161616;
		border-radius: 12px;
		overflow: hidden;
		margin-bottom: 1.25rem;
	}

	.provider-item {
		display: flex;
		align-items: center;
		gap: 0.875rem;
		background: #0a0a0a;
		padding: 1.1rem 1.25rem;
	}

	.provider-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: #2a2a2a;
		border: 1px solid #333;
		flex-shrink: 0;
	}

	.provider-dot.active {
		background: var(--accent);
		border-color: var(--accent);
		box-shadow: 0 0 6px rgba(16, 185, 129, 0.4);
	}

	.provider-text {
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
	}

	.provider-text strong { font-size: 0.85rem; font-weight: 600; color: var(--text); }
	.provider-text span   { font-size: 0.73rem; color: var(--muted); }
	.provider-future      { opacity: 0.45; }

	.tos-note {
		display: flex;
		align-items: flex-start;
		gap: 0.5rem;
		font-size: 0.75rem;
		color: var(--muted);
		line-height: 1.5;
	}

	/* ── Bottom CTA ── */
	.bottom-cta {
		text-align: center;
		padding: 5rem 2rem 8rem;
		max-width: 600px;
		margin: 0 auto;
	}

	.bottom-cta h2 {
		font-size: clamp(1.6rem, 3vw, 2.2rem);
		font-weight: 800;
		letter-spacing: -0.03em;
		margin-bottom: 0.75rem;
	}

	.bottom-cta p {
		font-size: 1rem;
		color: var(--muted);
		margin-bottom: 2rem;
		line-height: 1.6;
	}

	@media (max-width: 640px) {
		.try-inner  { flex-direction: column; }
		.try-cta    { width: 100%; justify-content: center; }
		.selfhost-steps { flex-direction: column; gap: 2rem; }
		.step-arrow { display: none; }
	}
</style>
