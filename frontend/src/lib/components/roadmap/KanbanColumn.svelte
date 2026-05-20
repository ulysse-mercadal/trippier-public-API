<script lang="ts">
	import type { RoadmapColumn, RoadmapItem, TagId } from '$lib/data/roadmap';
	import KanbanCard from './KanbanCard.svelte';

	export let column:    RoadmapColumn;
	export let activeTag: TagId | null = null;
	export let editing:   boolean      = false;
	export let onEditItem:   (item: RoadmapItem) => void = () => {};
	export let onDeleteItem: (colId: string, itemId: string) => void = () => {};
	export let onAddItem:    (colId: ColumnId) => void = () => {};

	$: visible = activeTag
		? column.items.filter(i => i.tag === activeTag)
		: column.items;
</script>

<div class="kk" class:kk-progress={column.id === 'progress'}>
	<div class="kk-head">
		<div class="kk-icon">
			{#if column.id === 'shipped'}
				<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
			{:else if column.id === 'progress'}
				<svg class="spin" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg>
			{:else if column.id === 'next'}
				<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
			{:else}
				<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M9 18h6M10 22h4M12 2a7 7 0 0 0-7 7c0 4 2 6 3 7h8c1-1 3-3 3-7a7 7 0 0 0-7-7z"/></svg>
			{/if}
		</div>
		<div>
			<h3 class="kk-label">{column.label}</h3>
			<p class="kk-sub">{column.sub}</p>
		</div>
		<span class="kk-count">{visible.length}</span>
	</div>

	<div class="kk-list">
		{#each visible as item (item.id)}
			<KanbanCard
				{item}
				{editing}
				onEdit={onEditItem}
				onDelete={(id) => onDeleteItem(column.id, id)}
			/>
		{/each}
		{#if editing}
			<button class="kk-add" on:click={() => onAddItem(column.id)}>
				<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
				ajouter
			</button>
		{/if}
	</div>
</div>

<style>
	.kk {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--r-lg);
		padding: 0;
		display: flex;
		flex-direction: column;
		min-width: 0;
	}
	.kk-progress { border-color: color-mix(in oklch, var(--accent) 30%, var(--border)); }
	.kk-head {
		display: flex;
		align-items: flex-start;
		gap: 10px;
		padding: 18px 18px 14px;
		border-bottom: 1px solid var(--border);
	}
	.kk-icon {
		width: 28px; height: 28px;
		border-radius: 7px;
		border: 1px solid var(--border);
		background: var(--bg-2);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--accent);
		flex-shrink: 0;
		margin-top: 1px;
	}
	.kk-progress .kk-icon { background: color-mix(in oklch, var(--accent) 8%, transparent); border-color: color-mix(in oklch, var(--accent) 25%, transparent); }
	.kk-label { font-size: 13.5px; font-weight: 600; color: var(--text); margin: 0; }
	.kk-sub   { font-size: 11px; color: var(--text-3); margin: 2px 0 0; font-family: var(--font-mono); }
	.kk-count {
		margin-left: auto;
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--text-3);
		background: var(--bg-2);
		border: 1px solid var(--border);
		border-radius: 4px;
		padding: 2px 7px;
		flex-shrink: 0;
	}
	.kk-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 14px;
		flex: 1;
	}
	.kk-add {
		display: flex;
		align-items: center;
		gap: 6px;
		width: 100%;
		padding: 10px 12px;
		border: 1px dashed var(--border);
		border-radius: var(--r-md);
		background: none;
		color: var(--text-3);
		font-family: var(--font-mono);
		font-size: 11.5px;
		cursor: pointer;
		transition: color .12s ease, border-color .12s ease;
		margin-top: 4px;
	}
	.kk-add:hover { color: var(--accent); border-color: color-mix(in oklch, var(--accent) 40%, transparent); }
	.spin { animation: spin 2s linear infinite; }
	@keyframes spin { to { transform: rotate(360deg); } }
</style>
