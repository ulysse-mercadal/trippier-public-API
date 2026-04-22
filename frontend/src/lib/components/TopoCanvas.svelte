<script lang="ts">
	import { onMount, onDestroy } from 'svelte';

	let canvas: HTMLCanvasElement;
	let animFrame: number;

	interface Hill {
		fx: number; fy: number;
		a: number;  b: number;
		angle: number;
		levels: number;
	}

	const HILLS: Hill[] = [
		{ fx: 0.10, fy: 0.22, a: 0.19, b: 0.13, angle: -0.4, levels: 10 },
		{ fx: 0.74, fy: 0.55, a: 0.23, b: 0.16, angle:  0.6, levels: 12 },
		{ fx: 0.90, fy: 0.10, a: 0.13, b: 0.08, angle:  0.2, levels: 7  },
		{ fx: 0.04, fy: 0.85, a: 0.15, b: 0.10, angle: -0.1, levels: 9  },
		{ fx: 0.50, fy: 0.92, a: 0.18, b: 0.09, angle:  0.9, levels: 7  },
		{ fx: 0.36, fy: 0.42, a: 0.10, b: 0.07, angle:  0.3, levels: 6  },
	];

	function drawEllipse(
		ctx: CanvasRenderingContext2D,
		cx: number, cy: number,
		rx: number, ry: number,
		angle: number, distort: number,
	) {
		const STEPS = 80;
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
			const rx = h.a  * W;
			const ry = h.b  * H;

			for (let i = 1; i <= h.levels; i++) {
				const s       = i / h.levels;
				const opacity = 0.03 + s * 0.075;
				ctx.strokeStyle = `rgba(255,255,255,${opacity.toFixed(3)})`;
				ctx.lineWidth   = 0.8;
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
		cancelAnimationFrame(animFrame);
	});
</script>

<canvas bind:this={canvas} />

<style>
	canvas {
		position: fixed;
		inset: 0;
		z-index: 0;
		pointer-events: none;
	}
</style>
