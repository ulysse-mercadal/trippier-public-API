<script lang="ts">
	import { TAGS } from '$lib/data/roadmap';
	import type { RoadmapItem, TagId, ColumnId } from '$lib/data/roadmap';

	export let item:    Partial<RoadmapItem> = {};
	export let colId:   ColumnId = 'next';
	export let onSave:  (colId: ColumnId, item: RoadmapItem) => void = () => {};
	export let onClose: () => void = () => {};

	const isNew = !item.id;

	let title  = item.title  ?? '';
	let tag    = (item.tag   ?? 'routes') as TagId;
	let meta   = item.meta   ?? '';
	let issue  = item.issue  ?? 0;
	let votes  = item.votes  ?? '';
	let column = colId;

	// Generates a short random ID for new items
	function genId() { return Math.random().toString(36).slice(2, 8); }

	function submit() {
		if (!title.trim() || !issue) return;
		const result: RoadmapItem = {
			id:    item.id ?? genId(),
			title: title.trim(),
			tag,
			issue: Number(issue),
			...(meta  ? { meta }               : {}),
			...(votes !== '' ? { votes: Number(votes) } : {}),
		};
		onSave(column, result);
	}

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') onClose();
	}
</script>

<svelte:window on:keydown={onKeydown} />

<div class="im-backdrop" on:click={onClose} role="presentation">
	<div class="im" on:click|stopPropagation role="dialog" aria-modal="true">
		<div class="im-head">
			<h3>{isNew ? 'Nouvel élément' : 'Modifier'}</h3>
			<button class="im-close" on:click={onClose}>
				<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
			</button>
		</div>

		<div class="im-body">
			<label class="im-field">
				<span>Titre <span class="im-req">*</span></span>
				<input bind:value={title} placeholder="ex: GraphQL API" />
			</label>

			<div class="im-row">
				<label class="im-field">
					<span>Colonne</span>
					<select bind:value={column}>
						<option value="shipped">Shipped</option>
						<option value="progress">In progress</option>
						<option value="next">Next up</option>
						<option value="later">Later / Ideas</option>
					</select>
				</label>
				<label class="im-field">
					<span>Tag</span>
					<select bind:value={tag}>
						{#each Object.entries(TAGS) as [id, t]}
							<option value={id}>{t.label}</option>
						{/each}
					</select>
				</label>
			</div>

			<div class="im-row">
				<label class="im-field">
					<span>Issue GitHub <span class="im-req">*</span></span>
					<input type="number" bind:value={issue} placeholder="42" min="1" />
				</label>
				<label class="im-field">
					<span>Meta (version, date…)</span>
					<input bind:value={meta} placeholder="v0.5" />
				</label>
			</div>

			{#if column === 'later'}
				<label class="im-field">
					<span>Votes</span>
					<input type="number" bind:value={votes} placeholder="0" min="0" />
				</label>
			{/if}
		</div>

		<div class="im-foot">
			<button class="im-cancel" on:click={onClose}>annuler</button>
			<button class="im-submit" on:click={submit} disabled={!title.trim() || !issue}>
				{isNew ? 'ajouter' : 'sauvegarder'}
			</button>
		</div>
	</div>
</div>

<style>
	.im-backdrop {
		position: fixed;
		inset: 0;
		background: oklch(0% 0 0 / 0.6);
		z-index: 200;
		display: flex;
		align-items: center;
		justify-content: center;
		backdrop-filter: blur(4px);
	}
	.im {
		background: var(--bg);
		border: 1px solid var(--border);
		border-radius: var(--r-lg);
		width: min(520px, calc(100vw - 32px));
		display: flex;
		flex-direction: column;
	}
	.im-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 18px 20px 14px;
		border-bottom: 1px solid var(--border);
	}
	.im-head h3 { font-size: 15px; font-weight: 600; margin: 0; }
	.im-close {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 28px; height: 28px;
		border-radius: 6px;
		border: 1px solid var(--border);
		background: none;
		color: var(--text-3);
		cursor: pointer;
	}
	.im-close:hover { color: var(--text); background: var(--surface); }
	.im-body { display: flex; flex-direction: column; gap: 14px; padding: 20px; }
	.im-row { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
	.im-field { display: flex; flex-direction: column; gap: 5px; }
	.im-field span { font-family: var(--font-mono); font-size: 11px; color: var(--text-3); text-transform: uppercase; letter-spacing: 0.05em; }
	.im-field input,
	.im-field select {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: 6px;
		padding: 8px 10px;
		font-family: var(--font-sans);
		font-size: 13.5px;
		color: var(--text);
		outline: none;
		transition: border-color .12s ease;
	}
	.im-field input:focus,
	.im-field select:focus { border-color: var(--accent); }
	.im-req { color: oklch(72% 0.16 25); }
	.im-foot {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: 8px;
		padding: 14px 20px;
		border-top: 1px solid var(--border);
	}
	.im-cancel {
		font-family: var(--font-mono);
		font-size: 12.5px;
		padding: 7px 14px;
		border-radius: 6px;
		border: 1px solid var(--border);
		background: none;
		color: var(--text-3);
		cursor: pointer;
	}
	.im-cancel:hover { color: var(--text); }
	.im-submit {
		font-family: var(--font-mono);
		font-size: 12.5px;
		font-weight: 600;
		padding: 7px 16px;
		border-radius: 6px;
		border: none;
		background: var(--accent);
		color: #08120e;
		cursor: pointer;
		transition: filter .12s ease;
	}
	.im-submit:hover { filter: brightness(1.08); }
	.im-submit:disabled { opacity: 0.4; cursor: not-allowed; filter: none; }
</style>
