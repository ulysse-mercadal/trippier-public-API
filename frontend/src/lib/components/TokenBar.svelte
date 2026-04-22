<script lang="ts">
	export let remaining: number;
	export let limit: number;
	export let resetsInSecs: number;

	$: pct     = limit > 0 ? Math.max(0, Math.min(100, (remaining / limit) * 100)) : 0;
	$: variant = pct > 25 ? 'ok' : pct > 10 ? 'warn' : 'empty';

	$: resetLabel = (() => {
		const m = Math.ceil(resetsInSecs / 60);
		return m > 60 ? `${Math.ceil(m / 60)}h` : `${m}m`;
	})();
</script>

<div class="usage">
	<div class="usage-meta">
		<span class="count"><strong>{remaining}</strong> / {limit} tokens</span>
		<span class="reset">Resets in {resetLabel}</span>
	</div>
	<div class="track">
		<div class="fill fill-{variant}" style="width: {pct}%" />
	</div>
</div>

<style>
	.usage { display: flex; flex-direction: column; gap: 0.5rem; }

	.usage-meta {
		display: flex;
		justify-content: space-between;
		font-size: 0.8rem;
	}
	.count { color: var(--text); }
	.count strong { font-weight: 700; }
	.reset { color: var(--muted); }

	.track {
		background: var(--bg);
		border-radius: 999px;
		height: 7px;
		overflow: hidden;
	}

	.fill {
		height: 100%;
		border-radius: 999px;
		transition: width 0.5s ease;
	}
	.fill-ok    { background: var(--accent); }
	.fill-warn  { background: #f59e0b; }
	.fill-empty { background: var(--danger); }
</style>
