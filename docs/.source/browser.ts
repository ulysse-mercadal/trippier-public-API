// @ts-nocheck
import { browser } from 'fumadocs-mdx/runtime/browser';
import type * as Config from '../source.config';

const create = browser<typeof Config, import("fumadocs-mdx/runtime/types").InternalTypeConfig & {
  DocData: {
  }
}>();
const browserCollections = {
  docs: create.doc("docs", {"algorithm.mdx": () => import("../content/docs/algorithm.mdx?collection=docs"), "index.mdx": () => import("../content/docs/index.mdx?collection=docs"), "itinerary-api/generate.mdx": () => import("../content/docs/itinerary-api/generate.mdx?collection=docs"), "poi-api/events.mdx": () => import("../content/docs/poi-api/events.mdx?collection=docs"), "poi-api/health.mdx": () => import("../content/docs/poi-api/health.mdx?collection=docs"), "poi-api/providers.mdx": () => import("../content/docs/poi-api/providers.mdx?collection=docs"), "poi-api/search-slim.mdx": () => import("../content/docs/poi-api/search-slim.mdx?collection=docs"), "poi-api/search.mdx": () => import("../content/docs/poi-api/search.mdx?collection=docs"), }),
};
export default browserCollections;