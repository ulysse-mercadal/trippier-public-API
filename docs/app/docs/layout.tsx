import type { ReactNode } from 'react';
import { DocsLayout } from 'fumadocs-ui/layouts/docs';
import { source } from '@/lib/source';

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <DocsLayout
      tree={source.pageTree}
      nav={{
        title: (
          <span className="text-sm font-semibold tracking-tight">
            tripp<em className="text-[#10b981] not-italic">ier</em>{' '}
            <span className="text-[#6b7280] font-normal">API</span>
          </span>
        ),
        url: '/docs',
      }}
      sidebar={{
        banner: (
          <div className="mx-2 mb-2 rounded-lg border border-[#1c1c1c] bg-[#0a0a0a] px-3 py-2 text-xs text-[#6b7280]">
            v1.0 — POI &amp; Itinerary APIs
          </div>
        ),
      }}
    >
      {children}
    </DocsLayout>
  );
}
