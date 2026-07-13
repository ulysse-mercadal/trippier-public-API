/** Roadmap tag definitions with display label and color hue. */
export const TAGS = {
	routes: { label: 'Routes',      hue: 152 },
	sdk:    { label: 'SDK',         hue: 200 },
	cloud:  { label: 'Cloud',       hue: 230 },
	self:   { label: 'Self-hosted', hue: 280 },
	dx:     { label: 'DX',          hue: 50  },
	perf:   { label: 'Perf',        hue: 25  },
	data:   { label: 'Data',        hue: 320 },
} as const;

/** Identifier of a roadmap tag, derived from the TAGS keys. */
export type TagId = keyof typeof TAGS;
/** Identifier of a roadmap column. */
export type ColumnId = 'shipped' | 'progress' | 'next' | 'later';

/** A single roadmap entry within a column. */
export interface RoadmapItem {
	id:     string;
	title:  string;
	tag:    TagId;
	meta?:  string;
	issue:  number;
	votes?: number;
}

/** A roadmap column grouping related items. */
export interface RoadmapColumn {
	id:    ColumnId;
	label: string;
	sub:   string;
	items: RoadmapItem[];
}

/** Full roadmap dataset made of columns. */
export interface RoadmapData {
	columns: RoadmapColumn[];
}
