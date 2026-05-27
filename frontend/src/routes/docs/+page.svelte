<svelte:head>
	<title>trippier/api · Docs</title>
</svelte:head>

<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { PAGES } from '$lib/data/docs';
	import { theme } from '$lib/stores/theme';
	import TopoBackground from '$lib/components/TopoBackground.svelte';
	import DocsNav      from '$lib/components/docs/DocsNav.svelte';
	import DocsSidebar  from '$lib/components/docs/DocsSidebar.svelte';
	import GuideContent from '$lib/components/docs/GuideContent.svelte';
	import RouteContent from '$lib/components/docs/RouteContent.svelte';
	import DocsPanel    from '$lib/components/docs/DocsPanel.svelte';

	let activeId = 'quickstart';

	$: currentPage = PAGES.find(p => p.id === activeId) ?? PAGES[0];
	$: topoColor = $theme === 'light' ? '#139450' : '#34d39c';
	$: topoOpacity = $theme === 'light' ? 0.12 : 0.14;
	$: routePage   = currentPage.kind === 'route' ? currentPage : null;
	$: pageIdx     = PAGES.indexOf(currentPage);
	$: prevPage    = PAGES[pageIdx - 1] ?? null;
	$: nextPage    = PAGES[pageIdx + 1] ?? null;

	// Updates URL hash and scrolls main column to top
	function select(id: string) {
		activeId = id;
		if (browser) {
			window.location.hash = id;
			document.querySelector('.d-main')?.scrollTo({ top: 0, behavior: 'smooth' });
		}
	}

	onMount(() => {
		const hash = window.location.hash.slice(1);
		if (hash && PAGES.some(p => p.id === hash)) activeId = hash;

		const handleHash = () => {
			const h = window.location.hash.slice(1);
			if (h && PAGES.some(p => p.id === h)) activeId = h;
		};
		window.addEventListener('hashchange', handleHash);
		return () => window.removeEventListener('hashchange', handleHash);
	});
</script>

<div class="d-page">
	{#if browser}<TopoBackground density={28} opacity={topoOpacity} color={topoColor} />{/if}

	<DocsNav />

	<div class="d-layout">
		<DocsSidebar {activeId} onSelect={select} />

		<main class="d-main">
			<div class="d-main-inner">
				{#if currentPage.kind === 'guide'}
					<GuideContent pageId={activeId} />
				{:else}
					<RouteContent page={currentPage} />
				{/if}

				<div class="d-pager">
					{#if prevPage}
						<button class="d-pager-link prev" on:click={() => select(prevPage.id)}>
							<span>← précédent</span>
							<strong>{prevPage.title}</strong>
						</button>
					{:else}
						<span></span>
					{/if}
					{#if nextPage}
						<button class="d-pager-link next" on:click={() => select(nextPage.id)}>
							<span>suivant →</span>
							<strong>{nextPage.title}</strong>
						</button>
					{:else}
						<span></span>
					{/if}
				</div>
			</div>
		</main>

		<DocsPanel page={routePage} />
	</div>
</div>

<style>
	.d-page {
		position: relative;
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		background: var(--bg);
		color: var(--text);
		font-family: var(--font-sans);
		-webkit-font-smoothing: antialiased;
	}
	.d-layout {
		display: grid;
		grid-template-columns: 280px minmax(0, 1fr) 460px;
		flex: 1;
		position: relative;
		z-index: 1;
	}
	.d-main {
		height: calc(100vh - 56px);
		overflow-y: auto;
		scroll-behavior: smooth;
	}
	.d-main-inner {
		max-width: 720px;
		padding: 56px 56px 80px;
	}
	.d-pager {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 16px;
		margin-top: 64px;
		padding-top: 24px;
		border-top: 1px solid var(--border);
	}
	.d-pager-link {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 16px 18px;
		background: var(--bg-2);
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		cursor: pointer;
		text-align: left;
		transition: border-color .15s ease;
	}
	.d-pager-link:hover { border-color: var(--accent); }
	.d-pager-link.next  { text-align: right; }
	.d-pager-link span  { font-family: var(--font-mono); font-size: 11px; color: var(--text-3); letter-spacing: 0.04em; }
	.d-pager-link strong { font-size: 14px; color: var(--text); font-weight: 500; }

	@media (max-width: 1280px) {
		.d-layout { grid-template-columns: 260px minmax(0, 1fr) 400px; }
		.d-main-inner { padding: 48px 40px 64px; }
	}
	@media (max-width: 1100px) {
		.d-layout { grid-template-columns: 240px minmax(0, 1fr); }
	}
	@media (max-width: 800px) {
		.d-layout { grid-template-columns: 1fr; }
		.d-main-inner { padding: 32px 24px 64px; }
	}
</style>
