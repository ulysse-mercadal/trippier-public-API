<svelte:head>
	<title>trippier/api · Roadmap</title>
</svelte:head>

<script lang="ts">
	import { browser } from '$app/environment';
	import { env as pubEnv } from '$env/dynamic/public';
	import { auth } from '$lib/stores/auth';
	import { t } from '$lib/i18n';
	import Footer       from '$lib/components/landing/Footer.svelte';
	import FilterBar    from '$lib/components/roadmap/FilterBar.svelte';
	import KanbanColumn from '$lib/components/roadmap/KanbanColumn.svelte';
	import ItemModal    from '$lib/components/roadmap/ItemModal.svelte';
	import type { PageData } from './$types';
	import type { RoadmapData, RoadmapItem, ColumnId, TagId } from '$lib/data/roadmap';

	export let data: PageData;

	let roadmap: RoadmapData = JSON.parse(JSON.stringify(data.roadmap));
	let activeTag: TagId | null = null;
	let editing  = false;
	let saving   = false;
	let modal: { item?: Partial<RoadmapItem>; colId: ColumnId } | null = null;

	$: adminEmails = (pubEnv.PUBLIC_ADMIN_EMAILS ?? '').split(',').map(e => e.trim()).filter(Boolean);
	$: isAdmin     = !!$auth.user && adminEmails.includes($auth.user.email);

	$: shippedCount  = roadmap.columns.find(c => c.id === 'shipped')?.items.length ?? 0;
	$: progressCount = roadmap.columns.find(c => c.id === 'progress')?.items.length ?? 0;
	$: nextCount     = roadmap.columns.find(c => c.id === 'next')?.items.length ?? 0;
	$: laterCount    = roadmap.columns.find(c => c.id === 'later')?.items.length ?? 0;

	/**
	 * Enters edit mode with a fresh snapshot of the roadmap.
	 */
	function startEdit() {
		roadmap = JSON.parse(JSON.stringify(data.roadmap));
		editing = true;
	}

	/**
	 * Discards pending changes and exits edit mode.
	 */
	function cancelEdit() {
		roadmap = JSON.parse(JSON.stringify(data.roadmap));
		editing = false;
	}

	/**
	 * Persists the current roadmap state to the server.
	 */
	async function saveEdit() {
		saving = true;
		const token = browser ? auth.getStoredToken() : null;
		try {
			const res = await fetch('/api/roadmap', {
				method:  'PUT',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body:    JSON.stringify(roadmap),
			});
			if (!res.ok) throw new Error(await res.text());
			data.roadmap = JSON.parse(JSON.stringify(roadmap));
			editing = false;
		} catch (e) {
			alert(`Erreur: ${e instanceof Error ? e.message : e}`);
		} finally {
			saving = false;
		}
	}

	/**
	 * Opens the item modal for adding a new item to a column.
	 * @param colId target column
	 */
	function openAdd(colId: ColumnId) {
		modal = { colId };
	}

	/**
	 * Opens the item modal pre-filled to edit an existing item.
	 * @param item roadmap item to edit
	 */
	function openEdit(item: RoadmapItem) {
		const col = roadmap.columns.find(c => c.items.some(i => i.id === item.id));
		modal = { item, colId: col?.id as ColumnId ?? 'next' };
	}

	/**
	 * Removes an item from the given column.
	 * @param colId column containing the item
	 * @param itemId item to delete
	 */
	function deleteItem(colId: string, itemId: string) {
		roadmap = {
			...roadmap,
			columns: roadmap.columns.map(c =>
				c.id === colId
					? { ...c, items: c.items.filter(i => i.id !== itemId) }
					: c
			),
		};
	}

	/**
	 * Handles save from modal: moves item to target column, removing from old one.
	 * @param targetColId column the item should end up in
	 * @param item item being saved
	 */
	function saveModal(targetColId: ColumnId, item: RoadmapItem) {
		roadmap = {
			...roadmap,
			columns: roadmap.columns.map(c => {
				const filtered = { ...c, items: c.items.filter(i => i.id !== item.id) };
				if (c.id !== targetColId) return filtered;
				const exists = c.items.some(i => i.id === item.id);
				return { ...filtered, items: exists
					? c.items.map(i => i.id === item.id ? item : i)
					: [...filtered.items, item]
				};
			}),
		};
		modal = null;
	}
</script>

<div class="rp-hero wrap">
	<div class="rp-hero-text">
		<p class="eyebrow">{$t('rm_eyebrow')}</p>
		<h1>{$t('rm_title')} <span class="accent-fg">{$t('rm_accent')}</span></h1>
		<p class="rp-sub">{$t('rm_sub')}</p>
		<div class="rp-stats">
			<div class="rp-stat"><strong>{shippedCount}</strong><span>{$t('rm_shipped')}</span></div>
			<div class="rp-stat-sep"></div>
			<div class="rp-stat"><strong>{progressCount + nextCount}</strong><span>{$t('rm_in_progress')}</span></div>
			<div class="rp-stat-sep"></div>
			<div class="rp-stat"><strong>{laterCount}</strong><span>{$t('rm_exploring')}</span></div>
		</div>
	</div>
	{#if isAdmin}
		<div class="rp-admin-bar">
			{#if editing}
				<button class="rp-btn cancel" on:click={cancelEdit} disabled={saving}>{$t('rm_cancel')}</button>
				<button class="rp-btn save"   on:click={saveEdit}   disabled={saving}>
					{saving ? '…' : $t('rm_save')}
				</button>
			{:else}
				<button class="rp-btn edit" on:click={startEdit}>
					<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
					{$t('rm_edit')}
				</button>
			{/if}
		</div>
	{/if}
</div>

<FilterBar {activeTag} onFilter={(t) => (activeTag = t)} />

<div class="rp-board wrap">
	{#each roadmap.columns as col (col.id)}
		<KanbanColumn
			column={col}
			{activeTag}
			{editing}
			onEditItem={openEdit}
			onDeleteItem={deleteItem}
			onAddItem={openAdd}
		/>
	{/each}
</div>

<div class="rp-contrib wrap">
	<h2>{$t('rm_contrib_title')}</h2>
	<div class="rp-contrib-grid">
		<a href="https://github.com/ulysse-mercadal/trippier-public-API/issues" target="_blank" rel="noopener" class="rp-ccard">
			<svg class="rp-ccard-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
			<strong>{$t('rm_report_title')}</strong>
			<p>{$t('rm_report_desc')}</p>
		</a>
		<a href="https://github.com/ulysse-mercadal/trippier-public-API/discussions" target="_blank" rel="noopener" class="rp-ccard">
			<svg class="rp-ccard-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
			<strong>{$t('rm_suggest_title')}</strong>
			<p>{$t('rm_suggest_desc')}</p>
		</a>
		<a href="https://github.com/ulysse-mercadal/trippier-public-API/pulls" target="_blank" rel="noopener" class="rp-ccard">
			<svg class="rp-ccard-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="18" cy="18" r="3"/><circle cx="6" cy="6" r="3"/><path d="M13 6h3a2 2 0 0 1 2 2v7"/><line x1="6" y1="9" x2="6" y2="21"/></svg>
			<strong>{$t('rm_pr_title')}</strong>
			<p>{$t('rm_pr_desc')}</p>
		</a>
		<a href="https://github.com/ulysse-mercadal/trippier-public-API/stargazers" target="_blank" rel="noopener" class="rp-ccard">
			<svg class="rp-ccard-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
			<strong>{$t('rm_star_title')}</strong>
			<p>{$t('rm_star_desc')}</p>
		</a>
	</div>
</div>

<Footer />

{#if modal}
	<ItemModal
		item={modal.item ?? {}}
		colId={modal.colId}
		onSave={saveModal}
		onClose={() => (modal = null)}
	/>
{/if}

<style>
	.rp-hero {
		padding-block: 64px 32px;
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 24px;
		flex-wrap: wrap;
	}
	.rp-hero h1 {
		font-size: clamp(32px, 4vw, 52px);
		font-weight: 600;
		letter-spacing: -0.025em;
		line-height: 1.08;
		margin: 8px 0 14px;
	}
	.rp-sub {
		font-size: 17px;
		color: var(--text-2);
		line-height: 1.55;
		max-width: 560px;
		margin: 0 0 28px;
	}
	.rp-stats {
		display: flex;
		align-items: center;
		gap: 20px;
	}
	.rp-stat { display: flex; flex-direction: column; gap: 2px; }
	.rp-stat strong { font-size: 22px; font-weight: 700; color: var(--text); line-height: 1; }
	.rp-stat span   { font-family: var(--font-mono); font-size: 11px; color: var(--text-3); letter-spacing: 0.04em; }
	.rp-stat-sep { width: 1px; height: 28px; background: var(--border); }
	.rp-admin-bar { display: flex; gap: 8px; align-items: center; padding-bottom: 4px; }
	.rp-btn {
		font-family: var(--font-mono);
		font-size: 12px;
		padding: 7px 14px;
		border-radius: 6px;
		border: 1px solid var(--border);
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		gap: 6px;
		transition: border-color .12s ease, color .12s ease;
	}
	.rp-btn.edit   { background: var(--surface); color: var(--text-2); }
	.rp-btn.edit:hover { border-color: var(--accent); color: var(--accent); }
	.rp-btn.cancel { background: none; color: var(--text-3); }
	.rp-btn.cancel:hover { color: var(--text); }
	.rp-btn.save   { background: var(--accent); color: #08120e; border-color: var(--accent); font-weight: 600; }
	.rp-btn.save:hover { filter: brightness(1.08); }
	.rp-btn:disabled { opacity: 0.5; cursor: not-allowed; filter: none; }
	.rp-board {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 16px;
		padding-bottom: 64px;
	}
	.rp-contrib { padding-bottom: 80px; }
	.rp-contrib h2 {
		font-size: 22px;
		font-weight: 600;
		letter-spacing: -0.015em;
		margin: 0 0 20px;
		padding-top: 32px;
		border-top: 1px solid var(--border);
	}
	.rp-contrib-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 12px;
	}
	.rp-ccard {
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding: 18px 20px;
		background: var(--bg-2);
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		text-decoration: none;
		transition: border-color .15s ease;
	}
	.rp-ccard:hover { border-color: var(--accent); }
	.rp-ccard-icon { color: var(--accent); margin-bottom: 4px; flex-shrink: 0; }
	.rp-ccard strong { font-size: 13.5px; font-weight: 600; color: var(--text); }
	.rp-ccard p { font-size: 12.5px; color: var(--text-3); line-height: 1.5; margin: 0; }

	@media (max-width: 1100px) {
		.rp-board         { grid-template-columns: repeat(2, 1fr); }
		.rp-contrib-grid  { grid-template-columns: repeat(2, 1fr); }
	}
	@media (max-width: 640px) {
		.rp-board         { grid-template-columns: 1fr; }
		.rp-contrib-grid  { grid-template-columns: 1fr; }
	}
</style>
