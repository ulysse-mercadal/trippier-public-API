<script lang="ts">
	import { t } from '$lib/i18n';
</script>

<section class="features-section" id="features">
	<div class="wrap">
		<div class="section-head">
			<span class="eyebrow">{$t('feat_eyebrow')}</span>
			<h2>{$t('feat_title')}</h2>
		</div>
		<div class="feature-grid">
			<article class="feature">
				<header>
					<span class="feature-tag">01 · poi</span>
					<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="feature-icon"><path d="M12 21s7-6.5 7-12a7 7 0 1 0-14 0c0 5.5 7 12 7 12Z"/><circle cx="12" cy="9" r="2.5"/></svg>
				</header>
				<h3>{$t('feat_poi_title')}</h3>
				<p>{$t('feat_poi_desc')}</p>
				<code class="feature-code">GET /v1/pois/search?lat=…&lng=…</code>
			</article>
			<article class="feature">
				<header>
					<span class="feature-tag">02 · events</span>
					<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="feature-icon"><rect x="3" y="5" width="18" height="16" rx="2"/><path d="M3 10h18M8 3v4M16 3v4"/></svg>
				</header>
				<h3>{$t('feat_events_title')}</h3>
				<p>{$t('feat_events_desc')}</p>
				<code class="feature-code">GET /v1/pois/events?lat=…&lng=…</code>
			</article>
			<article class="feature">
				<header>
					<span class="feature-tag">03 · itinerary</span>
					<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="feature-icon"><circle cx="6" cy="6" r="2"/><circle cx="18" cy="18" r="2"/><path d="M8 6h6a4 4 0 0 1 0 8h-4a4 4 0 0 0 0 8h6"/></svg>
				</header>
				<h3>{$t('feat_itin_title')}</h3>
				<p>{$t('feat_itin_desc')}</p>
				<code class="feature-code">POST /v1/itinerary/generate</code>
			</article>
			<article class="feature">
				<header>
					<span class="feature-tag">04 · routing</span>
					<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="feature-icon"><circle cx="12" cy="12" r="3"/><path d="M12 2v3M12 19v3M2 12h3M19 12h3"/><path d="M5.6 5.6l2.1 2.1M16.3 16.3l2.1 2.1M5.6 18.4l2.1-2.1M16.3 7.7l2.1-2.1"/></svg>
				</header>
				<h3>{$t('feat_routing_title')}</h3>
				<p>{$t('feat_routing_desc')}</p>
				<code class="feature-code">GET /v1/pois/providers/recommend</code>
			</article>
		</div>
	</div>
</section>

<style>
	.features-section { padding-block: 96px; }
	.feature-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 1px;
		background: var(--border);
		border: 1px solid var(--border);
		border-radius: var(--r-lg);
		overflow: hidden;
	}
	.feature {
		background: var(--bg-2);
		padding: 32px 28px 28px;
		display: flex;
		flex-direction: column;
		gap: 14px;
		transition: background .2s ease;
	}
	.feature:hover { background: var(--surface); }
	.feature header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 12px;
	}
	.feature-tag {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--text-3);
		letter-spacing: 0.06em;
		text-transform: lowercase;
	}
	.feature-icon { color: var(--accent); }
	.feature h3 {
		font-size: 20px;
		margin: 0;
		font-weight: 600;
		letter-spacing: -0.01em;
		line-height: 1.2;
	}
	.feature p {
		color: var(--text-2);
		font-size: 14.5px;
		line-height: 1.55;
		margin: 0;
		flex: 1;
	}
	.feature-code {
		display: inline-block;
		font-family: var(--font-mono);
		font-size: 12px;
		color: var(--accent);
		padding: 8px 12px;
		background: color-mix(in oklch, var(--accent) 6%, var(--bg));
		border: 1px solid color-mix(in oklch, var(--accent) 15%, transparent);
		border-radius: var(--r-sm);
		width: fit-content;
		margin-top: 4px;
	}

	@media (max-width: 1200px) {
		.feature-grid { grid-template-columns: repeat(2, 1fr); }
	}
	@media (max-width: 640px) {
		.feature-grid { grid-template-columns: 1fr; }
	}
</style>
