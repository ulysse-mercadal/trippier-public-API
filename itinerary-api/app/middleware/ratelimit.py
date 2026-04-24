"""Rate-limit middleware: delegates token deduction to auth-api."""

from __future__ import annotations

import hashlib
import hmac
import time

import httpx
from starlette.middleware.base import BaseHTTPMiddleware, RequestResponseEndpoint
from starlette.requests import Request
from starlette.responses import JSONResponse, Response
from starlette.types import ASGIApp

EXEMPT_PATHS = {"/health"}


def _build_internal_auth(secret: str) -> str:
    """Return an X-Internal-Auth header value: '<ts>.<hmac-sha256(secret, ts)>'.

    Using a timestamp-bound HMAC prevents replaying a captured header value.
    The receiving side (auth-api) rejects tokens older than 30 s.
    """
    ts = str(int(time.time()))
    sig = hmac.new(secret.encode(), ts.encode(), hashlib.sha256).hexdigest()
    return f"{ts}.{sig}"


class RateLimitMiddleware(BaseHTTPMiddleware):
    """Calls auth-api /internal/check-rate-limit before each protected request."""

    def __init__(self, app: ASGIApp, auth_api_url: str, internal_secret: str, cost: int) -> None:
        super().__init__(app)
        self._auth_api_url = auth_api_url.rstrip("/")
        self._internal_secret = internal_secret
        self._cost = cost
        self._client = httpx.AsyncClient(timeout=5.0)

    async def dispatch(self, request: Request, call_next: RequestResponseEndpoint) -> Response:
        if request.url.path in EXEMPT_PATHS:
            return await call_next(request)

        api_key = request.headers.get("X-API-Key")
        if not api_key:
            return JSONResponse({"error": "X-API-Key header required"}, status_code=401)

        try:
            resp = await self._client.post(
                f"{self._auth_api_url}/internal/check-rate-limit",
                json={"api_key": api_key, "cost": self._cost},
                headers={"X-Internal-Auth": _build_internal_auth(self._internal_secret)},
            )
            data = resp.json()
        except Exception:
            return JSONResponse({"error": "rate-limit check failed"}, status_code=503)

        if not data.get("allowed"):
            if data.get("error") == "invalid api key":
                return JSONResponse({"error": "invalid api key"}, status_code=401)
            resets_in = data.get("resets_in_secs", 0)
            return JSONResponse(
                {"error": "rate limit exceeded", "resets_in_secs": resets_in},
                status_code=429,
                headers={"Retry-After": str(resets_in)},
            )

        response = await call_next(request)
        response.headers["X-RateLimit-Limit"] = str(data.get("limit", 0))
        response.headers["X-RateLimit-Remaining"] = str(data.get("remaining", 0))
        return response
