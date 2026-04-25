'use client';

import { useState } from 'react';

const defaultBody = JSON.stringify(
  {
    poi_query: { lat: 48.8566, lng: 2.3522, radius: 3000, mode: 'radius' },
    days: 2,
    start_location: { lat: 48.8566, lng: 2.3522 },
    preferences: {
      pace: 'moderate',
      priorities: ['see', 'eat'],
      start_time: '09:00',
      end_time: '21:00',
    },
  },
  null,
  2
);

function statusClass(s: number) {
  if (s >= 200 && s < 300) return 'ok';
  if (s >= 400) return 'err';
  return '';
}

export function TryItinerary() {
  const [body, setBody] = useState(defaultBody);
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState(0);
  const [result, setResult] = useState('');

  async function run() {
    setLoading(true);
    setResult('');
    setStatus(0);
    try {
      const res = await fetch('/api/proxy/itinerary', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body,
      });
      setStatus(res.status);
      const json = await res.json();
      setResult(JSON.stringify(json, null, 2));
    } catch (e) {
      setResult(String(e));
    } finally {
      setLoading(false);
    }
  }

  const sc = statusClass(status);

  return (
    <div className="try-block">
      <div className="try-header">
        <span className="method-badge method-post">POST</span>
        <code className="try-endpoint">/itinerary/generate</code>
        <span className="try-label">Try it</span>
      </div>

      <label className="body-label">
        <span>Request body</span>
        <textarea
          value={body}
          onChange={(e) => setBody(e.target.value)}
          rows={14}
          spellCheck={false}
        />
      </label>

      <div className="try-req-row">
        <code className="try-req">POST /itinerary/generate</code>
        <button className="try-btn" disabled={loading} onClick={run}>
          {loading ? 'Loading…' : 'Send'}
        </button>
      </div>

      {result && (
        <div className="try-response">
          <div className="try-response-header">
            <span className={`status-badge ${sc}`}>{status}</span>
            <span className="status-label">{sc === 'ok' ? 'OK' : 'Error'}</span>
          </div>
          <pre className="try-result">{result}</pre>
        </div>
      )}

      <style jsx>{`
        .try-block {
          margin: 1.5rem 0;
          padding: 1.1rem 1.25rem;
          background: #0a0a0a;
          border: 1px solid #1c1c1c;
          border-radius: 10px;
        }
        .try-header {
          display: flex;
          align-items: center;
          gap: 0.6rem;
          margin-bottom: 1rem;
          flex-wrap: wrap;
        }
        .try-endpoint {
          font-family: 'SF Mono', 'Fira Code', monospace;
          font-size: 0.85rem;
          color: #e5e5e5;
          flex: 1;
        }
        .try-label {
          font-size: 0.68rem;
          font-weight: 700;
          text-transform: uppercase;
          letter-spacing: 0.1em;
          color: #10b981;
          background: rgba(16, 185, 129, 0.1);
          border: 1px solid rgba(16, 185, 129, 0.25);
          border-radius: 999px;
          padding: 0.2rem 0.6rem;
          margin-left: auto;
        }
        .body-label {
          display: flex;
          flex-direction: column;
          gap: 0.3rem;
          margin-bottom: 0.75rem;
        }
        .body-label span {
          font-size: 0.72rem;
          font-weight: 600;
          text-transform: uppercase;
          letter-spacing: 0.06em;
          color: #6b7280;
        }
        .body-label textarea {
          background: #050505;
          border: 1px solid #1c1c1c;
          border-radius: 6px;
          color: #e5e5e5;
          font-size: 0.8rem;
          font-family: 'SF Mono', 'Fira Code', monospace;
          line-height: 1.6;
          padding: 0.65rem 0.75rem;
          resize: vertical;
          transition: border-color 0.15s;
          width: 100%;
        }
        .body-label textarea:focus {
          outline: none;
          border-color: #10b981;
        }
        .try-req-row {
          display: flex;
          align-items: center;
          gap: 0.75rem;
        }
        .try-req {
          flex: 1;
          font-family: 'SF Mono', 'Fira Code', monospace;
          font-size: 0.78rem;
          color: #6b7280;
          background: #050505;
          border: 1px solid #1c1c1c;
          border-radius: 6px;
          padding: 0.5rem 0.75rem;
          display: block;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
        .try-btn {
          padding: 0.5rem 1.2rem;
          background: #10b981;
          color: #000;
          border: none;
          border-radius: 6px;
          font-size: 0.85rem;
          font-weight: 600;
          cursor: pointer;
          white-space: nowrap;
          transition: background 0.15s;
          flex-shrink: 0;
        }
        .try-btn:hover:not(:disabled) { background: #059669; }
        .try-btn:disabled { opacity: 0.6; cursor: not-allowed; }
        .try-response {
          margin-top: 0.9rem;
          border: 1px solid #1c1c1c;
          border-radius: 8px;
          overflow: hidden;
        }
        .try-response-header {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          padding: 0.45rem 0.75rem;
          background: #0f0f0f;
          border-bottom: 1px solid #1c1c1c;
        }
        .status-badge {
          font-family: 'SF Mono', 'Fira Code', monospace;
          font-size: 0.75rem;
          font-weight: 700;
          padding: 0.15rem 0.5rem;
          border-radius: 4px;
          border: 1px solid;
        }
        .status-badge.ok { background: #0d2e18; color: #4ade80; border-color: #166534; }
        .status-badge.err { background: #2d0d0d; color: #f87171; border-color: #991b1b; }
        .status-label { font-size: 0.75rem; color: #6b7280; }
        .try-result {
          background: #050505;
          color: #e5e5e5;
          font-family: 'SF Mono', 'Fira Code', monospace;
          font-size: 0.78rem;
          line-height: 1.65;
          padding: 0.9rem 1rem;
          margin: 0;
          max-height: 380px;
          overflow: auto;
          white-space: pre;
        }
      `}</style>
    </div>
  );
}
