import { loader } from 'fumadocs-core/source';
import { toFumadocsSource } from 'fumadocs-mdx/runtime/server';
import { docs } from '@/.source/server';

export const source = loader({
  baseUrl: '/docs',
  source: toFumadocsSource(docs.docs, docs.meta),
});
