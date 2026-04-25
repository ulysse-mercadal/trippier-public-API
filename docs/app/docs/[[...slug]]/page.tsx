import { source } from '@/lib/source';
import {
  DocsPage,
  DocsBody,
  DocsTitle,
  DocsDescription,
} from 'fumadocs-ui/page';
import { notFound } from 'next/navigation';
import defaultMdxComponents from 'fumadocs-ui/mdx';
import type { Metadata } from 'next';

// Custom MDX components registered globally for all doc pages
import { TryPois } from '@/components/try-pois';
import { TryPoisSlim } from '@/components/try-pois-slim';
import { TryEvents } from '@/components/try-events';
import { TryProviders } from '@/components/try-providers';
import { TryItinerary } from '@/components/try-itinerary';

const customComponents = {
  TryPois,
  TryPoisSlim,
  TryEvents,
  TryProviders,
  TryItinerary,
};

export default async function Page({
  params,
}: {
  params: Promise<{ slug?: string[] }>;
}) {
  const { slug } = await params;
  const page = source.getPage(slug);
  if (!page) notFound();

  const MDX = page.data.body;

  return (
    <DocsPage toc={page.data.toc} full={page.data.full}>
      <DocsTitle>{page.data.title}</DocsTitle>
      <DocsDescription>{page.data.description}</DocsDescription>
      <DocsBody>
        <MDX components={{ ...defaultMdxComponents, ...customComponents }} />
      </DocsBody>
    </DocsPage>
  );
}

export async function generateStaticParams() {
  return source.generateParams();
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug?: string[] }>;
}): Promise<Metadata> {
  const { slug } = await params;
  const page = source.getPage(slug);
  if (!page) return {};
  return {
    title: `${page.data.title} — trippier API`,
    description: page.data.description,
  };
}
