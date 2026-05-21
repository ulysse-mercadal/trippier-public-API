<script lang="ts">
	import type { DocPage } from '$lib/data/docs';
	import { methodColor } from '$lib/utils/docs';
	import CodeBlock from './CodeBlock.svelte';

	export let page: DocPage;
</script>

<div class="d-route-head">
	<div class="d-route-title">
		<span class="d-method" style="color:{methodColor(page.method ?? '')}">{page.method}</span>
		<code class="d-route-path">{page.path}</code>
		<span class="d-cost">{page.cost === 0 ? 'gratuit' : `${page.cost} tk`}</span>
	</div>
	<h1>{page.title}</h1>
	<p class="d-lead">{page.summary}</p>
</div>

{#if page.params}
	<h2>Paramètres</h2>
	<div class="d-params">
		{#each page.params as p}
			<div class="d-param">
				<div class="d-param-name">
					<code>{p.name}</code>
					{#if p.required}<span class="d-tag-req">requis</span>{/if}
					{#if p.in === 'path'}<span class="d-tag-in">path</span>{/if}
				</div>
				<div class="d-param-meta">
					<span class="d-param-type">{p.type}</span>
				</div>
				<div class="d-param-desc">{p.desc}</div>
			</div>
		{/each}
	</div>
{/if}

{#if page.body}
	<h2>Corps de requête</h2>
	<CodeBlock lang="json" code={page.body} />
{/if}

{#if page.response}
	<h2>Réponse · 200</h2>
	<CodeBlock lang="json" code={page.response} />
{/if}

<h2>Codes de réponse</h2>
<table class="d-table">
	<thead><tr><th>HTTP</th><th>Quand</th></tr></thead>
	<tbody>
		<tr><td><span class="d-status ok">200</span></td><td>Requête réussie. Le corps contient la ressource.</td></tr>
		<tr><td><span class="d-status warn">400</span></td><td>Paramètres invalides, vérifier les types.</td></tr>
		<tr><td><span class="d-status warn">401</span></td><td>Token manquant ou invalide.</td></tr>
		{#if (page.cost ?? 0) > 0}
			<tr><td><span class="d-status warn">402</span></td><td>Solde de tokens insuffisant.</td></tr>
		{/if}
		{#if (page.path ?? '').includes('{id}')}
			<tr><td><span class="d-status warn">404</span></td><td>Ressource introuvable.</td></tr>
		{/if}
		<tr><td><span class="d-status err">429</span></td><td>Plus de 60 req/min.</td></tr>
	</tbody>
</table>

<style>
	h1 {
		font-size: 38px;
		font-weight: 600;
		letter-spacing: -0.025em;
		line-height: 1.1;
		margin: 0 0 16px;
	}
	h2 {
		font-size: 20px;
		font-weight: 600;
		letter-spacing: -0.015em;
		margin: 48px 0 14px;
		padding-top: 16px;
		border-top: 1px solid var(--border);
	}
	p { font-size: 15px; line-height: 1.65; color: var(--text-2); margin: 0 0 14px; }
	.d-lead { font-size: 17px !important; color: var(--text-2); line-height: 1.6; margin-bottom: 24px !important; }
	.d-route-head { margin-bottom: 32px; }
	.d-route-title {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 18px;
		padding: 10px 14px;
		background: var(--bg-2);
		border: 1px solid var(--border);
		border-radius: var(--r-md);
	}
	.d-method { font-family: var(--font-mono); font-size: 12px; font-weight: 600; letter-spacing: 0.04em; }
	.d-route-path { font-family: var(--font-mono); font-size: 15px; color: var(--text); flex: 1; }
	.d-cost {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--text-3);
		padding: 3px 8px;
		border-radius: 4px;
		background: var(--bg);
		border: 1px solid var(--border);
	}
	.d-params { display: flex; flex-direction: column; border-top: 1px solid var(--border); }
	.d-param {
		display: grid;
		grid-template-columns: 200px 80px 1fr;
		gap: 16px;
		padding: 16px 0;
		border-bottom: 1px solid var(--border);
		align-items: start;
	}
	.d-param-name { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
	.d-param-name code {
		font-family: var(--font-mono);
		font-size: 13px;
		color: var(--text);
		background: none !important;
		border: none !important;
		padding: 0 !important;
	}
	.d-tag-req, .d-tag-in {
		font-family: var(--font-mono);
		font-size: 9.5px;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		padding: 2px 6px;
		border-radius: 4px;
	}
	.d-tag-req { color: oklch(78% 0.14 30); background: color-mix(in oklch, oklch(78% 0.14 30) 12%, transparent); }
	.d-tag-in  { color: var(--text-3); background: var(--surface); }
	.d-param-type { font-family: var(--font-mono); font-size: 11.5px; color: var(--text-3); }
	.d-param-desc { font-size: 13.5px; color: var(--text-2); line-height: 1.55; }
	.d-table { width: 100%; border-collapse: collapse; margin: 0 0 14px; }
	.d-table th,
	.d-table td {
		text-align: left;
		padding: 10px 12px;
		border-bottom: 1px solid var(--border);
		font-size: 13.5px;
		color: var(--text-2);
		vertical-align: top;
	}
	.d-table th {
		font-family: var(--font-mono);
		font-size: 11px;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--text-3);
		font-weight: 500;
		background: color-mix(in oklch, var(--surface) 40%, transparent);
	}
	.d-table th:first-child,
	.d-table td:first-child { padding-left: 16px; }
	.d-status {
		display: inline-block;
		font-family: var(--font-mono);
		font-size: 11px;
		font-weight: 600;
		padding: 2px 7px;
		border-radius: 4px;
	}
	.d-status.ok   { color: var(--accent); background: color-mix(in oklch, var(--accent) 12%, transparent); }
	.d-status.warn { color: oklch(82% 0.12 80); background: color-mix(in oklch, oklch(82% 0.12 80) 12%, transparent); }
	.d-status.err  { color: oklch(72% 0.16 25); background: color-mix(in oklch, oklch(72% 0.16 25) 12%, transparent); }
</style>
