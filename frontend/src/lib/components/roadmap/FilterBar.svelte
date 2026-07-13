<script lang="ts">
	import { TAGS } from '$lib/data/roadmap';
	import type { TagId } from '$lib/data/roadmap';

	export let activeTag: TagId | null = null;
	/**
	 * Callback invoked when a filter chip is clicked.
	 * @param tag - The tag id to filter by, or null to clear the filter (show all).
	 */
	export let onFilter: (tag: TagId | null) => void = () => {};
</script>

<div class="fb wrap">
	<button
		class="fb-chip"
		class:active={activeTag === null}
		on:click={() => onFilter(null)}
	>tout</button>
	{#each Object.entries(TAGS) as [id, tag]}
		<button
			class="fb-chip"
			class:active={activeTag === id}
			style="--th:{tag.hue}"
			on:click={() => onFilter(id as TagId)}
		>{tag.label}</button>
	{/each}
</div>

<style>
	.fb {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 6px;
		padding-block: 20px;
	}
	.fb-chip {
		font-family: var(--font-mono);
		font-size: 11.5px;
		padding: 4px 10px;
		border-radius: 20px;
		border: 1px solid var(--border);
		background: var(--surface);
		color: var(--text-3);
		cursor: pointer;
		transition: border-color .12s ease, color .12s ease, background .12s ease;
	}
	.fb-chip:hover { color: var(--text); }
	.fb-chip.active {
		color: oklch(82% 0.14 var(--th, 152));
		border-color: oklch(82% 0.14 var(--th, 152) / 0.4);
		background: oklch(82% 0.14 var(--th, 152) / 0.08);
	}
	.fb-chip:first-child.active {
		color: var(--accent);
		border-color: color-mix(in oklch, var(--accent) 40%, transparent);
		background: color-mix(in oklch, var(--accent) 8%, transparent);
	}
</style>
