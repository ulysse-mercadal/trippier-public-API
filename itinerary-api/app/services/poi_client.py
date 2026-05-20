"""Async HTTP client for the poi-api service."""

from __future__ import annotations

import httpx

from app.models.types import Poi, PoiQuery


class PoiClient:
    """Fetches POIs from the poi-api service. Reuses a single AsyncClient across calls."""

    def __init__(self, base_url: str, timeout: float = 10.0) -> None:
        self._base_url = base_url.rstrip("/")
        self._client = httpx.AsyncClient(timeout=timeout)

    async def search(self, query: PoiQuery) -> list[Poi]:
        params = query.model_dump(exclude_none=True)
        if params.get("types"):
            params["types"] = ",".join(params["types"])
        response = await self._client.get(f"{self._base_url}/pois/search", params=params)
        response.raise_for_status()
        data = response.json()
        return [Poi(**item) for item in data.get("results", [])]

    async def aclose(self) -> None:
        await self._client.aclose()
