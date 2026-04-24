"""HTTP router for itinerary generation endpoints."""

from __future__ import annotations

from fastapi import APIRouter, Depends, HTTPException, status

from app.config import get_itinerary_service, get_poi_client
from app.models.types import ItineraryRequest, ItineraryResponse
from app.services.itinerary_service import ItineraryService
from app.services.poi_client import PoiClient

router = APIRouter(prefix="/itinerary", tags=["itinerary"])


@router.post(
    "/generate",
    response_model=ItineraryResponse,
    summary="Generate a day-by-day itinerary from a set of POIs",
)
async def generate(
    request: ItineraryRequest,
    poi_client: PoiClient = Depends(get_poi_client),  # noqa: B008
    service: ItineraryService = Depends(get_itinerary_service),  # noqa: B008
) -> ItineraryResponse:
    if request.pois is None and request.poi_query is None:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail="Provide either 'pois' or 'poi_query'.",
        )

    pois = request.pois or []
    if not pois and request.poi_query:
        pois = await poi_client.search(request.poi_query)

    if not pois:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="No POIs found for the given query.",
        )

    return service.generate(request, pois)
