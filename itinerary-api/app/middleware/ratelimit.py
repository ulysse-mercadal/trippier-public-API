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
    """Build a signed X-Internal-Auth header value.

    Args:
        secret: shared secret used to sign the timestamp.

    Returns:
        Header value formatted as '<ts>.<hmac-sha256(secret, ts)>'.
    """
    ts = str(int(time.time()))
    sig = hmac.new(secret.encode(), ts.encode(), hashlib.sha256).hexdigest()
    return f"{ts}.{sig}"


class RateLimitMiddleware(BaseHTTPMiddleware):
    """Calls auth-api /internal/check-rate-limit before each protected request."""

    def __init__(self, app: ASGIApp, auth_api_url: str, internal_secret: str, cost: int) -> None:
        """Configure the middleware and its HTTP client.

        Args:
            app: wrapped ASGI application.
            auth_api_url: base URL of the auth-api service.
            internal_secret: shared secret for internal auth headers.
            cost: token cost charged per request.
        """
        super().__init__(app)
        self._auth_api_url = auth_api_url.rstrip("/")
        self._internal_secret = internal_secret
        self._cost = cost
        self._client = httpx.AsyncClient(timeout=5.0)

    @staticmethod
    def _validate_internal_auth(header: str, secret: str) -> bool:
        """Check whether an X-Internal-Auth header is valid and recent.

        Args:
            header: raw header value formatted as '<ts>.<sig>'.
            secret: shared secret used to verify the signature.

        Returns:
            True if the header is valid and within 30 seconds of now.
        """
        try:
            ts_str, sig = header.split(".", 1)
            ts = int(ts_str)
        except (ValueError, AttributeError):
            return False
        if abs(int(time.time()) - ts) > 30:
            return False
        expected = hmac.new(secret.encode(), ts_str.encode(), hashlib.sha256).hexdigest()
        return hmac.compare_digest(expected, sig)

    async def dispatch(self, request: Request, call_next: RequestResponseEndpoint) -> Response:
        """Enforce rate limiting via auth-api before letting the request through.

        Args:
            request: incoming HTTP request.
            call_next: handler that continues the middleware chain.

        Returns:
            The downstream response, or an error response if blocked.
        """
        if request.url.path in EXEMPT_PATHS:
            return await call_next(request)

        internal_auth = request.headers.get("X-Internal-Auth", "")
        if internal_auth and self._validate_internal_auth(internal_auth, self._internal_secret):
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
        except (httpx.RequestError, httpx.HTTPStatusError, ValueError):
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
