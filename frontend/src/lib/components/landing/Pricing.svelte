<script lang="ts">
	import { t, locale } from '$lib/i18n';

	const TOKEN_RATE = 0.003;
	const PRESETS    = [500, 2000, 10000, 50000];

	let qty = 5000;

	$: numLocale = $locale === 'en' ? 'en-US' : 'fr-FR';
	$: price    = qty * TOKEN_RATE;
	$: priceStr = price.toLocaleString(numLocale, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	$: qtyStr   = qty.toLocaleString(numLocale);
	$: itinStr  = Math.floor(qty / 3).toLocaleString(numLocale);

	function handleQtyInput(e: Event) {
		const v = parseInt((e.target as HTMLInputElement).value);
		if (!isNaN(v) && v > 0) qty = v;
	}
</script>

<section class="pricing-section" id="pricing">
	<div class="wrap">
		<div class="section-head">
			<span class="eyebrow">{$t('pricing_eyebrow')}</span>
			<h2>{$t('pricing_title')}</h2>
			<p class="section-sub">{$t('pricing_sub')}</p>
		</div>

		<div class="at-cost-banner">
			<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M12 3v6M12 15v6M3 12h6M15 12h6M6 6l3 3M15 15l3 3M18 6l-3 3M9 15l-3 3"/></svg>
			<div>
				<strong>{$t('pricing_banner_title')}</strong>
				<span>{$t('pricing_banner_body')}</span>
			</div>
		</div>

		<div class="calc">
			<div class="calc-left">
				<div class="calc-label">{$t('pricing_how_many')}</div>
				<input type="range" class="calc-slider" min="100" max="100000" step="100" bind:value={qty} />
				<div class="calc-presets">
					{#each PRESETS as n}
						<button class="calc-preset" class:active={qty === n} on:click={() => (qty = n)}>
							{n.toLocaleString(numLocale)}
						</button>
					{/each}
					<input
						type="number"
						class="calc-input"
						value={qty}
						min="1"
						on:change={handleQtyInput}
						placeholder="…"
					/>
				</div>
				<ul class="calc-meta">
					<li><span>{$t('pricing_meta_calls')}</span><strong>{qtyStr}</strong></li>
					<li><span>{$t('pricing_meta_itin')}</span><strong>{itinStr}</strong></li>
					<li><span>{$t('pricing_meta_rate')}</span><strong>0,30 ct</strong></li>
					<li><span>{$t('pricing_meta_expiry')}</span><strong>{$t('pricing_meta_never')}</strong></li>
				</ul>
			</div>
			<div class="calc-right">
				<div class="calc-total-label">{$t('pricing_you_pay')}</div>
				<div class="calc-total">{priceStr}<span>€</span></div>
				<div class="calc-total-sub">{$t('pricing_once')}</div>
				<a href="/login" class="btn-primary big calc-cta">
					{$t('pricing_buy', { qty: qtyStr })}
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M5 12h14M13 6l6 6-6 6"/></svg>
				</a>
			</div>
		</div>

		<div class="pricing-aside">
			<div class="aside-card">
				<div class="aside-tag">{$t('pricing_free_tag')}</div>
				<strong>{$t('pricing_free_title')}</strong>
				<span>{$t('pricing_free_desc')}</span>
			</div>
			<div class="aside-card">
				<div class="aside-tag">{$t('pricing_self_tag')}</div>
				<strong>{$t('pricing_self_title')}</strong>
				<span>{$t('pricing_self_desc')}</span>
			</div>
			<div class="aside-card">
				<div class="aside-tag">{$t('pricing_enterprise_tag')}</div>
				<strong>{$t('pricing_enterprise_title')}</strong>
				<span>{$t('pricing_enterprise_desc')} <a href="https://github.com/ulysse-mercadal/trippier-public-API/issues" target="_blank" rel="noopener">{$t('pricing_enterprise_link')}</a>.</span>
			</div>
		</div>
	</div>
</section>

<style>
	.pricing-section { padding-block: 96px; }
	.at-cost-banner {
		display: flex;
		align-items: flex-start;
		gap: 12px;
		padding: 16px 20px;
		margin-bottom: 28px;
		border: 1px dashed color-mix(in oklch, var(--accent) 30%, var(--border));
		border-radius: var(--r-md);
		background: color-mix(in oklch, var(--accent) 4%, transparent);
		font-size: 13px;
		line-height: 1.5;
		color: var(--text-2);
	}
	.at-cost-banner :global(svg) { color: var(--accent); margin-top: 3px; flex-shrink: 0; }
	.at-cost-banner strong { color: var(--text); font-weight: 600; margin-right: 4px; }
	.calc {
		display: grid;
		grid-template-columns: 1.6fr 1fr;
		border: 1px solid var(--border);
		border-radius: var(--r-lg);
		background: var(--bg-2);
		overflow: hidden;
	}
	.calc-left {
		padding: 32px 32px 28px;
		display: flex;
		flex-direction: column;
		gap: 16px;
		border-right: 1px solid var(--border);
	}
	.calc-label {
		font-family: var(--font-mono);
		font-size: 11px;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--text-3);
	}
	.calc-slider {
		-webkit-appearance: none;
		appearance: none;
		width: 100%;
		height: 6px;
		background: var(--surface-2);
		border-radius: 999px;
		outline: none;
		cursor: pointer;
	}
	.calc-slider::-webkit-slider-thumb {
		-webkit-appearance: none;
		width: 22px; height: 22px;
		border-radius: 50%;
		background: var(--accent);
		border: 3px solid var(--bg-2);
		box-shadow: 0 0 0 1px color-mix(in oklch, var(--accent) 40%, transparent), 0 8px 18px -8px var(--accent);
		cursor: grab;
		transition: transform .1s ease;
	}
	.calc-slider::-webkit-slider-thumb:active { cursor: grabbing; transform: scale(1.1); }
	.calc-slider::-moz-range-thumb {
		width: 22px; height: 22px;
		border-radius: 50%;
		background: var(--accent);
		border: 3px solid var(--bg-2);
		box-shadow: 0 0 0 1px color-mix(in oklch, var(--accent) 40%, transparent);
		cursor: grab;
	}
	.calc-presets { display: flex; gap: 6px; flex-wrap: wrap; align-items: center; }
	.calc-preset {
		font-family: var(--font-mono);
		font-size: 12px;
		padding: 6px 12px;
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: 999px;
		color: var(--text-2);
		cursor: pointer;
		transition: border-color .15s ease, color .15s ease, background .15s ease;
	}
	.calc-preset:hover { border-color: var(--border-2); color: var(--text); }
	.calc-preset.active {
		border-color: var(--accent);
		background: color-mix(in oklch, var(--accent) 10%, transparent);
		color: var(--accent);
	}
	.calc-input {
		font-family: var(--font-mono);
		font-size: 12px;
		padding: 6px 12px;
		width: 90px;
		background: var(--bg);
		border: 1px solid var(--border);
		border-radius: 999px;
		color: var(--text);
		outline: none;
		transition: border-color .15s ease;
		-moz-appearance: textfield;
	}
	.calc-input::-webkit-outer-spin-button,
	.calc-input::-webkit-inner-spin-button { -webkit-appearance: none; margin: 0; }
	.calc-input:focus { border-color: var(--accent); }
	.calc-meta {
		list-style: none;
		padding: 16px 0 0;
		margin: 12px 0 0;
		border-top: 1px solid var(--border);
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 12px 24px;
	}
	.calc-meta li {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		gap: 12px;
		font-family: var(--font-mono);
		font-size: 12px;
	}
	.calc-meta span { color: var(--text-3); }
	.calc-meta strong { color: var(--text); font-weight: 500; font-feature-settings: "tnum"; }
	.calc-right {
		padding: 32px 32px 28px;
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: flex-start;
		gap: 6px;
		background:
			radial-gradient(80% 60% at 100% 0%, color-mix(in oklch, var(--accent) 7%, transparent), transparent 60%),
			var(--bg-2);
	}
	.calc-total-label { font-family: var(--font-mono); font-size: 11px; letter-spacing: 0.06em; text-transform: uppercase; color: var(--text-3); }
	.calc-total {
		font-size: 64px;
		font-weight: 600;
		letter-spacing: -0.04em;
		color: var(--accent);
		line-height: 1;
		font-feature-settings: "tnum";
		margin-top: 4px;
	}
	.calc-total span { font-size: 28px; color: var(--text-3); margin-left: 4px; }
	.calc-total-sub { font-family: var(--font-mono); font-size: 11.5px; color: var(--text-3); margin-bottom: 18px; }
	.calc-cta { width: 100%; justify-content: center; }
	.pricing-aside {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 12px;
		margin-top: 20px;
	}
	.aside-card {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 18px 20px;
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		background: color-mix(in oklch, var(--surface) 40%, transparent);
	}
	.aside-tag { font-family: var(--font-mono); font-size: 10.5px; text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-3); margin-bottom: 4px; }
	.aside-card strong { font-size: 16px; font-weight: 600; color: var(--text); letter-spacing: -0.01em; }
	.aside-card span { font-size: 13px; color: var(--text-2); line-height: 1.5; }
	.aside-card a { color: var(--accent); text-decoration: underline; text-underline-offset: 3px; }

	@media (max-width: 980px) {
		.calc { grid-template-columns: 1fr; }
		.calc-left { border-right: none; border-bottom: 1px solid var(--border); }
		.pricing-aside { grid-template-columns: 1fr; }
	}
</style>
