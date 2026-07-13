"""Async HTTP client for the poi-api service."""

from __future__ import annotations

import hashlib
import hmac
import time

import httpx

from app.models.types import Poi, PoiQuery


def _build_internal_auth(secret: str) -> str:
    """Build the HMAC-signed internal auth header value.

    Args:
        secret: shared secret used to sign the timestamp.

    Returns:
        The timestamp and signature joined as "ts.sig".
    """
    ts = str(int(time.time()))
    sig = hmac.new(secret.encode(), ts.encode(), hashlib.sha256).hexdigest()
    return f"{ts}.{sig}"


class PoiClient:
    """Fetches POIs from the poi-api service. Reuses a single AsyncClient across calls."""

    def __init__(self, base_url: str, timeout: float = 10.0, internal_secret: str = "") -> None:
        """Initialize the client and its underlying async HTTP session.

        Args:
            base_url: base URL of the poi-api service.
            timeout: request timeout in seconds.
            internal_secret: shared secret for internal auth signing.
        """
        self._base_url = base_url.rstrip("/")
        self._internal_secret = internal_secret
        self._client = httpx.AsyncClient(timeout=timeout)

    async def search(self, query: PoiQuery) -> list[Poi]:
        """Search POIs matching the given query via the poi-api service.

        Args:
            query: search filters and parameters.

        Returns:
            The list of matching POIs.
        """
        params = query.model_dump(exclude_none=True)
        if params.get("types"):
            params["types"] = ",".join(params["types"])
        headers = (
            {"X-Internal-Auth": _build_internal_auth(self._internal_secret)}
            if self._internal_secret
            else {}
        )
        response = await self._client.get(
            f"{self._base_url}/v1/pois/search",
            params=params,
            headers=headers,
        )
        response.raise_for_status()
        data = response.json()
        return [Poi(**item) for item in data.get("results", [])]

    async def aclose(self) -> None:
        """Close the underlying async HTTP client."""
        await self._client.aclose()
