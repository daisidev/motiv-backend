# Location Data Update

This document describes the changes made to support Google Maps location data in the MOTIV backend.

## Database Changes

### New Fields Added to `events` Table

- `latitude` (DOUBLE PRECISION) - Latitude coordinate for event location
- `longitude` (DOUBLE PRECISION) - Longitude coordinate for event location  
- `place_id` (VARCHAR(255)) - Google Places API place ID for event location

### Migration

Run the migration SQL file to add the new fields:

```sql
-- See migrations/add_location_coordinates.sql
ALTER TABLE events 
ADD COLUMN latitude DOUBLE PRECISION,
ADD COLUMN longitude DOUBLE PRECISION,
ADD COLUMN place_id VARCHAR(255);

CREATE INDEX idx_events_coordinates ON events(latitude, longitude);
```

## API Changes

### Request Structure

The `CreateEventRequest` now supports optional location data:

```json
{
  "title": "My Event",
  "location": "Lagos, Nigeria",
  "locationData": {
    "address": "Victoria Island, Lagos, Nigeria",
    "coordinates": {
      "lat": 6.4281,
      "lng": 3.4219
    },
    "placeId": "ChIJzQhTmzKLOxARYGJkNzx3wAQ"
  }
}
```

### Response Structure

Event responses now include the location coordinates:

```json
{
  "id": "uuid",
  "title": "My Event",
  "location": "Lagos, Nigeria",
  "latitude": 6.4281,
  "longitude": 3.4219,
  "place_id": "ChIJzQhTmzKLOxARYGJkNzx3wAQ"
}
```

## Backward Compatibility

- Existing events without coordinates will have `null` values for the new fields
- The `location` field (string) is still required and used as fallback
- Frontend can check for coordinate availability before showing maps

## Usage

### Creating Events with Location Data

```json
POST /api/v1/hosts/me/events
{
  "title": "Beach Party",
  "location": "Tarkwa Bay Beach, Lagos",
  "locationData": {
    "address": "Tarkwa Bay Beach, Lagos Island, Lagos, Nigeria",
    "coordinates": {
      "lat": 6.4167,
      "lng": 3.4167
    }
  }
}
```

### Updating Events with Location Data

```json
PUT /api/v1/hosts/me/events/{id}
{
  "location": "New Location",
  "locationData": {
    "address": "New Address",
    "coordinates": {
      "lat": 6.5000,
      "lng": 3.5000
    }
  }
}
```

## Future Enhancements

- Location-based event search using coordinates
- Distance calculations for nearby events
- Geofencing for event check-ins
- Integration with mapping services for directions