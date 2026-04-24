<script lang="ts">
	import { onMount, onDestroy } from 'svelte';

	let canvas: HTMLCanvasElement;

	// Deterministic pseudo-random seeded by grid indices + channel
	function seeded(c: number, r: number, i: number): number {
		const x = Math.sin(c * 127.1 + r * 311.7 + i * 74.3) * 43758.5453;
		return x - Math.floor(x);
	}

	// Grid layout — 5×4 cells covering the full canvas with overlap
	const COLS = 5;
	const ROWS = 4;

	// Normalised params computed once: positions as fractions of canvas size,
	// radii as fractions of width/height, angle in radians, level count.
	interface HillNorm { fx: number; fy: number; arx: number; ary: number; angle: number; levels: number }
	const HILLS: HillNorm[] = [];
	for (let r = 0; r < ROWS; r++) {
		for (let c = 0; c < COLS; c++) {
			HILLS.push({
				fx:     (c + 0.5 + (seeded(c, r, 0) - 0.5) * 0.45) / COLS,
				fy:     (r + 0.5 + (seeded(c, r, 1) - 0.5) * 0.45) / ROWS,
				arx:    0.11 + seeded(c, r, 2) * 0.08,
				ary:    0.07 + seeded(c, r, 3) * 0.06,
				angle:  (seeded(c, r, 4) - 0.5) * Math.PI,
				levels: 7 + Math.floor(seeded(c, r, 5) * 6),
			});
		}
	}

	function drawEllipse(
		ctx: CanvasRenderingContext2D,
		cx: number, cy: number,
		rx: number, ry: number,
		angle: number, distort: number,
	) {
		const STEPS = 90;
		ctx.beginPath();
		for (let i = 0; i <= STEPS; i++) {
			const t = (i / STEPS) * Math.PI * 2;
			const d = 1 + distort * (Math.sin(t * 3 + 0.7) * 0.12 + Math.cos(t * 5 - 1.1) * 0.09);
			const lx = rx * d * Math.cos(t);
			const ly = ry * d * Math.sin(t);
			const x  = cx + lx * Math.cos(angle) - ly * Math.sin(angle);
			const y  = cy + lx * Math.sin(angle) + ly * Math.cos(angle);
			i === 0 ? ctx.moveTo(x, y) : ctx.lineTo(x, y);
		}
		ctx.closePath();
		ctx.stroke();
	}

	function draw() {
		const ctx = canvas.getContext('2d');
		if (!ctx) return;

		const W = canvas.width;
		const H = canvas.height;
		ctx.clearRect(0, 0, W, H);

		for (const h of HILLS) {
			const cx = h.fx * W;
			const cy = h.fy * H;
			const rx = h.arx * W;
			const ry = h.ary * H;

			for (let i = 1; i <= h.levels; i++) {
				const s       = i / h.levels;
				// Inner rings dim, outer rings brighter — classic topo look
				const opacity = 0.04 + s * 0.13;
				ctx.strokeStyle = `rgba(255,255,255,${opacity.toFixed(3)})`;
				ctx.lineWidth   = 0.75;
				drawEllipse(ctx, cx, cy, rx * s, ry * s, h.angle, 1 - s * 0.5);
			}
		}
	}

	function resize() {
		canvas.width  = window.innerWidth;
		canvas.height = window.innerHeight;
		draw();
	}

	onMount(() => {
		resize();
		window.addEventListener('resize', resize);
	});

	onDestroy(() => {
		window.removeEventListener('resize', resize);
	});
</script>

<canvas bind:this={canvas}></canvas>

<style>
	canvas {
		position: fixed;
		inset: 0;
		z-index: 0;
		pointer-events: none;
	}
</style>
