<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { t } from '$lib/i18n';

	const CURL_CMD = `curl https://api.trippier.dev/itinerary/generate \\
  -H "X-API-Key: tk_demo_a3f9…" \\
  -d '{"days":1,"poi_query":{"lat":40.7549,"lng":-73.9840,"radius":5000},"preferences":{"pace":"moderate","start_time":"09:00"}}'`;

	const RESPONSE_JSON = `{
  "days": [
    {
      "day": 1,
      "pois": [
        { "id": "osm_empire_state",  "name": "Empire State Building", "type": "monument" },
        { "id": "osm_central_park",  "name": "Central Park",          "type": "park" },
        { "id": "osm_moma",          "name": "MoMA",                  "type": "museum" },
        { "id": "osm_high_line",     "name": "The High Line",         "type": "viewpoint" }
      ],
      "estimated_duration_hours": 6.5,
      "description": "Icons of Midtown & Upper West Side"
    }
  ],
  "total_pois": 4,
  "summary": "Manhattan highlights, moderate pace"
}`;

	type Phase = 'typing' | 'pending' | 'done';

	let typed    = '';
	let phase: Phase = 'typing';
	let revealed = 0;
	let timer: ReturnType<typeof setTimeout> | null = null;

	function tick() {
		if (phase === 'typing') {
			if (typed.length < CURL_CMD.length) {
				timer = setTimeout(() => { typed = CURL_CMD.slice(0, typed.length + 1); tick(); }, 18);
			} else {
				timer = setTimeout(() => { phase = 'pending'; tick(); }, 400);
			}
		} else if (phase === 'pending') {
			timer = setTimeout(() => { phase = 'done'; tick(); }, 850);
		} else if (phase === 'done' && revealed < RESPONSE_JSON.length) {
			timer = setTimeout(() => {
				revealed = Math.min(RESPONSE_JSON.length, revealed + 18);
				if (revealed < RESPONSE_JSON.length) tick();
			}, 10);
		}
	}

	function replay() {
		if (timer) { clearTimeout(timer); timer = null; }
		typed = ''; phase = 'typing'; revealed = 0;
		tick();
	}

	onMount(() => tick());
	onDestroy(() => { if (timer) clearTimeout(timer); });
</script>

<section class="console-section">
	<div class="wrap">
		<div class="section-head">
			<span class="eyebrow">{$t('console_eyebrow')}</span>
			<h2>{$t('console_title')}</h2>
		</div>
		<div class="terminal">
			<div class="terminal-bar">
				<span class="tdot" style="background:#e78f8f"></span>
				<span class="tdot" style="background:#e4c87a"></span>
				<span class="tdot" style="background:var(--accent)"></span>
				<span class="terminal-title">~ / trippier · zsh</span>
				<button class="terminal-replay" on:click={replay}>{$t('console_replay')}</button>
			</div>
			<div class="terminal-body">
				<pre class="t-cmd"><span class="tprompt">›</span> {typed}{#if phase === 'typing'}<span class="caret">▍</span>{/if}</pre>
				{#if phase !== 'typing'}
					<div class="t-status">
						{#if phase === 'pending'}
							<span class="status-pending">… POST /itinerary/generate</span>
						{:else}
							<span class="status-ok">HTTP/2 200 · 87ms · application/json</span>
						{/if}
					</div>
				{/if}
				{#if phase === 'done'}
					<pre class="t-resp">{RESPONSE_JSON.slice(0, revealed)}{#if revealed < RESPONSE_JSON.length}<span class="caret">▍</span>{/if}</pre>
				{/if}
			</div>
		</div>
	</div>
</section>

<style>
	.console-section { padding-block: 96px; }
	.terminal {
		border: 1px solid var(--border);
		border-radius: var(--r-lg);
		background: oklch(10% 0.01 175);
		overflow: hidden;
		box-shadow: 0 20px 60px -20px rgba(0,0,0,0.6);
	}
	.terminal-bar {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 12px 16px;
		background: var(--surface);
		border-bottom: 1px solid var(--border);
	}
	.tdot { width: 11px; height: 11px; border-radius: 50%; flex-shrink: 0; }
	.terminal-title {
		margin-left: 12px;
		font-family: var(--font-mono);
		font-size: 12px;
		color: var(--text-3);
		flex: 1;
	}
	.terminal-replay {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--text-3);
		padding: 4px 8px;
		border-radius: 4px;
		transition: background .15s ease, color .15s ease;
		cursor: pointer;
	}
	.terminal-replay:hover { background: var(--bg); color: var(--text); }
	.terminal-body {
		padding: 20px 22px;
		font-family: var(--font-mono);
		font-size: 13px;
		line-height: 1.55;
		min-height: 320px;
	}
	.t-cmd { margin: 0; color: var(--text); white-space: pre-wrap; word-break: break-all; }
	.tprompt { color: var(--accent); margin-right: 6px; }
	.caret {
		display: inline-block;
		width: 8px;
		color: var(--accent);
		animation: blink 1s steps(2) infinite;
	}
	@keyframes blink { 50% { opacity: 0; } }
	.t-status { margin-top: 14px; font-size: 12px; color: var(--text-3); }
	.status-pending { color: var(--text-3); }
	.status-pending::before {
		content: '';
		display: inline-block;
		width: 8px; height: 8px;
		border-radius: 50%;
		background: var(--warn);
		margin-right: 8px;
		vertical-align: middle;
		animation: pulse-warn 1s ease-in-out infinite;
	}
	@keyframes pulse-warn { 50% { opacity: 0.3; } }
	.status-ok { color: var(--accent); }
	.status-ok::before { content: '●'; margin-right: 8px; }
	.t-resp { margin: 14px 0 0; color: var(--text-2); white-space: pre-wrap; }
</style>
