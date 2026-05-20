<script lang="ts">
	import { TAGS } from '$lib/data/roadmap';
	import type { RoadmapItem, TagId } from '$lib/data/roadmap';

	export let item:    RoadmapItem;
	export let editing: boolean = false;
	export let onEdit:  (item: RoadmapItem) => void = () => {};
	export let onDelete: (id: string) => void = () => {};

	$: tag = TAGS[item.tag as TagId];
</script>

<div class="kc" class:kc-editing={editing}>
	<div class="kc-top">
		<span
			class="kc-tag"
			style="--th:{tag?.hue ?? 152}"
		>{tag?.label ?? item.tag}</span>
		{#if item.meta}
			<span class="kc-meta">{item.meta}</span>
		{/if}
		{#if item.votes != null}
			<span class="kc-votes">
				<svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor"><path d="M12 4l8 8H4z"/></svg>
				{item.votes}
			</span>
		{/if}
	</div>
	<p class="kc-title">{item.title}</p>
	<div class="kc-foot">
		<a
			href="https://github.com/ulysse-mercadal/trippier-public-API/issues/{item.issue}"
			target="_blank"
			rel="noopener"
			class="kc-issue"
		>#{item.issue}</a>
		{#if editing}
			<div class="kc-actions">
				<button class="kc-act" on:click={() => onEdit(item)} title="modifier">
					<svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
				</button>
				<button class="kc-act del" on:click={() => onDelete(item.id)} title="supprimer">
					<svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/><path d="M10 11v6M14 11v6"/></svg>
				</button>
			</div>
		{/if}
	</div>
</div>

<style>
	.kc {
		background: var(--bg-2);
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		padding: 14px 16px;
		transition: border-color .15s ease;
	}
	.kc:hover { border-color: color-mix(in oklch, var(--text-3) 50%, transparent); }
	.kc-editing { border-style: dashed; }
	.kc-top {
		display: flex;
		align-items: center;
		gap: 6px;
		margin-bottom: 8px;
	}
	.kc-tag {
		font-family: var(--font-mono);
		font-size: 10px;
		padding: 2px 7px;
		border-radius: 4px;
		color: oklch(82% 0.14 var(--th, 152));
		background: oklch(82% 0.14 var(--th, 152) / 0.1);
		border: 1px solid oklch(82% 0.14 var(--th, 152) / 0.25);
		letter-spacing: 0.04em;
		text-transform: uppercase;
	}
	.kc-meta {
		font-family: var(--font-mono);
		font-size: 10px;
		color: var(--text-3);
		margin-left: auto;
	}
	.kc-votes {
		font-family: var(--font-mono);
		font-size: 10.5px;
		color: var(--text-3);
		display: inline-flex;
		align-items: center;
		gap: 3px;
		margin-left: auto;
	}
	.kc-title {
		font-size: 13.5px;
		color: var(--text);
		line-height: 1.45;
		margin: 0 0 10px;
	}
	.kc-foot {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}
	.kc-issue {
		font-family: var(--font-mono);
		font-size: 10.5px;
		color: var(--text-3);
		text-decoration: none;
		transition: color .12s ease;
	}
	.kc-issue:hover { color: var(--accent); }
	.kc-actions { display: flex; gap: 4px; }
	.kc-act {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 24px; height: 24px;
		border-radius: 4px;
		border: 1px solid var(--border);
		background: var(--surface);
		color: var(--text-3);
		cursor: pointer;
		transition: color .12s ease, border-color .12s ease;
	}
	.kc-act:hover { color: var(--text); border-color: var(--text-3); }
	.kc-act.del:hover { color: oklch(72% 0.16 25); border-color: oklch(72% 0.16 25 / 0.5); }
</style>
