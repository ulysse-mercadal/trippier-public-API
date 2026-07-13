<script lang="ts">
	import { PAGES } from '$lib/data/docs';
	import { methodColor } from '$lib/utils/docs';

	export let activeId: string;
	export let onSelect: (id: string) => void;

	let query = '';

	/**
	 * Filters PAGES by the current search query and groups the results by their `group` field.
	 * @returns An array of `[group, items]` tuples, each pairing a group name with its matching pages.
	 */
	$: groups = (() => {
		const q = query.trim().toLowerCase();
		const filtered = q
			? PAGES.filter(p =>
				p.title.toLowerCase().includes(q) ||
				(p.path ?? '').toLowerCase().includes(q) ||
				(p.summary ?? '').toLowerCase().includes(q))
			: PAGES;
		const m = new Map<string, typeof PAGES>();
		for (const p of filtered) {
			if (!m.has(p.group)) m.set(p.group, []);
			m.get(p.group)!.push(p);
		}
		return [...m.entries()];
	})();
</script>

<aside class="d-side">
	<div class="d-search">
		<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="11" cy="11" r="7"/><path d="m20 20-3-3"/></svg>
		<input type="text" placeholder="Rechercher une route…" bind:value={query} />
		<span class="d-kbd">⌘K</span>
	</div>
	<nav class="d-tree">
		{#each groups as [group, items]}
			<div class="d-group">
				<div class="d-group-label">{group}</div>
				<ul>
					{#each items as p}
						<li
							class="d-item"
							class:active={p.id === activeId}
							on:click={() => onSelect(p.id)}
							on:keydown={(e) => e.key === 'Enter' && onSelect(p.id)}
							role="button"
							tabindex="0"
						>
							{#if p.kind === 'route'}
								<span class="d-method" style="color:{methodColor(p.method ?? '')}">{p.method}</span>
								<span class="d-item-path">{p.path}</span>
							{:else}
								<span class="d-item-title">{p.title}</span>
							{/if}
						</li>
					{/each}
				</ul>
			</div>
		{/each}
	</nav>
	<div class="d-side-foot">
		<a href="/" class="d-side-link">
			<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" style="transform:rotate(180deg)"><path d="M5 12h14M13 6l6 6-6 6"/></svg>
			Retour au site
		</a>
	</div>
</aside>

<style>
	.d-side {
		border-right: 1px solid var(--border);
		background: color-mix(in oklch, var(--bg) 92%, transparent);
		height: calc(100vh - 56px);
		position: sticky;
		top: 56px;
		display: flex;
		flex-direction: column;
	}
	.d-search {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 16px 20px;
		border-bottom: 1px solid var(--border);
		color: var(--text-3);
	}
	.d-search input {
		flex: 1;
		background: none;
		border: none;
		outline: none;
		font-family: var(--font-sans);
		font-size: 13px;
		color: var(--text);
	}
	.d-search input::placeholder { color: var(--text-3); }
	.d-kbd {
		font-family: var(--font-mono);
		font-size: 10.5px;
		padding: 2px 6px;
		border: 1px solid var(--border);
		border-radius: 4px;
		color: var(--text-3);
	}
	.d-tree { flex: 1; overflow-y: auto; padding: 16px 0 24px; }
	.d-group { margin-bottom: 20px; }
	.d-group-label {
		font-family: var(--font-mono);
		font-size: 10.5px;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: var(--text-3);
		padding: 0 20px;
		margin-bottom: 8px;
	}
	.d-tree ul { list-style: none; padding: 0; margin: 0; }
	.d-item {
		display: grid;
		grid-template-columns: 48px 1fr;
		align-items: center;
		gap: 8px;
		padding: 7px 20px;
		cursor: pointer;
		font-size: 13px;
		color: var(--text-2);
		border-left: 2px solid transparent;
		transition: color .12s ease, background .12s ease;
	}
	.d-item:hover { color: var(--text); background: color-mix(in oklch, var(--surface) 40%, transparent); }
	.d-item.active {
		color: var(--text);
		background: color-mix(in oklch, var(--accent) 5%, var(--surface));
		border-left-color: var(--accent);
	}
	.d-item-title { grid-column: 1 / -1; font-weight: 500; }
	.d-item-path {
		font-family: var(--font-mono);
		font-size: 12px;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.d-method { font-family: var(--font-mono); font-size: 10.5px; font-weight: 600; letter-spacing: 0.04em; text-align: left; }
	.d-side-foot { border-top: 1px solid var(--border); padding: 14px 20px; }
	.d-side-link {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 12px;
		color: var(--text-3);
		font-family: var(--font-mono);
		text-decoration: none;
	}
	.d-side-link:hover { color: var(--accent); }
</style>
