// @ts-nocheck
import * as __fd_glob_10 from "../content/docs/poi-api/search.mdx?collection=docs"
import * as __fd_glob_9 from "../content/docs/poi-api/search-slim.mdx?collection=docs"
import * as __fd_glob_8 from "../content/docs/poi-api/providers.mdx?collection=docs"
import * as __fd_glob_7 from "../content/docs/poi-api/health.mdx?collection=docs"
import * as __fd_glob_6 from "../content/docs/poi-api/events.mdx?collection=docs"
import * as __fd_glob_5 from "../content/docs/itinerary-api/generate.mdx?collection=docs"
import * as __fd_glob_4 from "../content/docs/index.mdx?collection=docs"
import * as __fd_glob_3 from "../content/docs/algorithm.mdx?collection=docs"
import { default as __fd_glob_2 } from "../content/docs/poi-api/meta.json?collection=docs"
import { default as __fd_glob_1 } from "../content/docs/itinerary-api/meta.json?collection=docs"
import { default as __fd_glob_0 } from "../content/docs/meta.json?collection=docs"
import { server } from 'fumadocs-mdx/runtime/server';
import type * as Config from '../source.config';

const create = server<typeof Config, import("fumadocs-mdx/runtime/types").InternalTypeConfig & {
  DocData: {
  }
}>({"doc":{"passthroughs":["extractedReferences"]}});

export const docs = await create.docs("docs", "content/docs", {"meta.json": __fd_glob_0, "itinerary-api/meta.json": __fd_glob_1, "poi-api/meta.json": __fd_glob_2, }, {"algorithm.mdx": __fd_glob_3, "index.mdx": __fd_glob_4, "itinerary-api/generate.mdx": __fd_glob_5, "poi-api/events.mdx": __fd_glob_6, "poi-api/health.mdx": __fd_glob_7, "poi-api/providers.mdx": __fd_glob_8, "poi-api/search-slim.mdx": __fd_glob_9, "poi-api/search.mdx": __fd_glob_10, });