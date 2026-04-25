import { NextResponse } from 'next/server';
import { internalAuth, POI_URL } from '@/lib/internal-auth';

export async function GET() {
  try {
    const res = await fetch(`${POI_URL}/pois/providers`, {
      headers: { 'X-Internal-Auth': internalAuth() },
    });
    const body = await res.text();
    return new NextResponse(body, {
      status: res.status,
      headers: { 'content-type': res.headers.get('content-type') ?? 'application/json' },
    });
  } catch {
    return NextResponse.json({ error: 'poi-api unreachable' }, { status: 503 });
  }
}
