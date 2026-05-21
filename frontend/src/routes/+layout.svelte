<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { auth } from '$lib/stores/auth';
	import { page } from '$app/stores';
	import { browser } from '$app/environment';
	import TopoBackground from '$lib/components/TopoBackground.svelte';
	import DocsNav from '$lib/components/docs/DocsNav.svelte';
	import { initLocale } from '$lib/i18n';

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
		<DocsNav />
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
	main { position: relative; }
</style>
