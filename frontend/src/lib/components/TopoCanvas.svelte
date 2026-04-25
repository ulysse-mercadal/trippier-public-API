<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { ContourLevel } from '$lib/topo-data';

	let canvas: HTMLCanvasElement;
	let contours: ContourLevel[] = [];
	let minElev = 0;
	let maxElev = 1;

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
		if (!ctx) return;

		const W = canvas.width;
		const H = canvas.height;
		ctx.clearRect(0, 0, W, H);
		for (const level of contours) {
			const t = (level.elevation - minElev) / (maxElev - minElev);
			ctx.strokeStyle = `rgba(255,255,255,${(0.10 + t * 0.28).toFixed(3)})`;
			ctx.lineWidth   = 0.8 + t * 0.9;
			for (const polygon of level.polygons) {
				for (const ring of polygon) {
					drawSmooth(ctx, ring, W, H);
				}
			}
		}
	}

	function resize() {
		canvas.width  = window.innerWidth;
		canvas.height = window.innerHeight;
		draw();
	}

	onMount(async () => {
		const { MONT_BLANC_CONTOURS } = await import('$lib/topo-data');
		contours = MONT_BLANC_CONTOURS;
		const elevations = contours.map(c => c.elevation).sort((a, b) => a - b);
		minElev = elevations[0];
		maxElev = elevations[elevations.length - 1];
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
