'use client';

import { useState } from 'react';
import { TryPanel } from './try-panel';

export function TryPoisSlim() {
  const [mode, setMode] = useState('radius');
  const [district, setDistrict] = useState('Paris');
  const [radius, setRadius] = useState('2000');
  const [types, setTypes] = useState('');
  const [limit, setLimit] = useState('10');

  const LAT = '48.8566';
  const LNG = '2.3522';

  function buildQs() {
    if (mode === 'district') {
      const p = new URLSearchParams({ mode, district, limit });
      if (types) p.set('types', types);
      return p.toString();
    }
    const p = new URLSearchParams({ mode, lat: LAT, lng: LNG, radius, limit });
    if (types) p.set('types', types);
    return p.toString();
  }

  return (
    <TryPanel
      method="GET"
      endpointLabel="/pois/search/slim"
      fetchPath={() => `/api/proxy/pois-slim?${buildQs()}`}
      urlPreview={`GET /pois/search/slim?${buildQs()}`}
    >
      <div className="try-fields">
        <label className="try-field">
          <span>mode</span>
          <select value={mode} onChange={(e) => setMode(e.target.value)}>
            <option>radius</option>
            <option>district</option>
          </select>
        </label>

        {mode === 'radius' ? (
          <>
            <label className="try-field">
              <span>lat</span>
              <input type="text" value={LAT} disabled />
            </label>
            <label className="try-field">
              <span>lng</span>
              <input type="text" value={LNG} disabled />
            </label>
            <label className="try-field">
              <span>radius</span>
              <input type="text" value={radius} onChange={(e) => setRadius(e.target.value)} placeholder="metres" />
            </label>
          </>
        ) : (
          <label className="try-field">
            <span>district</span>
            <input type="text" value={district} onChange={(e) => setDistrict(e.target.value)} placeholder="Paris, Montmartre…" />
          </label>
        )}

        <label className="try-field">
          <span>types</span>
          <input type="text" value={types} onChange={(e) => setTypes(e.target.value)} placeholder="see,eat,do…" />
        </label>
        <label className="try-field">
          <span>limit</span>
          <input type="text" value={limit} onChange={(e) => setLimit(e.target.value)} />
        </label>
      </div>
    </TryPanel>
  );
}
