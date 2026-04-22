"""Application configuration loaded from environment variables."""

from __future__ import annotations

from functools import lru_cache

from pydantic_settings import BaseSettings

from app.services.itinerary_service import ItineraryService
from app.services.poi_client import PoiClient


class Settings(BaseSettings):
    """Runtime settings for the itinerary-api server."""

    poi_api_url: str = "http://localhost:8080"
    poi_client_timeout: float = 10.0
    log_level: str = "info"

    model_config = {"env_file": ".env", "env_file_encoding": "utf-8"}


@lru_cache
def get_settings() -> Settings:
    return Settings()


def get_poi_client() -> PoiClient:
    s = get_settings()
    return PoiClient(base_url=s.poi_api_url, timeout=s.poi_client_timeout)


def get_itinerary_service() -> ItineraryService:
    return ItineraryService()
