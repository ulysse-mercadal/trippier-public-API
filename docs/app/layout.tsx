import type { ReactNode } from 'react';
import { RootProvider } from 'fumadocs-ui/provider/next';
import './globals.css';

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" className="dark" suppressHydrationWarning>
      <body className="antialiased">
        <RootProvider>{children}</RootProvider>
      </body>
    </html>
  );
}
