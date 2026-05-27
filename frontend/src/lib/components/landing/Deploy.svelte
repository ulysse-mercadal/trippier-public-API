<script lang="ts">
	import { t } from '$lib/i18n';
</script>

<section class="deploy-section" id="deploy">
	<div class="wrap">
		<div class="section-head">
			<span class="eyebrow">{$t('deploy_eyebrow')}</span>
			<h2>{$t('deploy_title')}</h2>
			<p class="section-sub">{$t('deploy_sub')}</p>
		</div>
		<div class="deploy-grid">
			<article class="deploy-card hero-card">
				<div class="reco-flag">
					<span class="reco-dot"></span> {$t('deploy_recommended')}
				</div>
				<header>
					<svg width="22" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3" y="12" width="3" height="3"/><rect x="7" y="12" width="3" height="3"/><rect x="11" y="12" width="3" height="3"/><rect x="7" y="8" width="3" height="3"/><rect x="11" y="8" width="3" height="3"/><rect x="11" y="4" width="3" height="3"/><path d="M2 15c0 3 3 5 8 5 7 0 11-4 12-8 0 0-2 1-3-1-2 1-4 0-4 0"/></svg>
					<span class="deploy-kind">{$t('deploy_self_kind')}</span>
				</header>
				<h3>{$t('deploy_self_title')} <span class="accent-fg">{$t('deploy_self_accent')}</span></h3>
				<p>{$t('deploy_self_desc')}</p>
				<pre class="deploy-cmd">{`# 1. pull
docker pull trippier/api:latest

# 2. run
docker run -p 8080:8080 \\
  -v trippier-data:/data \\
  trippier/api

# 3. go
curl localhost:8080/v1/healthz`}</pre>
				<ul class="deploy-list compact">
					<li>
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12l5 5 9-11"/></svg>
						{$t('deploy_self_li1')}
					</li>
					<li>
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12l5 5 9-11"/></svg>
						{$t('deploy_self_li2')}
					</li>
					<li>
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12l5 5 9-11"/></svg>
						{$t('deploy_self_li3')}
					</li>
				</ul>
				<div class="card-cta-row">
					<a href="https://github.com/ulysse-mercadal/trippier-public-API" target="_blank" rel="noopener" class="btn-primary">
						{$t('deploy_self_guide')}
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M5 12h14M13 6l6 6-6 6"/></svg>
					</a>
					<a href="https://github.com/ulysse-mercadal/trippier-public-API/blob/main/docker-compose.simple.yml" target="_blank" rel="noopener" class="btn-ghost">docker-compose.yml</a>
				</div>
			</article>

			<article class="deploy-card secondary">
				<header>
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M7 18a4 4 0 0 1-1-7.8 6 6 0 0 1 11.6 1.8A4 4 0 0 1 17 18H7Z"/></svg>
					<span class="deploy-kind">{$t('deploy_cloud_kind')}</span>
				</header>
				<h3>{$t('deploy_cloud_title')} <span class="muted-fg">api.trippier.dev</span></h3>
				<p>{$t('deploy_cloud_desc')}</p>
				<ul class="deploy-list">
					<li>
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12l5 5 9-11"/></svg>
						<span><strong>100</strong> {$t('deploy_cloud_li1')}</span>
					</li>
					<li>
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12l5 5 9-11"/></svg>
						{$t('deploy_cloud_li2')}
					</li>
					<li>
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12l5 5 9-11"/></svg>
						{$t('deploy_cloud_li3')}
					</li>
				</ul>
				<div class="at-cost-note">
					<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M12 3v6M12 15v6M3 12h6M15 12h6M6 6l3 3M15 15l3 3M18 6l-3 3M9 15l-3 3"/></svg>
					<span>{$t('deploy_cloud_note')}</span>
				</div>
				<a href="#pricing" class="btn-ghost">
					{$t('deploy_cloud_cta')}
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M5 12h14M13 6l6 6-6 6"/></svg>
				</a>
			</article>
		</div>
	</div>
</section>

<style>
	.deploy-section { padding-block: 96px; }
	.deploy-grid {
		display: grid;
		grid-template-columns: 1.35fr 1fr;
		gap: 20px;
		align-items: stretch;
	}
	.deploy-card {
		border: 1px solid var(--border);
		border-radius: var(--r-lg);
		padding: 32px 28px;
		background: color-mix(in oklch, var(--surface) 50%, transparent);
		backdrop-filter: blur(4px);
		display: flex;
		flex-direction: column;
		gap: 16px;
		position: relative;
	}
	.deploy-card.hero-card {
		background:
			radial-gradient(120% 80% at 0% 0%, color-mix(in oklch, var(--accent) 8%, transparent), transparent 55%),
			linear-gradient(180deg, color-mix(in oklch, var(--accent) 4%, transparent), transparent 80%),
			color-mix(in oklch, var(--surface) 70%, transparent);
		border-color: color-mix(in oklch, var(--accent) 40%, var(--border));
		box-shadow:
			0 0 0 1px color-mix(in oklch, var(--accent) 12%, transparent) inset,
			0 30px 60px -30px color-mix(in oklch, var(--accent) 25%, transparent);
	}
	.deploy-card.secondary { opacity: 0.92; }
	.reco-flag {
		position: absolute;
		top: -1px; right: 22px;
		transform: translateY(-50%);
		display: inline-flex;
		align-items: center;
		gap: 8px;
		background: var(--accent);
		color: #08120e;
		font-size: 11px;
		font-weight: 600;
		padding: 5px 12px 5px 10px;
		border-radius: 999px;
		letter-spacing: 0.02em;
		text-transform: uppercase;
		box-shadow: 0 6px 18px -6px color-mix(in oklch, var(--accent) 50%, transparent);
	}
	.reco-dot { width: 6px; height: 6px; border-radius: 50%; background: #08120e; }
	.deploy-card header { display: flex; align-items: center; gap: 10px; color: var(--accent); }
	.deploy-card.secondary header { color: var(--text-2); }
	.deploy-kind {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--text-3);
		letter-spacing: 0.06em;
		text-transform: lowercase;
	}
	.deploy-card h3 {
		font-size: 28px;
		font-weight: 600;
		letter-spacing: -0.02em;
		margin: 0;
		line-height: 1.15;
	}
	.deploy-card.secondary h3 { font-size: 22px; }
	.deploy-card > p { color: var(--text-2); font-size: 14.5px; line-height: 1.55; margin: 0; }
	.deploy-list {
		list-style: none;
		padding: 0;
		margin: 8px 0 16px;
		display: flex;
		flex-direction: column;
		gap: 10px;
		font-size: 14px;
	}
	.deploy-list.compact { margin: 4px 0 0; gap: 8px; font-size: 13.5px; }
	.deploy-list li { display: flex; align-items: flex-start; gap: 10px; color: var(--text-2); }
	.deploy-list :global(svg) { color: var(--accent); margin-top: 3px; flex-shrink: 0; }
	.deploy-list strong { color: var(--text); font-weight: 600; }
	.deploy-cmd {
		background: var(--code-bg-2);
		border: 1px solid color-mix(in oklch, var(--accent) 12%, var(--code-border));
		border-radius: var(--r-md);
		padding: 16px 18px;
		font-family: var(--font-mono);
		font-size: 12.5px;
		color: var(--code-text-2);
		margin: 4px 0 8px;
		line-height: 1.6;
		white-space: pre;
		overflow-x: auto;
	}
	.card-cta-row { display: flex; gap: 10px; flex-wrap: wrap; margin-top: auto; }
	.at-cost-note {
		display: flex;
		align-items: flex-start;
		gap: 10px;
		padding: 12px 14px;
		background: color-mix(in oklch, var(--accent) 7%, transparent);
		border: 1px solid color-mix(in oklch, var(--accent) 18%, transparent);
		border-radius: var(--r-md);
		font-size: 12.5px;
		line-height: 1.45;
		color: var(--text-2);
	}
	.at-cost-note :global(svg) { color: var(--accent); margin-top: 1px; flex-shrink: 0; }

	@media (max-width: 980px) {
		.deploy-grid { grid-template-columns: 1fr; }
	}
</style>
