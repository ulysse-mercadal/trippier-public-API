<script lang="ts">
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { t, locale, setLocale } from '$lib/i18n';
	import { page } from '$app/stores';
	import { activeNavHash } from '$lib/stores/nav';
	import { theme, toggleTheme } from '$lib/stores/theme';

	export let active: string = '';

	function handleLogout() {
		auth.logout();
		goto('/');
	}

	$: currentActive = active || $activeNavHash || $page.url.pathname;
</script>

<nav class="d-nav">
	<a class="d-brand" href="/">
		<img src="/favicon.png" alt="trippier logo" width="28" height="28" style="border-radius:3px" />
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
		<button class="locale-btn" on:click={() => setLocale($locale === 'fr' ? 'en' : 'fr')}>
			{$locale === 'fr' ? 'EN' : 'FR'}
		</button>
		{#if $auth.user}
			<a href="/dashboard" class="btn-ghost">{$t('nav_dashboard')}</a>
			<button class="btn-ghost" on:click={handleLogout}>{$t('nav_logout')}</button>
		{:else}
			<a href="/login" class="btn-ghost">{$t('nav_login')}</a>
		{/if}
	</div>
</nav>

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

	@media (max-width: 800px) {
		.d-nav-links { display: none; }
	}
</style>
