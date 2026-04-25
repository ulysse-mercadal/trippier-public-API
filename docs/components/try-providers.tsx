'use client';

import { TryBase } from './try-base';

export function TryProviders() {
  return (
    <TryBase
      method="GET"
      endpointLabel="/pois/providers"
      fields={[]}
      buildUrl={() => 'GET /pois/providers'}
      fetchPath={() => '/api/proxy/providers'}
    />
  );
}
