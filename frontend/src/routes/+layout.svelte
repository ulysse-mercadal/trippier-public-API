<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { browser } from '$app/environment';
	import { env } from '$env/dynamic/public';
	import TopoCanvas from '$lib/components/TopoCanvas.svelte';

	const docsUrl = env.PUBLIC_DOCS_URL ?? 'http://localhost:5173';

	$: isLogin     = $page.url.pathname === '/login';
	$: isDashboard = $page.url.pathname.startsWith('/dashboard');

	let scrolled = false;

	onMount(() => {
		const onScroll = () => { scrolled = window.scrollY > 40; };
		window.addEventListener('scroll', onScroll, { passive: true });
		return () => window.removeEventListener('scroll', onScroll);
	});

	function handleLogout() {
		auth.logout();
		goto('/');
	}
</script>

{#if isLogin}
	<slot />
{:else}
	{#if browser}<TopoCanvas fixed={true} dark={false} />{/if}

	<div class="layout">
		<nav class:scrolled>
			<a href="/" class="logo">tripp<em>ier</em></a>
			<div class="nav-right">
				<a href={docsUrl} class="btn btn-ghost btn-sm" target="_blank" rel="noopener noreferrer">Docs</a>
				{#if $auth.user}
					<span class="nav-email">{$auth.user.email}</span>
					<button class="btn btn-ghost btn-sm" on:click={handleLogout}>Sign out</button>
				{:else if isDashboard}
					<!-- dashboard redirects if unauthed -->
				{:else}
					<a href="/login" class="btn btn-ghost btn-sm nav-signin">Sign in</a>
					<a href="/login" class="btn btn-primary btn-sm">Get API key</a>
				{/if}
			</div>
		</nav>

		<main>
			<slot />
		</main>
	</div>
{/if}

<style>
	.layout {
		position: relative;
		z-index: 1;
		min-height: 100vh;
		/* padding-top so content doesn't hide under the fixed nav */
		padding-top: 80px;
	}

	nav {
		position: fixed;
		top: 1rem;
		left: 1.25rem;
		right: 1.25rem;
		z-index: 100;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 1.75rem;
		height: 60px;
		background: #0a0a0a;
		border-radius: 999px;
		border: 1px solid rgba(255, 255, 255, 0.1);
		box-shadow: 0 8px 32px rgba(0, 0, 0, 0.35);
		transition: top 0.3s ease, height 0.3s ease, padding 0.3s ease, box-shadow 0.3s ease;
	}

	nav.scrolled {
		top: 0.5rem;
		height: 46px;
		padding: 0 1.25rem;
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.45);
	}

	.logo {
		font-size: 1.05rem;
		font-weight: 700;
		letter-spacing: -0.02em;
		color: #f0f0f0;
	}
	.logo em {
		font-style: normal;
		color: var(--accent);
	}

	.nav-right {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.nav-email {
		font-size: 0.8rem;
		color: rgba(255,255,255,0.5);
	}

	/* Ghost btn override for dark nav */
	:global(nav .btn-ghost) {
		color: rgba(255,255,255,0.65);
		border-color: rgba(255,255,255,0.12);
	}
	:global(nav .btn-ghost:hover) {
		color: #fff;
		border-color: rgba(255,255,255,0.25);
	}

	main { position: relative; }
</style>
