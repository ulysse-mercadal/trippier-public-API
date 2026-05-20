"""Shared Pydantic models for the itinerary-api request and response contracts."""

from __future__ import annotations

from enum import StrEnum

from pydantic import BaseModel, Field


class PoiType(StrEnum):
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
    coords: Coordinates | None = None
    description: str | None = None
    distance: float | None = None


class Preferences(BaseModel):
    """Caller-defined constraints for the generated itinerary."""

    pace: str = Field(default="moderate", pattern="^(relaxed|moderate|intensive)$")
    priorities: list[PoiType] = Field(default_factory=list)
    avoid: list[PoiType] = Field(default_factory=list)
    start_time: str = Field(default="09:00", pattern=r"^\d{2}:\d{2}$")
    end_time: str = Field(default="21:00", pattern=r"^\d{2}:\d{2}$")


class PoiQuery(BaseModel):
    """Validated query forwarded to poi-api /pois/search — mirrors poi-api SearchQuery."""

    mode: str = Field(default="radius", pattern="^(radius|polygon|district)$")
    lat: float | None = None
    lng: float | None = None
    radius: int | None = Field(default=None, ge=0, le=50_000)
    polygon: str | None = None
    district: str | None = None
    types: list[PoiType] | None = None
    lang: str | None = None
    limit: int | None = Field(default=None, ge=1, le=400)
    offset: int | None = Field(default=None, ge=0)
    min_score: float | None = None


class ItineraryRequest(BaseModel):
    """Request body for POST /itinerary/generate."""

    pois: list[Poi] | None = None
    poi_query: PoiQuery | None = Field(
        default=None,
        description="Typed query forwarded to poi-api if pois is not provided.",
    )
    days: int = Field(default=1, ge=1, le=30)
    start_location: Coordinates | None = None
    preferences: Preferences = Field(default_factory=Preferences)


class DayPlan(BaseModel):
    """A single day within the generated itinerary."""

    day: int
    pois: list[Poi]
    estimated_duration_hours: float
    description: str | None = None


class ItineraryResponse(BaseModel):
    """Response body for POST /itinerary/generate."""

    days: list[DayPlan]
    total_pois: int
    summary: str | None = None
