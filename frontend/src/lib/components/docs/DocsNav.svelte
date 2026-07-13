<script lang="ts">
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { t, locale, setLocale } from '$lib/i18n';
	import { page } from '$app/stores';
	import { activeNavHash } from '$lib/stores/nav';
	import { theme, toggleTheme } from '$lib/stores/theme';
	import { onMount } from 'svelte';

	export let active: string = '';

	let menuOpen = false;

	/**
	 * Locks or unlocks page scrolling by toggling body overflow.
	 * @param lock whether to lock (true) or unlock (false) scrolling
	 */
	function lockScroll(lock: boolean) {
		if (typeof document === 'undefined') return;
		document.body.style.overflow = lock ? 'hidden' : '';
	}

	/**
	 * Logs the user out, closes the mobile menu, and redirects to home.
	 */
	function handleLogout() {
		auth.logout();
		menuOpen = false;
		lockScroll(false);
		goto('/');
	}

	/**
	 * Toggles the mobile menu open state and locks/unlocks scroll to match.
	 */
	function toggleMenu() {
		menuOpen = !menuOpen;
		lockScroll(menuOpen);
	}
	/**
	 * Closes the mobile menu and unlocks scroll.
	 */
	function closeMenu() {
		menuOpen = false;
		lockScroll(false);
	}

	/**
	 * Closes the menu then navigates to the given destination.
	 * @param href target URL or path to navigate to
	 */
	function navigate(href: string) {
		closeMenu();
		if (href.startsWith('http')) {
			window.open(href, '_blank', 'noopener');
		} else {
			goto(href);
		}
	}

	/**
	 * Switches the active locale between French and English.
	 */
	function switchLocale() {
		setLocale($locale === 'fr' ? 'en' : 'fr');
	}

	$: currentActive = active || $activeNavHash || $page.url.pathname;

	onMount(() => {
		const onKey = (e: KeyboardEvent) => {
			if (e.key === 'Escape') closeMenu();
		};
		window.addEventListener('keydown', onKey);
		return () => {
			window.removeEventListener('keydown', onKey);
			lockScroll(false);
		};
	});
</script>

<nav class="d-nav">
	<a class="d-brand" href="/" on:click={closeMenu}>
		<img class="d-brand-logo" src="/favicon.png" alt="trippier logo" width="28" height="28" />
		<span>trippier<span class="brand-dot">/</span>api</span>
	</a>

	<div class="d-nav-links">
		<a href="/#features" class:d-nav-active={currentActive === '/#features'}>{$t('nav_features')}</a>
		<a href="/#deploy"   class:d-nav-active={currentActive === '/#deploy'}>{$t('nav_deploy')}</a>
		<a href="/docs"      class:d-nav-active={currentActive.startsWith('/docs')}>{$t('nav_docs')}</a>
		<a href="/roadmap"   class:d-nav-active={currentActive.startsWith('/roadmap')}>{$t('nav_roadmap')}</a>
		<a href="https://github.com/ulysse-mercadal/trippier-public-API" target="_blank" rel="noopener" class="d-nav-gh">
			<svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M12 2a10 10 0 0 0-3.16 19.49c.5.09.68-.22.68-.48v-1.7c-2.78.6-3.37-1.34-3.37-1.34-.45-1.16-1.11-1.47-1.11-1.47-.91-.62.07-.61.07-.61 1 .07 1.53 1.03 1.53 1.03.89 1.53 2.34 1.09 2.91.83.09-.65.35-1.09.63-1.34-2.22-.25-4.55-1.11-4.55-4.94 0-1.09.39-1.99 1.03-2.69-.1-.25-.45-1.27.1-2.65 0 0 .84-.27 2.75 1.02a9.6 9.6 0 0 1 5 0c1.91-1.29 2.75-1.02 2.75-1.02.55 1.38.2 2.4.1 2.65.64.7 1.03 1.6 1.03 2.69 0 3.84-2.34 4.68-4.57 4.93.36.31.68.92.68 1.85v2.74c0 .27.18.58.69.48A10 10 0 0 0 12 2z"/></svg>
			GitHub
		</a>
	</div>

	<div class="d-nav-cta">
		<button class="theme-btn" on:click={toggleTheme} aria-label="Toggle theme" title="Toggle theme">
			{#if $theme === 'dark'}
				<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
					<circle cx="12" cy="12" r="4" />
					<path d="M12 2v2M12 20v2M4 12H2M22 12h-2M5.6 5.6l1.4 1.4M17 17l1.4 1.4M5.6 18.4l1.4-1.4M17 7l1.4-1.4" />
				</svg>
			{:else}
				<svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
					<path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z" />
				</svg>
			{/if}
		</button>
		<button class="locale-btn" on:click={switchLocale}>
			{$locale === 'fr' ? 'EN' : 'FR'}
		</button>
		{#if $auth.user}
			<a href="/dashboard" class="btn-ghost">{$t('nav_dashboard')}</a>
			<button class="btn-ghost" on:click={handleLogout}>{$t('nav_logout')}</button>
		{:else}
			<a href="/login" class="btn-ghost">{$t('nav_login')}</a>
		{/if}
	</div>

	<button class="d-burger" class:open={menuOpen} on:click={toggleMenu} aria-label="Menu" aria-expanded={menuOpen}>
		<span></span>
		<span></span>
		<span></span>
	</button>
</nav>

{#if menuOpen}
	<div class="d-menu-overlay" role="dialog" aria-modal="true">
		<div class="d-menu-inner">
			<nav class="d-menu-links">
				<button class="d-menu-link" class:active={currentActive === '/#features'} on:click={() => navigate('/#features')}>
					{$t('nav_features')}
				</button>
				<button class="d-menu-link" class:active={currentActive === '/#deploy'} on:click={() => navigate('/#deploy')}>
					{$t('nav_deploy')}
				</button>
				<button class="d-menu-link" class:active={currentActive.startsWith('/docs')} on:click={() => navigate('/docs')}>
					{$t('nav_docs')}
				</button>
				<button class="d-menu-link" class:active={currentActive.startsWith('/roadmap')} on:click={() => navigate('/roadmap')}>
					{$t('nav_roadmap')}
				</button>
				<button class="d-menu-link" on:click={() => navigate('https://github.com/ulysse-mercadal/trippier-public-API')}>
					GitHub
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M14 4h6v6M10 14L20 4M19 14v5a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1h5"/></svg>
				</button>
			</nav>

			<div class="d-menu-actions">
				<button class="d-menu-action" on:click={toggleTheme}>
					{#if $theme === 'dark'}
						<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
							<circle cx="12" cy="12" r="4" />
							<path d="M12 2v2M12 20v2M4 12H2M22 12h-2M5.6 5.6l1.4 1.4M17 17l1.4 1.4M5.6 18.4l1.4-1.4M17 7l1.4-1.4" />
						</svg>
						Thème clair
					{:else}
						<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
							<path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z" />
						</svg>
						Thème sombre
					{/if}
				</button>
				<button class="d-menu-action" on:click={switchLocale}>
					<span class="d-menu-locale">{$locale.toUpperCase()}</span>
					{$locale === 'fr' ? 'English' : 'Français'}
				</button>

				{#if $auth.user}
					<button class="d-menu-cta" on:click={() => navigate('/dashboard')}>{$t('nav_dashboard')}</button>
					<button class="d-menu-cta ghost" on:click={handleLogout}>{$t('nav_logout')}</button>
				{:else}
					<button class="d-menu-cta" on:click={() => navigate('/login')}>{$t('nav_login')}</button>
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	.d-nav {
		display: flex;
		align-items: center;
		gap: 32px;
		padding: 14px 32px;
		border-bottom: 1px solid var(--border);
		background: color-mix(in oklch, var(--bg) 80%, transparent);
		backdrop-filter: blur(12px);
		position: sticky;
		top: 0;
		z-index: 20;
	}
	.d-brand {
		display: inline-flex;
		align-items: center;
		gap: 10px;
		font-weight: 600;
		font-size: 15px;
		color: var(--text);
		letter-spacing: -0.01em;
		text-decoration: none;
		flex-shrink: 0;
	}
	.d-brand-logo {
		width: 28px;
		height: 28px;
		border-radius: 4px;
		object-fit: contain;
	}
	.brand-dot { color: var(--text-3); margin: 0 1px; }
	.d-nav-section {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--text-3);
		padding: 3px 8px;
		border: 1px solid var(--border);
		border-radius: 4px;
		margin-left: 6px;
		letter-spacing: 0.04em;
		text-transform: uppercase;
	}
	.d-nav-links {
		display: flex;
		gap: 24px;
		flex: 1;
		margin-left: 20px;
	}
	.d-nav-links a {
		font-size: 13.5px;
		color: var(--text-2);
		transition: color .12s ease;
		padding: 4px 0;
		position: relative;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: 6px;
	}
	.d-nav-links a:hover { color: var(--text); }
	.d-nav-active { color: var(--text) !important; }
	.d-nav-active::after {
		content: '';
		position: absolute;
		inset: auto 0 -15px;
		height: 2px;
		background: var(--accent);
	}
	.d-nav-cta { display: flex; gap: 10px; align-items: center; }
	.locale-btn {
		font-family: var(--font-mono);
		font-size: 11px;
		font-weight: 600;
		letter-spacing: 0.06em;
		padding: 5px 10px;
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		background: none;
		color: var(--text-3);
		cursor: pointer;
		transition: color .12s ease, border-color .12s ease;
	}
	.locale-btn:hover { color: var(--accent); border-color: var(--accent); }
	.theme-btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		background: none;
		color: var(--text-3);
		cursor: pointer;
		transition: color .12s ease, border-color .12s ease;
	}
	.theme-btn:hover { color: var(--accent); border-color: var(--accent); }
	.btn-ghost {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 7px 14px;
		border-radius: var(--r-md);
		border: 1px solid var(--border);
		color: var(--text-2);
		font-size: 13px;
		font-family: var(--font-sans);
		background: none;
		cursor: pointer;
		transition: border-color .15s ease, color .15s ease;
		text-decoration: none;
	}
	.btn-ghost:hover { border-color: var(--border-2, var(--border)); color: var(--text); }

	.d-burger {
		display: none;
		flex-direction: column;
		justify-content: center;
		gap: 5px;
		width: 36px;
		height: 36px;
		padding: 0 8px;
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		background: none;
		cursor: pointer;
		transition: border-color .12s ease;
		margin-left: auto;
	}
	.d-burger:hover { border-color: var(--accent); }
	.d-burger span {
		display: block;
		height: 1.5px;
		background: var(--text-2);
		border-radius: 1px;
		transition: transform .2s ease, opacity .2s ease;
	}
	.d-burger.open span:nth-child(1) { transform: translateY(6.5px) rotate(45deg); }
	.d-burger.open span:nth-child(2) { opacity: 0; }
	.d-burger.open span:nth-child(3) { transform: translateY(-6.5px) rotate(-45deg); }

	.d-menu-overlay {
		position: fixed;
		inset: 56px 0 0 0;
		z-index: 100;
		background: var(--bg);
		overflow-y: auto;
		animation: menu-in .18s ease;
	}
	@keyframes menu-in {
		from { opacity: 0; transform: translateY(-8px); }
		to   { opacity: 1; transform: translateY(0); }
	}
	.d-menu-inner {
		display: flex;
		flex-direction: column;
		max-width: 720px;
		margin: 0 auto;
		padding: 24px 24px 48px;
	}
	.d-menu-links {
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding-bottom: 24px;
		border-bottom: 1px solid var(--border);
		margin-bottom: 24px;
	}
	.d-menu-link {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 16px 4px;
		font-size: 22px;
		font-weight: 500;
		letter-spacing: -0.01em;
		color: var(--text);
		background: none;
		border: none;
		border-bottom: 1px solid var(--border);
		text-align: left;
		cursor: pointer;
		transition: color .12s ease;
	}
	.d-menu-link:last-child { border-bottom: none; }
	.d-menu-link:hover { color: var(--accent); }
	.d-menu-link.active { color: var(--accent); }
	.d-menu-link :global(svg) { color: var(--text-3); }
	.d-menu-actions {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.d-menu-action {
		display: inline-flex;
		align-items: center;
		gap: 12px;
		padding: 12px 16px;
		font-size: 14.5px;
		color: var(--text-2);
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		cursor: pointer;
		text-align: left;
		transition: color .12s ease, border-color .12s ease, background .12s ease;
	}
	.d-menu-action:hover {
		color: var(--accent);
		border-color: var(--accent);
		background: color-mix(in oklch, var(--accent) 6%, var(--surface));
	}
	.d-menu-action :global(svg) { color: var(--text-3); }
	.d-menu-action:hover :global(svg) { color: var(--accent); }
	.d-menu-locale {
		font-family: var(--font-mono);
		font-size: 11px;
		font-weight: 600;
		letter-spacing: 0.06em;
		padding: 3px 8px;
		border: 1px solid var(--border);
		border-radius: 4px;
		color: var(--text-3);
	}
	.d-menu-cta {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		padding: 14px 20px;
		background: var(--accent);
		color: var(--accent-on);
		border-radius: var(--r-md);
		font-weight: 600;
		font-size: 15px;
		border: none;
		cursor: pointer;
		margin-top: 8px;
		transition: filter .12s ease;
	}
	.d-menu-cta:hover { filter: brightness(1.05); }
	.d-menu-cta.ghost {
		background: none;
		color: var(--text-2);
		border: 1px solid var(--border);
		font-weight: 500;
		margin-top: 0;
	}
	.d-menu-cta.ghost:hover { color: var(--text); border-color: var(--border-2); }

	@media (max-width: 1100px) {
		.d-nav { gap: 16px; padding: 12px 20px; }
		.d-nav-links,
		.d-nav-cta { display: none !important; }
		.d-burger { display: flex; }
	}
</style>
