"""Rate-limit middleware: delegates token deduction to auth-api."""

from __future__ import annotations

import httpx
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request
from starlette.responses import JSONResponse, Response


EXEMPT_PATHS = {"/health"}


class RateLimitMiddleware(BaseHTTPMiddleware):
    """Calls auth-api /internal/check-rate-limit before each protected request."""

    def __init__(self, app, auth_api_url: str, internal_secret: str, cost: int) -> None:
        super().__init__(app)
        self._auth_api_url = auth_api_url.rstrip("/")
        self._internal_secret = internal_secret
        self._cost = cost
        self._client = httpx.AsyncClient(timeout=5.0)

    async def dispatch(self, request: Request, call_next) -> Response:
        if request.url.path in EXEMPT_PATHS:
            return await call_next(request)

        api_key = request.headers.get("X-API-Key")
        if not api_key:
            return JSONResponse({"error": "X-API-Key header required"}, status_code=401)

        try:
            resp = await self._client.post(
                f"{self._auth_api_url}/internal/check-rate-limit",
                json={"api_key": api_key, "cost": self._cost},
                headers={"X-Internal-Secret": self._internal_secret},
            )
            data = resp.json()
        except Exception:
            return JSONResponse({"error": "rate-limit check failed"}, status_code=503)

        if not data.get("allowed"):
            if data.get("error") == "invalid api key":
                return JSONResponse({"error": "invalid api key"}, status_code=401)
            return JSONResponse(
                {"error": "rate limit exceeded", "resets_in_secs": data.get("resets_in_secs", 0)},
                status_code=429,
            )

        response = await call_next(request)
        response.headers["X-RateLimit-Limit"] = str(data.get("limit", 0))
        response.headers["X-RateLimit-Remaining"] = str(data.get("remaining", 0))
        return response
