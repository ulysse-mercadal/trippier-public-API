<script lang="ts">
	import { browser } from '$app/environment';
	import { t } from '$lib/i18n';

	const dockerCmd = 'docker pull trippier/api:latest';

	let copied = false;

	/**
	 * Copies the docker pull command to the clipboard and briefly flags it as copied.
	 */
	function copyCmd() {
		if (browser) navigator.clipboard?.writeText(dockerCmd);
		copied = true;
		setTimeout(() => (copied = false), 1500);
	}
</script>

<section class="hero">
<h1 class="hero-title">
		{$t('hero_title')} <span class="accent-fg">{$t('hero_accent')}</span>
	</h1>
	<p class="hero-sub">{$t('hero_sub')}</p>
	<div class="hero-actions">
		<a href="/docs" class="btn-primary big">
			{$t('hero_try')}
			<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M5 12h14M13 6l6 6-6 6"/></svg>
		</a>
		<button class="btn-copy" on:click={copyCmd}>
			<span class="cprompt">$</span>
			<span class="ccmd">{dockerCmd}</span>
			<span class="copy-pill">
				{#if copied}
					<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12l5 5 9-11"/></svg>
					{$t('hero_copied')}
				{:else}
					{$t('hero_copy')}
				{/if}
			</span>
		</button>
	</div>
	<ul class="hero-meta">
		<li><strong>17</strong><span>{$t('hero_stat_sources')}</span></li>
		<li><strong>12</strong><span>{$t('hero_stat_routes')}</span></li>
		<li><strong>MIT</strong><span>{$t('hero_stat_license')}</span></li>
		<li><strong>0€</strong><span>{$t('hero_stat_selfhosted')}</span></li>
	</ul>
</section>

<style>
	.hero {
		padding: 80px 32px 110px;
		text-align: left;
		max-width: var(--max);
		margin: 0 auto;
	}
.hero-title {
		font-size: clamp(40px, 5.4vw, 76px);
		line-height: 1;
		letter-spacing: -0.035em;
		font-weight: 600;
		margin: 0 0 28px;
		max-width: 1000px;
		text-wrap: balance;
	}
	.hero-sub {
		font-size: 18px;
		line-height: 1.55;
		color: var(--text-2);
		max-width: 640px;
		margin: 0 0 36px;
	}
	.hero-actions {
		display: flex;
		gap: 14px;
		align-items: center;
		flex-wrap: wrap;
	}
	.btn-copy {
		display: inline-flex;
		align-items: center;
		gap: 12px;
		padding: 12px 16px;
		border: 1px solid var(--border);
		background: color-mix(in oklch, var(--surface) 70%, transparent);
		border-radius: var(--r-md);
		font-family: var(--font-mono);
		font-size: 13.5px;
		color: var(--text);
		cursor: pointer;
		transition: border-color .15s ease;
	}
	.btn-copy:hover { border-color: var(--accent); }
	.cprompt { color: var(--accent); }
	.ccmd    { color: var(--text-2); }
	.copy-pill {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 3px 8px;
		font-size: 11px;
		font-family: var(--font-sans);
		background: var(--surface-2);
		border-radius: 4px;
		color: var(--text-3);
		margin-left: 6px;
	}
	.hero-meta {
		display: flex;
		gap: 56px;
		list-style: none;
		padding: 32px 0 0;
		margin: 56px 0 0;
		border-top: 1px solid var(--border);
		flex-wrap: wrap;
	}
	.hero-meta li { display: flex; flex-direction: column; gap: 2px; }
	.hero-meta strong {
		font-size: 28px;
		font-weight: 500;
		letter-spacing: -0.02em;
		color: var(--text);
		font-feature-settings: "tnum";
	}
	.hero-meta span {
		font-size: 12px;
		font-family: var(--font-mono);
		color: var(--text-3);
		letter-spacing: 0.02em;
	}

	@media (max-width: 980px) {
		.hero-title { font-size: 44px; }
		.hero { padding-inline: 20px; }
	}
</style>
