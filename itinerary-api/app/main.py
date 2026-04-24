"""FastAPI application factory for the itinerary-api service."""

from fastapi import FastAPI

from app.config import get_settings
from app.middleware.ratelimit import RateLimitMiddleware
from app.routers.itinerary import router as itinerary_router

app = FastAPI(
    title="Itinerary API",
    description="Generates day-by-day travel itineraries from POI collections.",
    version="0.1.0",
)

_s = get_settings()
app.add_middleware(
    RateLimitMiddleware,
    auth_api_url=_s.auth_api_url,
    internal_secret=_s.internal_secret,
    cost=50,
)

app.include_router(itinerary_router)


@app.get("/health", tags=["meta"])
def health() -> dict[str, str]:
    return {"status": "ok"}
