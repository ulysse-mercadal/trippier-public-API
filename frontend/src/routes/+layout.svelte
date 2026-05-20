<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { browser } from '$app/environment';
	import { env } from '$env/dynamic/public';
	import TopoBackground from '$lib/components/TopoBackground.svelte';
	import { t, locale, setLocale, initLocale, type Locale } from '$lib/i18n';

	const docsUrl = env.PUBLIC_DOCS_URL ?? 'http://localhost:5173';

	$: isLogin     = $page.url.pathname === '/login';
	$: isDocs      = $page.url.pathname.startsWith('/docs');
	$: isDashboard = $page.url.pathname.startsWith('/dashboard');

	onMount(async () => {
		initLocale();
		if ($auth.user) return;
		const token = auth.getStoredToken();
		if (!token) return;
		try {
			const res = await fetch('/api/auth/me', { headers: { Authorization: `Bearer ${token}` } });
			if (res.ok) auth.init(token, await res.json());
		} catch {}
	});

	function handleLogout() {
		auth.logout();
		goto('/');
	}
</script>

<svelte:head>
	<link rel="icon" type="image/png" href="/favicon.png">
	<link rel="preconnect" href="https://fonts.googleapis.com">
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous">
	<link href="https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
</svelte:head>

{#if isLogin || isDocs}
	<slot />
{:else}
	{#if browser}<TopoBackground density={28} opacity={0.24} color="#4ee39a" />{/if}

	<div class="page">
		<header class="nav">
			<a class="brand" href="/">
				<img src="/favicon.png" alt="trippier logo" width="88" height="88" style="border-radius:3px" />
				<span>trippier<span class="brand-dot">/</span>api</span>
			</a>

			<nav class="nav-links">
				<a href="/#features">{$t('nav_features')}</a>
				<a href="/#deploy">{$t('nav_deploy')}</a>
				<a href="/docs">{$t('nav_docs')}</a>
				<a href="/roadmap">{$t('nav_roadmap')}</a>
				<a href="https://github.com/ulysse-mercadal/trippier-public-API" target="_blank" rel="noopener" class="nav-gh">
					<svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M12 2a10 10 0 0 0-3.16 19.49c.5.09.68-.22.68-.48v-1.7c-2.78.6-3.37-1.34-3.37-1.34-.45-1.16-1.11-1.47-1.11-1.47-.91-.62.07-.61.07-.61 1 .07 1.53 1.03 1.53 1.03.89 1.53 2.34 1.09 2.91.83.09-.65.35-1.09.63-1.34-2.22-.25-4.55-1.11-4.55-4.94 0-1.09.39-1.99 1.03-2.69-.1-.25-.45-1.27.1-2.65 0 0 .84-.27 2.75 1.02a9.6 9.6 0 0 1 5 0c1.91-1.29 2.75-1.02 2.75-1.02.55 1.38.2 2.4.1 2.65.64.7 1.03 1.6 1.03 2.69 0 3.84-2.34 4.68-4.57 4.93.36.31.68.92.68 1.85v2.74c0 .27.18.58.69.48A10 10 0 0 0 12 2z"/></svg>
					GitHub
				</a>
			</nav>

			<div class="nav-cta">
				<button class="locale-btn" on:click={() => setLocale($locale === 'fr' ? 'en' : 'fr')}>
					{$locale === 'fr' ? 'EN' : 'FR'}
				</button>
				{#if $auth.user}
					<a href="/dashboard" class="btn-ghost">{$t('nav_dashboard')}</a>
					<button class="btn-ghost" on:click={handleLogout}>{$t('nav_logout')}</button>
				{:else if isDashboard}
					<!-- redirected by dashboard itself -->
				{:else}
					<a href="/login" class="btn-ghost">{$t('nav_login')}</a>
				{/if}
			</div>
		</header>

		<main>
			<slot />
		</main>
	</div>
{/if}

<style>
	.page {
		position: relative;
		z-index: 1;
		min-height: 100vh;
	}

	.nav {
		display: flex;
		align-items: center;
		gap: 32px;
		padding: 22px 32px;
		margin: 8px auto 0;
		max-width: var(--max);
		position: relative;
		z-index: 10;
	}

	.brand {
		display: inline-flex;
		align-items: center;
		gap: 10px;
		font-weight: 600;
		font-size: 16px;
		color: var(--text);
		letter-spacing: -0.01em;
		flex-shrink: 0;
	}
	.brand svg { color: var(--accent); flex-shrink: 0; }
	.brand-dot { color: var(--text-3); margin: 0 1px; }

	.nav-links {
		display: flex;
		gap: 28px;
		flex: 1;
		margin-left: 24px;
	}
	.nav-links a {
		font-size: 14px;
		color: var(--text-2);
		transition: color .15s ease;
		display: inline-flex;
		align-items: center;
		gap: 6px;
	}
	.nav-links a:hover { color: var(--text); }

	.nav-cta {
		display: flex;
		gap: 10px;
		align-items: center;
	}
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

	main { position: relative; }

	@media (max-width: 980px) {
		.nav-links { display: none; }
	}
</style>
