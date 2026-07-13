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
    auth_api_url: str = "http://auth-api:8081"
    internal_secret: str = "change-me-internal-secret"
    auth_disabled: bool = False

    model_config = {"env_file": ".env", "env_file_encoding": "utf-8"}


@lru_cache
def get_settings() -> Settings:
    """Build and cache the application settings instance.

    Returns:
        The cached Settings instance.
    """
    return Settings()


def get_poi_client() -> PoiClient:
    """Create a POI client configured from the current settings.

    Returns:
        A configured PoiClient instance.
    """
    s = get_settings()
    return PoiClient(
        base_url=s.poi_api_url,
        timeout=s.poi_client_timeout,
        internal_secret=s.internal_secret,
    )


def get_itinerary_service() -> ItineraryService:
    """Create a new itinerary service instance.

    Returns:
        A new ItineraryService instance.
    """
    return ItineraryService()
