export const TAGS = {
	routes: { label: 'Routes',      hue: 152 },
	sdk:    { label: 'SDK',         hue: 200 },
	cloud:  { label: 'Cloud',       hue: 230 },
	self:   { label: 'Self-hosted', hue: 280 },
	dx:     { label: 'DX',          hue: 50  },
	perf:   { label: 'Perf',        hue: 25  },
	data:   { label: 'Data',        hue: 320 },
} as const;

export type TagId = keyof typeof TAGS;
export type ColumnId = 'shipped' | 'progress' | 'next' | 'later';

export interface RoadmapItem {
	id:     string;
	title:  string;
	tag:    TagId;
	meta?:  string;
	issue:  number;
	votes?: number;
}

export interface RoadmapColumn {
	id:    ColumnId;
	label: string;
	sub:   string;
	items: RoadmapItem[];
}

export interface RoadmapData {
	columns: RoadmapColumn[];
}
