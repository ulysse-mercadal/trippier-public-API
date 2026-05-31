"""Async HTTP client for the poi-api service."""

from __future__ import annotations

import hashlib
import hmac
import time

import httpx

from app.models.types import Poi, PoiQuery


def _build_internal_auth(secret: str) -> str:
    ts = str(int(time.time()))
    sig = hmac.new(secret.encode(), ts.encode(), hashlib.sha256).hexdigest()
    return f"{ts}.{sig}"


class PoiClient:
    """Fetches POIs from the poi-api service. Reuses a single AsyncClient across calls."""

    def __init__(self, base_url: str, timeout: float = 10.0, internal_secret: str = "") -> None:
        self._base_url = base_url.rstrip("/")
        self._internal_secret = internal_secret
        self._client = httpx.AsyncClient(timeout=timeout)

    async def search(self, query: PoiQuery) -> list[Poi]:
        params = query.model_dump(exclude_none=True)
        if params.get("types"):
            params["types"] = ",".join(params["types"])
        headers = {"X-Internal-Auth": _build_internal_auth(self._internal_secret)} if self._internal_secret else {}
        response = await self._client.get(f"{self._base_url}/pois/search", params=params, headers=headers)
        response.raise_for_status()
        data = response.json()
        return [Poi(**item) for item in data.get("results", [])]

    async def aclose(self) -> None:
        await self._client.aclose()
