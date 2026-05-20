import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { RoadmapData } from '$lib/data/roadmap';

export const load: PageServerLoad = async ({ fetch }) => {
	const res = await fetch('/api/roadmap');
	if (!res.ok) throw error(500, 'Failed to load roadmap');
	const data: RoadmapData = await res.json();
	return { roadmap: data };
};
