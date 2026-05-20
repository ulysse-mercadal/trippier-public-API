<script lang="ts">
	import { t } from '$lib/i18n';
</script>

<footer class="foot">
	<div class="foot-grid">
		<div class="foot-brand">
			<div class="brand-inline">
				<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" style="color:var(--accent)">
					<path d="M4 18c4-8 12-8 16 0"/>
					<path d="M6 14c3-5 9-5 12 0" opacity="0.6"/>
					<circle cx="12" cy="20" r="1.2" fill="currentColor"/>
				</svg>
				<span>trippier<span class="brand-dot">/</span>api</span>
			</div>
			<p>{$t('footer_tagline')}</p>
		</div>
		<div>
			<h4>{$t('footer_product')}</h4>
			<a href="/#features">{$t('footer_features')}</a>
			<a href="/docs">{$t('footer_docs')}</a>
			<a href="/#pricing">{$t('footer_tokens')}</a>
			<a href="https://github.com/ulysse-mercadal/trippier-public-API/releases" target="_blank" rel="noopener">{$t('footer_changelog')}</a>
		</div>
		<div>
			<h4>{$t('footer_community')}</h4>
			<a href="https://github.com/ulysse-mercadal/trippier-public-API" target="_blank" rel="noopener">{$t('footer_github')}</a>
			<a href="/roadmap">{$t('footer_roadmap')}</a>
			<a href="https://github.com/ulysse-mercadal/trippier-public-API/blob/main/CONTRIBUTING.md" target="_blank" rel="noopener">{$t('footer_contribute')}</a>
			<a href="https://github.com/ulysse-mercadal/trippier-public-API/issues" target="_blank" rel="noopener">{$t('footer_issues')}</a>
		</div>
		<div>
			<h4>{$t('footer_app')}</h4>
			<a href="#">{$t('footer_ios')}</a>
			<a href="#">{$t('footer_android')}</a>
			<a href="#">{$t('footer_web')}</a>
			<a href="https://github.com/ulysse-mercadal/trippier-public-API" target="_blank" rel="noopener">{$t('footer_status')}</a>
		</div>
	</div>
	<div class="foot-bottom">
		<span>© 2026 trippier · MIT License</span>
		<span>46.7°N · 4.8°E</span>
	</div>
</footer>

<style>
	.foot {
		border-top: 1px solid var(--border);
		margin-top: 80px;
		padding-block: 56px 32px;
		position: relative;
		z-index: 1;
	}
	.foot-grid {
		display: grid;
		grid-template-columns: 2fr 1fr 1fr 1fr;
		gap: 48px;
		max-width: var(--max);
		margin: 0 auto;
		padding-inline: 32px;
	}
	.brand-inline {
		display: inline-flex;
		align-items: center;
		gap: 10px;
		font-weight: 600;
		font-size: 16px;
		color: var(--text);
		letter-spacing: -0.01em;
	}
	.brand-dot { color: var(--text-3); margin: 0 1px; }
	.foot-brand p { color: var(--text-3); font-size: 13px; margin: 14px 0 0; max-width: 280px; line-height: 1.5; }
	.foot-grid h4 {
		font-size: 12px;
		font-family: var(--font-mono);
		font-weight: 500;
		color: var(--text-3);
		letter-spacing: 0.06em;
		text-transform: uppercase;
		margin: 0 0 16px;
	}
	.foot-grid a {
		display: block;
		color: var(--text-2);
		font-size: 13.5px;
		padding: 5px 0;
		transition: color .12s ease;
	}
	.foot-grid a:hover { color: var(--accent); }
	.foot-bottom {
		max-width: var(--max);
		margin: 48px auto 0;
		padding: 24px 32px 0;
		border-top: 1px solid var(--border);
		display: flex;
		justify-content: space-between;
		font-family: var(--font-mono);
		font-size: 11.5px;
		color: var(--text-4);
	}

	@media (max-width: 980px) {
		.foot-grid { grid-template-columns: 1fr 1fr; padding-inline: 20px; }
	}
</style>
