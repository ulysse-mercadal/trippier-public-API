<script lang="ts">
	import { t } from '$lib/i18n';

	const PROVIDERS = [
		{ name: 'OpenStreetMap',  kind: 'POI · free',      env: 'OSM_OVERPASS_URL'    },
		{ name: 'Wikivoyage',     kind: 'POI · free',      env: '(built-in)'          },
		{ name: 'Wikipedia',      kind: 'POI + Events',    env: '(built-in)'          },
		{ name: 'GeoNames',       kind: 'POI · optional',  env: 'GEONAMES_USERNAME'   },
		{ name: 'Ticketmaster',   kind: 'Events · BYOK',   env: 'X-Ticketmaster-Key'  },
		{ name: 'Eventbrite',     kind: 'Events · BYOK',   env: 'X-Eventbrite-Token'  },
		{ name: 'Meetup',         kind: 'Events · BYOK',   env: 'X-Meetup-Token'      },
		{ name: 'OpenAgenda',     kind: 'Events · BYOK',   env: 'X-OpenAgenda-Key'    },
		{ name: 'Foursquare',     kind: 'POI · BYOK',      env: 'X-Foursquare-Key'    },
		{ name: 'Baidu Maps',     kind: 'POI · BYOK (CN)', env: 'X-Baidu-Key'         },
		{ name: 'Kakao Maps',     kind: 'POI · BYOK (KR)', env: 'X-Kakao-Key'         },
		{ name: 'Navitime',       kind: 'POI · BYOK (JP)', env: 'X-Navitime-Key'      },
	];

	$: envLines = PROVIDERS.slice(0, 6).map(p => `${p.env.padEnd(24)} = sk_•••••••••  # ${p.kind}`).join('\n');
</script>

<section class="byok-section" id="byok">
	<div class="wrap">
		<div class="byok-grid">
			<div class="byok-copy">
				<span class="eyebrow">{$t('byok_eyebrow')}</span>
				<h2>{$t('byok_title')} <span class="accent-fg">{$t('byok_accent')}</span> {$t('byok_title_end')}</h2>
				<p class="section-sub">{$t('byok_sub')}</p>
				<ul class="byok-bullets">
					<li>
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12l5 5 9-11"/></svg>
						{$t('byok_li1')}
					</li>
					<li>
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12l5 5 9-11"/></svg>
						{$t('byok_li2')}
					</li>
					<li>
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12l5 5 9-11"/></svg>
						{$t('byok_li3')}
					</li>
					<li>
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12l5 5 9-11"/></svg>
						{$t('byok_li4')}
					</li>
				</ul>
			</div>
			<div class="byok-panel">
				<div class="byok-panel-head">
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="8" cy="15" r="4"/><path d="M10.8 12.2L20 3M16 7l3 3M14 9l3 3"/></svg>
					<span>.env</span>
					<span class="byok-spacer"></span>
					<span class="byok-status">
						<span class="byok-dot"></span>
						{PROVIDERS.length} {$t('byok_detected')}
					</span>
				</div>
				<pre class="byok-env">{envLines}</pre>
				<div class="byok-providers">
					{#each PROVIDERS as p}
						<div class="byok-prov">
							<span class="byok-prov-name">{p.name}</span>
							<span class="byok-prov-kind">{p.kind}</span>
						</div>
					{/each}
				</div>
			</div>
		</div>
	</div>
</section>

<style>
	.byok-section { padding-block: 96px; }
	.byok-grid {
		display: grid;
		grid-template-columns: 1fr 1.1fr;
		gap: 48px;
		align-items: start;
	}
	.byok-copy h2 {
		font-size: clamp(28px, 3.2vw, 44px);
		letter-spacing: -0.02em;
		line-height: 1.08;
		margin: 8px 0 0;
		font-weight: 600;
		text-wrap: balance;
	}
	.byok-bullets {
		list-style: none;
		padding: 0;
		margin: 24px 0 0;
		display: flex;
		flex-direction: column;
		gap: 12px;
		font-size: 14.5px;
	}
	.byok-bullets li { display: flex; align-items: flex-start; gap: 12px; color: var(--text-2); }
	.byok-bullets :global(svg) { color: var(--accent); margin-top: 3px; flex-shrink: 0; }
	.byok-panel {
		border: 1px solid var(--code-border);
		border-radius: var(--r-lg);
		background: var(--code-bg);
		overflow: hidden;
		box-shadow: 0 24px 60px -28px rgba(0,0,0,0.5);
	}
	.byok-panel-head {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 12px 16px;
		background: var(--code-surface);
		border-bottom: 1px solid var(--code-border);
		font-family: var(--font-mono);
		font-size: 12px;
		color: var(--code-text-2);
	}
	.byok-panel-head :global(svg) { color: var(--accent); }
	.byok-spacer { flex: 1; }
	.byok-status { display: inline-flex; align-items: center; gap: 6px; font-size: 11px; color: var(--code-text-3); }
	.byok-dot {
		width: 6px; height: 6px;
		border-radius: 50%;
		background: var(--accent);
		box-shadow: 0 0 0 3px color-mix(in oklch, var(--accent) 22%, transparent);
	}
	.byok-env {
		padding: 18px 20px 14px;
		margin: 0;
		font-family: var(--font-mono);
		font-size: 12px;
		line-height: 1.65;
		color: var(--code-text-2);
		white-space: pre;
		overflow-x: auto;
	}
	.byok-providers {
		border-top: 1px dashed var(--code-border);
		padding: 14px 18px;
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 8px 18px;
		background: color-mix(in oklch, var(--accent) 3%, transparent);
	}
	.byok-prov { display: flex; justify-content: space-between; align-items: center; gap: 12px; font-size: 12.5px; padding: 4px 0; }
	.byok-prov-name { color: var(--code-text); font-weight: 500; }
	.byok-prov-kind { color: var(--code-text-3); font-family: var(--font-mono); font-size: 11px; }

	@media (max-width: 980px) {
		.byok-grid { grid-template-columns: 1fr; }
	}
</style>
