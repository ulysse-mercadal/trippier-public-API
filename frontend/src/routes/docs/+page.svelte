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
	let sidebarOpen = false;
	let panelOpen = false;

	$: currentPage = PAGES.find(p => p.id === activeId) ?? PAGES[0];
	$: topoColor = $theme === 'light' ? '#139450' : '#34d39c';
	$: topoOpacity = $theme === 'light' ? 0.12 : 0.14;
	$: routePage   = currentPage.kind === 'route' ? currentPage : null;
	$: pageIdx     = PAGES.indexOf(currentPage);
	$: prevPage    = PAGES[pageIdx - 1] ?? null;
	$: nextPage    = PAGES[pageIdx + 1] ?? null;

	/**
	 * Sets active doc page, updates URL hash and scrolls main column to top.
	 * @param id doc page id to activate
	 */
	function select(id: string) {
		activeId = id;
		sidebarOpen = false;
		if (browser) {
			window.location.hash = id;
			document.querySelector('.d-main')?.scrollTo({ top: 0, behavior: 'smooth' });
		}
	}

	/** Closes both the sidebar and panel drawers. */
	function closeDrawers() {
		sidebarOpen = false;
		panelOpen = false;
	}

	/** Toggles the mobile sidebar drawer, closing the panel drawer if opened. */
	function toggleSidebar() {
		sidebarOpen = !sidebarOpen;
		if (sidebarOpen) panelOpen = false;
	}

	/** Toggles the mobile panel drawer, closing the sidebar drawer if opened. */
	function togglePanel() {
		panelOpen = !panelOpen;
		if (panelOpen) sidebarOpen = false;
	}

	onMount(() => {
		const hash = window.location.hash.slice(1);
		if (hash && PAGES.some(p => p.id === hash)) activeId = hash;

		/** Syncs activeId from the current URL hash. */
		const handleHash = () => {
			const h = window.location.hash.slice(1);
			if (h && PAGES.some(p => p.id === h)) activeId = h;
		};
		/**
		 * Closes the drawers when the Escape key is pressed.
		 * @param e keyboard event from the window listener
		 */
		const handleKey = (e: KeyboardEvent) => {
			if (e.key === 'Escape') closeDrawers();
		};
		window.addEventListener('hashchange', handleHash);
		window.addEventListener('keydown', handleKey);
		return () => {
			window.removeEventListener('hashchange', handleHash);
			window.removeEventListener('keydown', handleKey);
		};
	});
</script>

<div class="d-page">
	{#if browser}<TopoBackground density={28} opacity={topoOpacity} color={topoColor} />{/if}

	<DocsNav />

	<div class="d-mobile-bar">
		<button class="d-mb-btn" class:active={sidebarOpen} on:click={toggleSidebar}>
			<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" aria-hidden="true"><path d="M4 6h16M4 12h16M4 18h16"/></svg>
			Routes
		</button>
		{#if routePage}
			<button class="d-mb-btn" class:active={panelOpen} on:click={togglePanel}>
				<svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M8 5v14l11-7z"/></svg>
				Tester
			</button>
		{:else}
			<span></span>
		{/if}
	</div>

	<div class="d-layout">
		<div class="d-side-wrap" class:open={sidebarOpen}>
			<DocsSidebar {activeId} onSelect={select} />
		</div>

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

		<div class="d-panel-wrap" class:open={panelOpen}>
			<DocsPanel page={routePage} />
		</div>
	</div>

	{#if sidebarOpen || panelOpen}
		<button class="d-backdrop" on:click={closeDrawers} aria-label="Fermer"></button>
	{/if}
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
	}
	.d-main {
		height: calc(100vh - 56px);
		overflow-y: auto;
		scroll-behavior: smooth;
	}
	.d-main-inner {
		max-width: none;
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

	.d-mobile-bar { display: none; }

	.d-side-wrap,
	.d-panel-wrap { display: contents; }

	@media (max-width: 1280px) {
		.d-layout { grid-template-columns: 260px minmax(0, 1fr) 400px; }
		.d-main-inner { padding: 48px 40px 64px; }
	}

	@media (max-width: 1100px) {
		.d-layout { grid-template-columns: 1fr; }

		.d-mobile-bar {
			display: flex;
			justify-content: space-between;
			align-items: center;
			gap: 12px;
			padding: 10px 20px;
			border-bottom: 1px solid var(--border);
			background: color-mix(in oklch, var(--bg) 90%, transparent);
			backdrop-filter: blur(10px);
			position: sticky;
			top: 56px;
			z-index: 15;
		}
		.d-mb-btn {
			display: inline-flex;
			align-items: center;
			gap: 8px;
			padding: 7px 14px;
			border: 1px solid var(--border);
			border-radius: var(--r-md);
			background: none;
			color: var(--text-2);
			font-size: 13px;
			font-family: var(--font-sans);
			cursor: pointer;
			transition: color .15s ease, border-color .15s ease, background .15s ease;
		}
		.d-mb-btn:hover,
		.d-mb-btn.active {
			color: var(--accent);
			border-color: var(--accent);
			background: color-mix(in oklch, var(--accent) 8%, transparent);
		}

		.d-side-wrap,
		.d-panel-wrap {
			display: block;
			position: fixed;
			top: 56px;
			bottom: 0;
			width: min(360px, 90vw);
			z-index: 30;
			transition: transform .25s ease;
			background: var(--bg);
			box-shadow: 0 24px 60px -20px rgba(0, 0, 0, 0.35);
		}
		.d-side-wrap {
			left: 0;
			transform: translateX(-100%);
			border-right: 1px solid var(--border);
		}
		.d-side-wrap.open { transform: translateX(0); }
		.d-panel-wrap {
			right: 0;
			width: min(480px, 94vw);
			transform: translateX(100%);
			border-left: 1px solid var(--border);
		}
		.d-panel-wrap.open { transform: translateX(0); }

		.d-side-wrap :global(.d-side),
		.d-panel-wrap :global(.d-panel) {
			position: static;
			height: 100%;
			top: auto;
			border-left: 0;
			border-right: 0;
			background: var(--bg);
		}

		.d-backdrop {
			position: fixed;
			inset: 56px 0 0;
			background: rgba(0, 0, 0, 0.45);
			border: none;
			cursor: pointer;
			z-index: 25;
			padding: 0;
		}
		[data-theme="light"] .d-backdrop { background: rgba(20, 35, 28, 0.35); }
	}

	@media (max-width: 800px) {
		.d-main-inner { padding: 32px 24px 64px; }
		.d-mobile-bar { padding: 10px 16px; }
	}
</style>
