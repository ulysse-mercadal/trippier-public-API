"""Shared Pydantic models for the itinerary-api request and response contracts."""

from __future__ import annotations

from enum import Enum
from typing import Optional

from pydantic import BaseModel, Field


class PoiType(str, Enum):
    """Category of a point of interest, mirroring the poi-api taxonomy."""

    see = "see"
    eat = "eat"
    drink = "drink"
    do = "do"
    buy = "buy"
    sleep = "sleep"
    generic = "generic"


class Coordinates(BaseModel):
    """Geographic position of a POI."""

    lat: float
    lng: float
    approximate: bool = False


class Poi(BaseModel):
    """Minimal POI representation consumed by the itinerary builder."""

    id: str
    name: str
    type: PoiType
    coords: Optional[Coordinates] = None
    description: Optional[str] = None
    distance: Optional[float] = None


class Preferences(BaseModel):
    """Caller-defined constraints for the generated itinerary."""

    pace: str = Field(default="moderate", pattern="^(relaxed|moderate|intensive)$")
    priorities: list[PoiType] = Field(default_factory=list)
    avoid: list[PoiType] = Field(default_factory=list)
    start_time: str = Field(default="09:00", pattern=r"^\d{2}:\d{2}$")
    end_time: str = Field(default="21:00", pattern=r"^\d{2}:\d{2}$")


class ItineraryRequest(BaseModel):
    """Request body for POST /itinerary/generate."""

    pois: Optional[list[Poi]] = None
    poi_query: Optional[dict] = Field(
        default=None,
        description="Pass-through query forwarded to poi-api if pois is not provided.",
    )
    days: int = Field(default=1, ge=1, le=30)
    start_location: Optional[Coordinates] = None
    preferences: Preferences = Field(default_factory=Preferences)


class DayPlan(BaseModel):
    """A single day within the generated itinerary."""

    day: int
    pois: list[Poi]
    estimated_duration_hours: float
    description: Optional[str] = None


class ItineraryResponse(BaseModel):
    """Response body for POST /itinerary/generate."""

    days: list[DayPlan]
    total_pois: int
    summary: Optional[str] = None
