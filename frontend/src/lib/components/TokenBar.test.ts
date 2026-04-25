import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import TokenBar from './TokenBar.svelte';

describe('TokenBar', () => {
	it('displays remaining and limit', () => {
		render(TokenBar, { remaining: 750, limit: 1000, resetsInSecs: 3600 });
		expect(screen.getByText(/750/)).toBeTruthy();
		expect(screen.getByText(/1000 tokens/)).toBeTruthy();
	});

	it('formats reset time in hours when > 60 minutes', () => {
		render(TokenBar, { remaining: 500, limit: 1000, resetsInSecs: 7200 }); // 2h
		expect(screen.getByText(/2h/)).toBeTruthy();
	});

	it('formats reset time in minutes when <= 60 minutes', () => {
		render(TokenBar, { remaining: 500, limit: 1000, resetsInSecs: 1800 }); // 30m
		expect(screen.getByText(/30m/)).toBeTruthy();
	});

	it('rounds up partial minutes', () => {
		render(TokenBar, { remaining: 500, limit: 1000, resetsInSecs: 90 }); // 1.5m → 2m
		expect(screen.getByText(/2m/)).toBeTruthy();
	});

	it('applies ok variant when remaining > 25%', () => {
		const { container } = render(TokenBar, { remaining: 800, limit: 1000, resetsInSecs: 3600 });
		expect(container.querySelector('.fill-ok')).toBeTruthy();
	});

	it('applies warn variant when remaining is between 10% and 25%', () => {
		const { container } = render(TokenBar, { remaining: 150, limit: 1000, resetsInSecs: 3600 });
		expect(container.querySelector('.fill-warn')).toBeTruthy();
	});

	it('applies empty variant when remaining <= 10%', () => {
		const { container } = render(TokenBar, { remaining: 50, limit: 1000, resetsInSecs: 3600 });
		expect(container.querySelector('.fill-empty')).toBeTruthy();
	});

	it('handles zero limit without division by zero', () => {
		// pct = 0 when limit === 0 → fill-empty
		const { container } = render(TokenBar, { remaining: 0, limit: 0, resetsInSecs: 3600 });
		expect(container.querySelector('.fill-empty')).toBeTruthy();
	});
});
