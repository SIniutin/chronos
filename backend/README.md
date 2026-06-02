# history-app-backend

## Figma course normalization

After extracting Figma text nodes to `backend/seeds/figma_course_raw.json`, normalize them into an intermediate course shape:

```bash
python3 ../scripts/normalize_figma_course.py
```

The command writes `backend/seeds/normalized_course.json` with `sections -> units -> skills -> facts/cause_effect/results` and keeps uncertain text in `ambiguous_items`.

Interactive challenge types supported by the current pipeline:

- `map_point`: point-on-map task using `payload.center/zoom/tile_url_template` and `answers.lat/lng/radius_m`.
- `map_area`: approximate area task using `payload.center/zoom` and `answers.center/area_m2/center_radius_m/area_tolerance`.
- `match_photos`: photo-to-label matching task using `options.photos`, `options.labels`, and `answers.photo_id/label_id`.

User answer formats:

```json
{
  "map_point": {
    "lat": 59.93,
    "lng": 30.33
  },
  "map_area": {
    "center": { "lat": 59.93, "lng": 30.33 },
    "area_m2": 2500000000
  },
  "match_photos": [
    { "photo_id": "p1", "label_id": "l1" }
  ]
}
```

For `map_area`, the Flutter UI lets the learner draw an approximate contour, then submits only the calculated centroid and area. The backend also accepts the older `{ "points": [...] }` polygon payload for compatibility.

The normalizer extracts potential map points, map areas, and photo matches into `map_points`, `map_areas`, and `photo_matches` on each normalized skill. Coordinates come from a small local reviewable dictionary for places explicitly mentioned in the Figma text. Image URLs are placeholders unless already present in source data, so generated photo tasks are marked `draft` with a `needs_review` tag.

To append normalized themes that are not covered by the curated opening
sections, generate their base challenges, and merge normalized interactive
candidates into the structured seed:

```bash
python3 ../scripts/import_figma_course.py --no-fetch
```

The merge is idempotent, keeps existing curated sections and challenges,
creates missing `sections -> units -> skills` from `normalized_course.json`,
generates required `theory/single_choice/true_false` challenges for every
skill, appends interactive placeholders when source data is missing, and
validates `answers` references for map/photo formats.
