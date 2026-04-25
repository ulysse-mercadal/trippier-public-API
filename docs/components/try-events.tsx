'use client';

import { useState } from 'react';
import { TryPanel } from './try-panel';

export function TryEvents() {
  const [mode, setMode] = useState('radius');
  const [district, setDistrict] = useState('Paris');
  const [radius, setRadius] = useState('5000');
  const [limit, setLimit] = useState('10');

  const LAT = '48.8566';
  const LNG = '2.3522';

  function buildQs() {
    if (mode === 'district') {
      return new URLSearchParams({ mode, district, limit }).toString();
    }
    return new URLSearchParams({ mode, lat: LAT, lng: LNG, radius, limit }).toString();
  }

  return (
    <TryPanel
      method="GET"
      endpointLabel="/pois/events"
      fetchPath={() => `/api/proxy/events?${buildQs()}`}
      urlPreview={`GET /pois/events?${buildQs()}`}
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
          <span>limit</span>
          <input type="text" value={limit} onChange={(e) => setLimit(e.target.value)} />
        </label>
      </div>
    </TryPanel>
  );
}
