<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { ContourLevel } from '$lib/topo-data';

	/** true  → position:fixed, white lines on dark bg (hero)
	 *  false → position:absolute, fills parent, dark lines on light bg (login panel) */
	export let fixed = true;
	/** true = white lines (dark bg), false = gray lines (light bg) */
	export let dark  = true;

	let canvas: HTMLCanvasElement;
	let contours: ContourLevel[] = [];
	let minElev = 0;
	let maxElev = 1;
	let ro: ResizeObserver | null = null;

	function drawSmooth(ctx: CanvasRenderingContext2D, ring: [number, number][], W: number, H: number) {
		const n = ring.length;
		if (n < 2) return;
		const px = (i: number) => ring[i % n][0] * W;
		const py = (i: number) => ring[i % n][1] * H;
		ctx.beginPath();
		ctx.moveTo((px(n - 1) + px(0)) / 2, (py(n - 1) + py(0)) / 2);
		for (let i = 0; i < n; i++) {
			const mx = (px(i) + px(i + 1)) / 2;
			const my = (py(i) + py(i + 1)) / 2;
			ctx.quadraticCurveTo(px(i), py(i), mx, my);
		}
		ctx.closePath();
		ctx.stroke();
	}

	function draw() {
		const ctx = canvas.getContext('2d');
		if (!ctx || canvas.width === 0 || canvas.height === 0) return;
		const W = canvas.width;
		const H = canvas.height;
		ctx.clearRect(0, 0, W, H);
		for (const level of contours) {
			const t = (level.elevation - minElev) / (maxElev - minElev);
			if (dark) {
				ctx.strokeStyle = `rgba(255,255,255,${(0.10 + t * 0.28).toFixed(3)})`;
				ctx.lineWidth   = 0.8 + t * 0.9;
			} else {
				ctx.strokeStyle = `rgba(60,60,60,${(0.10 + t * 0.22).toFixed(3)})`;
				ctx.lineWidth   = 0.7 + t * 0.8;
			}
			for (const polygon of level.polygons) {
				for (const ring of polygon) drawSmooth(ctx, ring, W, H);
			}
		}
	}

	function resize() {
		if (fixed) {
			canvas.width  = window.innerWidth;
			canvas.height = window.innerHeight;
		} else {
			// read the PARENT dimensions — reliable regardless of canvas's own layout state
			const p = canvas.parentElement;
			if (!p) return;
			canvas.width  = p.clientWidth;
			canvas.height = p.clientHeight;
		}
		draw();
	}

	onMount(async () => {
		const { MONT_BLANC_CONTOURS } = await import('$lib/topo-data');
		contours = MONT_BLANC_CONTOURS;
		const elevations = contours.map(c => c.elevation).sort((a, b) => a - b);
		minElev = elevations[0];
		maxElev = elevations[elevations.length - 1];

		if (fixed) {
			resize();
			window.addEventListener('resize', resize);
		} else {
			const parent = canvas.parentElement;
			if (parent) {
				ro = new ResizeObserver(() => resize());
				ro.observe(parent);
			}
			resize();
		}
	});

	onDestroy(() => {
		window.removeEventListener('resize', resize);
		ro?.disconnect();
	});
</script>

<canvas
	bind:this={canvas}
	style:position={fixed ? 'fixed' : 'absolute'}
	style:inset="0"
></canvas>

<style>
	canvas {
		z-index: 0;
		pointer-events: none;
	}
</style>
