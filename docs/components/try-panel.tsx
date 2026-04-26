'use client';

import { useState, type ReactNode } from 'react';
import { statusClass } from '@/lib/status';

interface TryPanelProps {
  method?: 'GET' | 'POST';
  endpointLabel: string;
  fetchPath: () => string;
  urlPreview: string;
  bodyMethod?: boolean;
  buildBody?: () => string;
  children?: ReactNode;
}

export function TryPanel({
  method = 'GET',
  endpointLabel,
  fetchPath,
  urlPreview,
  bodyMethod = false,
  buildBody,
  children,
}: TryPanelProps) {
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState(0);
  const [result, setResult] = useState('');

  async function run() {
    setLoading(true);
    setResult('');
    setStatus(0);
    try {
      const opts: RequestInit = bodyMethod
        ? {
            method: 'POST',
            headers: { 'content-type': 'application/json' },
            body: buildBody ? buildBody() : undefined,
          }
        : { method: 'GET' };
      const res = await fetch(fetchPath(), opts);
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
      {/* Header */}
      <div className="try-header">
        <span
          className="method-badge"
          style={{
            background: method === 'POST' ? '#0d1f3d' : '#0d2e18',
            color: method === 'POST' ? '#60a5fa' : '#4ade80',
            borderColor: method === 'POST' ? '#1d4ed8' : '#166534',
          }}
        >
          {method}
        </span>
        <code className="try-endpoint">{endpointLabel}</code>
        <span className="try-label">Try it</span>
      </div>

      {/* Fields (children) */}
      {children}

      {/* Request preview */}
      <div className="try-req-row">
        <code className="try-req">{urlPreview}</code>
        <button className="try-btn" disabled={loading} onClick={run}>
          {loading ? 'Loading…' : 'Send'}
        </button>
      </div>

      {/* Response */}
      {result && (
        <div className="try-response">
          <div className="try-response-header">
            <span className={`status-badge ${sc}`}>{status}</span>
            <span className="status-label">
              {sc === 'ok' ? 'OK' : sc === 'err' ? 'Error' : ''}
            </span>
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
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
        }

        .try-header {
          display: flex;
          align-items: center;
          gap: 0.6rem;
          margin-bottom: 1rem;
          flex-wrap: wrap;
        }

        .method-badge {
          font-family: 'SF Mono', 'Fira Code', monospace;
          font-size: 0.72rem;
          font-weight: 700;
          padding: 0.2rem 0.55rem;
          border-radius: 4px;
          border: 1px solid;
          letter-spacing: 0.05em;
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

        .try-req-row {
          display: flex;
          align-items: center;
          gap: 0.75rem;
          margin-top: 0.75rem;
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
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
          display: block;
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

        .try-btn:hover:not(:disabled) {
          background: #059669;
        }

        .try-btn:disabled {
          opacity: 0.6;
          cursor: not-allowed;
        }

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

        .status-badge.ok {
          background: #0d2e18;
          color: #4ade80;
          border-color: #166534;
        }

        .status-badge.err {
          background: #2d0d0d;
          color: #f87171;
          border-color: #991b1b;
        }

        .status-label {
          font-size: 0.75rem;
          color: #6b7280;
        }

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
