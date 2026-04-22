"""Async HTTP client for the poi-api service."""

from __future__ import annotations

import httpx

from app.models.types import Poi


class PoiClient:
    """Fetches POIs from the poi-api service."""

    def __init__(self, base_url: str, timeout: float = 10.0) -> None:
        self._base_url = base_url.rstrip("/")
        self._timeout = timeout

    async def search(self, query: dict) -> list[Poi]:
        async with httpx.AsyncClient(timeout=self._timeout) as client:
            response = await client.get(f"{self._base_url}/pois/search", params=query)
            response.raise_for_status()
            data = response.json()
            return [Poi(**item) for item in data.get("results", [])]
