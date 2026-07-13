<script lang="ts">
	/** Number of contour levels (design default: 28) */
	export let density: number = 28;
	/** Base stroke opacity (design default: 0.24) */
	export let opacity: number = 0.24;
	/** Stroke color hex */
	export let color: string = '#4ee39a';
	/** Seed for the deterministic PRNG (changes the mountain pattern) */
	export let seed: number = 7;

	const W    = 1600;
	const H    = 1100;
	const cell = 14;
	const cols = Math.ceil(W / cell);
	const rows = Math.ceil(H / cell);

	/**
	 * Creates a deterministic linear-congruential pseudo-random number generator.
	 * @param s seed value
	 * @returns function producing the next random number in [0, 1)
	 */
	function makePrng(s: number) {
		let state = s * 9301 + 49297;
		return () => {
			state = (state * 9301 + 49297) % 233280;
			return state / 233280;
		};
	}

	/**
	 * Computes seven Gaussian peaks/valleys seeded from `seed`.
	 * @returns array of peak descriptors {x, y, sigma, amp}
	 */
	const peaks = (() => {
		const r = makePrng(seed);
		const arr: { x: number; y: number; sigma: number; amp: number }[] = [];
		for (let i = 0; i < 7; i++) {
			arr.push({
				x:     r() * W,
				y:     r() * H,
				sigma: 180 + r() * 220,
				amp:   (r() > 0.3 ? 1 : -0.6) * (0.7 + r() * 0.6),
			});
		}
		return arr;
	})();

	/**
	 * Evaluates the heightfield value at a point by summing Gaussian peaks.
	 * @param x x coordinate
	 * @param y y coordinate
	 * @returns height value at (x, y)
	 */
	function heightAt(x: number, y: number): number {
		let h = 0;
		for (const p of peaks) {
			const dx = x - p.x, dy = y - p.y;
			h += p.amp * Math.exp(-(dx * dx + dy * dy) / (2 * p.sigma * p.sigma));
		}
		return h;
	}

	/**
	 * Samples the heightfield on a grid and tracks its min/max values.
	 * @returns grid samples with their min and max height
	 */
	const { f, fMin, fMax } = (() => {
		const f = new Float32Array((cols + 1) * (rows + 1));
		let fMin = Infinity, fMax = -Infinity;
		for (let j = 0; j <= rows; j++) {
			for (let i = 0; i <= cols; i++) {
				const v = heightAt(i * cell, j * cell);
				f[j * (cols + 1) + i] = v;
				if (v < fMin) fMin = v;
				if (v > fMax) fMax = v;
			}
		}
		return { f, fMin, fMax };
	})();

	/** One SVG path string for a given contour level index. */
	type PathEntry = { d: string; idx: number };

	/**
	 * Runs marching squares over the sampled heightfield to build contour paths.
	 * @returns list of path entries, one per contour level
	 */
	const paths: PathEntry[] = (() => {
		const levels: number[] = [];
		for (let k = 1; k < density; k++) {
			levels.push(fMin + (fMax - fMin) * (k / density));
		}
		const stride = cols + 1;
		return levels.map((level, idx) => {
			let d = '';
			for (let j = 0; j < rows; j++) {
				for (let i = 0; i < cols; i++) {
					const tl = f[j * stride + i];
					const tr = f[j * stride + (i + 1)];
					const br = f[(j + 1) * stride + (i + 1)];
					const bl = f[(j + 1) * stride + i];
					let key = 0;
					if (tl > level) key |= 8;
					if (tr > level) key |= 4;
					if (br > level) key |= 2;
					if (bl > level) key |= 1;
					if (key === 0 || key === 15) continue;
					const x = i * cell, y = j * cell;
					const lerp = (a: number, b: number) => (level - a) / (b - a);
					const top:   [number, number] = [x + cell * lerp(tl, tr), y];
					const right: [number, number] = [x + cell, y + cell * lerp(tr, br)];
					const bot:   [number, number] = [x + cell * lerp(bl, br), y + cell];
					const left:  [number, number] = [x, y + cell * lerp(tl, bl)];
					const seg = (a: [number, number], b: [number, number]) => {
						d += `M${a[0].toFixed(1)} ${a[1].toFixed(1)}L${b[0].toFixed(1)} ${b[1].toFixed(1)}`;
					};
					switch (key) {
						case  1: case 14: seg(left, bot);   break;
						case  2: case 13: seg(bot, right);  break;
						case  3: case 12: seg(left, right); break;
						case  4: case 11: seg(top, right);  break;
						case  5: seg(left, top); seg(bot, right); break;
						case  6: case  9: seg(top, bot);   break;
						case  7: case  8: seg(left, top);  break;
						case 10: seg(left, bot); seg(top, right); break;
					}
				}
			}
			return { d, idx };
		});
	})();
</script>

<svg
	viewBox={`0 0 ${W} ${H}`}
	preserveAspectRatio="xMidYMid slice"
	aria-hidden="true"
>
	<defs>
		<radialGradient id="topo-fade" cx="50%" cy="35%" r="80%">
			<stop offset="0%"   stop-color="white" stop-opacity="1" />
			<stop offset="70%"  stop-color="white" stop-opacity="0.55" />
			<stop offset="100%" stop-color="white" stop-opacity="0.1" />
		</radialGradient>
		<mask id="topo-mask">
			<rect width="100%" height="100%" fill="url(#topo-fade)" />
		</mask>
	</defs>
	<g mask="url(#topo-mask)" stroke={color} fill="none" stroke-linecap="round">
		{#each paths as { d, idx }}
			{@const major = idx % 5 === 0}
			<path
				{d}
				stroke-width={major ? 1.1 : 0.6}
				opacity={major ? opacity * 1.6 : opacity}
			/>
		{/each}
	</g>
</svg>

<style>
	svg {
		position: fixed;
		inset: 0;
		width: 100%;
		height: 100%;
		pointer-events: none;
		z-index: 0;
	}
</style>
