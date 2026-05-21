<svelte:head>
	<title>trippier/api · l'API de voyage open source</title>
</svelte:head>

<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import Hero     from '$lib/components/landing/Hero.svelte';
	import Features from '$lib/components/landing/Features.svelte';
	import Console  from '$lib/components/landing/Console.svelte';
	import Deploy   from '$lib/components/landing/Deploy.svelte';
	import Byok     from '$lib/components/landing/Byok.svelte';
	import Routes   from '$lib/components/landing/Routes.svelte';
	import Pricing  from '$lib/components/landing/Pricing.svelte';
	import Footer   from '$lib/components/landing/Footer.svelte';
	import { activeNavHash } from '$lib/stores/nav';

	let obs: IntersectionObserver;

	onMount(() => {
		obs = new IntersectionObserver(entries => {
			for (const e of entries) {
				if (e.isIntersecting) activeNavHash.set(`/#${e.target.id}`);
			}
		}, { threshold: 0.4 });
		['features', 'deploy'].forEach(id => {
			const el = document.getElementById(id);
			if (el) obs.observe(el);
		});
	});

	onDestroy(() => {
		obs?.disconnect();
		activeNavHash.set('');
	});
</script>

<Hero />
<Features />
<Console />
<Deploy />
<Byok />
<Routes />
<Pricing />
<Footer />

<style>
	:global(.section-head) {
		max-width: 760px;
		margin-bottom: 48px;
	}
	:global(.section-head h2) {
		font-size: clamp(28px, 3.2vw, 44px);
		letter-spacing: -0.02em;
		line-height: 1.08;
		margin: 8px 0 0;
		font-weight: 600;
		text-wrap: balance;
	}
	:global(.section-sub) {
		color: var(--text-2);
		font-size: 17px;
		margin-top: 14px;
		max-width: 640px;
		line-height: 1.5;
	}
</style>
