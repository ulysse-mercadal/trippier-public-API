"""FastAPI application factory for the itinerary-api service."""

from fastapi import FastAPI

from app.routers.itinerary import router as itinerary_router

app = FastAPI(
    title="Itinerary API",
    description="Generates day-by-day travel itineraries from POI collections.",
    version="0.1.0",
)

app.include_router(itinerary_router)


@app.get("/health", tags=["meta"])
def health() -> dict:
    return {"status": "ok"}
