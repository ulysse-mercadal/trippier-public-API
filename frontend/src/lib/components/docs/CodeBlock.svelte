<script lang="ts">
	import { browser } from '$app/environment';

	export let code: string;
	export let lang: string;

	let copied = false;

	// Copies code to clipboard and shows brief feedback
	function copy() {
		if (browser) navigator.clipboard?.writeText(code);
		copied = true;
		setTimeout(() => (copied = false), 1200);
	}
</script>

<div class="d-code">
	<div class="d-code-bar">
		<span class="d-code-lang">{lang}</span>
		<button on:click={copy}>
			{#if copied}
				<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12l5 5 9-11"/></svg>
				copié
			{:else}
				<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="9" y="9" width="11" height="11" rx="2"/><path d="M5 15V5a2 2 0 0 1 2-2h10"/></svg>
				copier
			{/if}
		</button>
	</div>
	<pre>{code}</pre>
</div>

<style>
	.d-code {
		margin: 8px 0 16px;
		border: 1px solid var(--code-border);
		border-radius: var(--r-md);
		overflow: hidden;
		background: var(--code-bg-2);
	}
	.d-code-bar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 8px 14px;
		background: var(--code-surface);
		border-bottom: 1px solid var(--code-border);
	}
	.d-code-lang {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--code-text-3);
		text-transform: lowercase;
		letter-spacing: 0.04em;
	}
	.d-code-bar button {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--code-text-3);
		padding: 3px 8px;
		border-radius: 4px;
		border: none;
		background: none;
		cursor: pointer;
		transition: color .12s ease, background .12s ease;
	}
	.d-code-bar button:hover { color: var(--code-text); background: var(--code-bg); }
	pre {
		margin: 0;
		padding: 14px 18px;
		font-family: var(--font-mono);
		font-size: 12.5px;
		color: var(--code-text-2);
		line-height: 1.6;
		white-space: pre-wrap;
		word-break: break-word;
		overflow-x: auto;
	}
</style>
