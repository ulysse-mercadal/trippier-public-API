import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { RoadmapData } from '$lib/data/roadmap';

/**
 * Loads roadmap data for the roadmap page.
 * @param event server load event
 * @param event.fetch fetch function for the request
 * @returns roadmap data for the page
 */
export const load: PageServerLoad = async ({ fetch }) => {
	const res = await fetch('/api/roadmap');
	if (!res.ok) throw error(500, 'Failed to load roadmap');
	const data: RoadmapData = await res.json();
	return { roadmap: data };
};
