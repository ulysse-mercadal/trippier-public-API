import { type NextRequest, NextResponse } from 'next/server';
import { internalAuth, ITIN_URL } from '@/lib/internal-auth';

export async function POST(req: NextRequest) {
  try {
    const body = await req.text();
    if (body.length > 100_000) {
      return NextResponse.json({ error: 'request too large' }, { status: 413 });
    }
    const res = await fetch(`${ITIN_URL}/itinerary/generate`, {
      method: 'POST',
      signal: AbortSignal.timeout(60_000),
      headers: {
        'content-type': 'application/json',
        'X-Internal-Auth': internalAuth(),
      },
      body,
    });
    const resBody = await res.text();
    return new NextResponse(resBody, {
      status: res.status,
      headers: { 'content-type': res.headers.get('content-type') ?? 'application/json' },
    });
  } catch {
    return NextResponse.json({ error: 'itinerary-api unreachable' }, { status: 503 });
  }
}
