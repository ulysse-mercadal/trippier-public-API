<script lang="ts">
	import '../app.css';
	import TopoCanvas from '$lib/components/TopoCanvas.svelte';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';

	$: isDashboard = $page.url.pathname.startsWith('/dashboard');

	function handleLogout() {
		auth.logout();
		goto('/');
	}
</script>

<TopoCanvas />

<div class="layout">
	<nav>
		<a href="/" class="logo">tripp<em>ier</em></a>

		<div class="nav-right">
			{#if $auth.user}
				<span class="nav-email">{$auth.user.email}</span>
				<button class="btn btn-ghost btn-sm" on:click={handleLogout}>Sign out</button>
			{:else if isDashboard}
				<!-- nothing: dashboard redirects if unauthed -->
			{:else}
				<a href="#auth" class="btn btn-ghost">Sign in</a>
				<a href="#auth" class="btn btn-primary">Get API key</a>
			{/if}
		</div>
	</nav>

	<main>
		<slot />
	</main>
</div>

<style>
	.layout {
		position: relative;
		z-index: 1;
		min-height: 100vh;
	}

	nav {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1.25rem 2rem;
		border-bottom: 1px solid var(--border);
		backdrop-filter: blur(8px);
		background: rgba(5, 5, 5, 0.7);
		position: sticky;
		top: 0;
		z-index: 100;
	}

	.logo {
		font-size: 1.1rem;
		font-weight: 700;
		letter-spacing: -0.02em;
		color: var(--text);
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
		color: var(--muted);
	}

	main { position: relative; }
</style>
