#!/usr/bin/env python3
"""Import Figma course text and idempotently enrich the history seed.

The script intentionally uses only Python stdlib so it can run in this repo
without installing extra tooling. It never reads tokens from files and never
writes secrets to disk.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import sys
import urllib.request
from collections import Counter
from pathlib import Path
from typing import Any


ROOT = Path(__file__).resolve().parents[1]
DEFAULT_SEED = ROOT / "backend" / "seeds" / "history_course_structured.json"
DEFAULT_RAW = ROOT / "backend" / "seeds" / "figma_course_raw.json"
DEFAULT_NORMALIZED = ROOT / "backend" / "seeds" / "normalized_course.json"

SUPPORTED_TYPES = {
    "theory",
    "single_choice",
    "multiple_choice",
    "timeline",
    "match_pairs",
    "image_question",
    "match_image",
    "match_photos",
    "quote_question",
    "true_false",
    "fill_in_blank",
    "map_point",
    "map_area",
}


def main() -> int:
    parser = argparse.ArgumentParser(description="Fetch Figma course text and enrich the history seed.")
    parser.add_argument("--seed", default=str(DEFAULT_SEED), help="Path to history_course_structured.json")
    parser.add_argument("--raw", default=str(DEFAULT_RAW), help="Path for extracted Figma raw JSON")
    parser.add_argument("--normalized", default=str(DEFAULT_NORMALIZED), help="Path to normalized_course.json for interactive merge")
    parser.add_argument("--no-fetch", action="store_true", help="Skip Figma API request and use an existing raw file if present")
    parser.add_argument("--dry-run", action="store_true", help="Validate and print a summary without writing the seed")
    args = parser.parse_args()

    seed_path = Path(args.seed)
    raw_path = Path(args.raw)
    normalized_path = Path(args.normalized)

    figma = None
    if not args.no_fetch:
        figma = fetch_figma_from_env()
        if figma is not None:
            raw = extract_figma_text(figma)
            raw_path.parent.mkdir(parents=True, exist_ok=True)
            raw_path.write_text(json.dumps(raw, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
            print(f"saved Figma text nodes: {raw_path}")
    if figma is None and raw_path.exists():
        print(f"using existing Figma raw: {raw_path}")
        raw = json.loads(raw_path.read_text(encoding="utf-8"))
    elif figma is None:
        raw = {"text_nodes": [], "ambiguous_items": ["FIGMA_TOKEN/FIGMA_FILE_KEY are not set; seed-only enrichment was used."]}
        print("Figma env is missing; using seed-only enrichment.", file=sys.stderr)

    seed = json.loads(seed_path.read_text(encoding="utf-8"))
    normalized = json.loads(normalized_path.read_text(encoding="utf-8")) if normalized_path.exists() else None
    before = count_challenges(seed)
    added = enrich_seed(seed, raw, normalized)
    validate_seed(seed)
    after = count_challenges(seed)
    media_stats = match_photos_stats(seed)

    if args.dry_run:
        print(json.dumps({"before": before, "after": after, "would_add": added, "match_photos": media_stats}, ensure_ascii=False, indent=2))
        return 0

    seed_path.write_text(json.dumps(seed, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({"before": before, "after": after, "added": added, "match_photos": media_stats}, ensure_ascii=False, indent=2))
    return 0


def fetch_figma_from_env() -> dict[str, Any] | None:
    token = os.environ.get("FIGMA_TOKEN", "").strip()
    file_key = os.environ.get("FIGMA_FILE_KEY", "").strip()
    if not token or not file_key:
        return None
    req = urllib.request.Request(
        f"https://api.figma.com/v1/files/{file_key}",
        headers={"X-Figma-Token": token},
    )
    with urllib.request.urlopen(req, timeout=60) as response:
        return json.loads(response.read().decode("utf-8"))


def extract_figma_text(figma: dict[str, Any]) -> dict[str, Any]:
    nodes: list[dict[str, Any]] = []

    def walk(node: dict[str, Any], path: list[str]) -> None:
        node_type = node.get("type", "")
        name = str(node.get("name", "")).strip()
        next_path = path
        if node_type in {"DOCUMENT", "CANVAS", "FRAME", "SECTION", "GROUP", "COMPONENT", "INSTANCE"} and name:
            next_path = [*path, name]
        if node_type == "TEXT":
            box = node.get("absoluteBoundingBox") or {}
            text = str(node.get("characters", "")).strip()
            if text:
                nodes.append(
                    {
                        "text": text,
                        "path": next_path,
                        "x": box.get("x", 0),
                        "y": box.get("y", 0),
                        "width": box.get("width", 0),
                        "height": box.get("height", 0),
                    }
                )
        for child in node.get("children") or []:
            if isinstance(child, dict):
                walk(child, next_path)

    document = figma.get("document") or {}
    walk(document, [])
    nodes.sort(key=lambda item: (float(item.get("x") or 0), float(item.get("y") or 0), item["text"]))
    return {"text_nodes": nodes, "ambiguous_items": []}


def enrich_seed(seed: dict[str, Any], raw: dict[str, Any], normalized: dict[str, Any] | None = None) -> dict[str, int]:
    if normalized:
        merge_normalized_course(seed, normalized)
    all_facts = collect_all_facts(seed)
    figma_texts = [node.get("text", "") for node in raw.get("text_nodes", []) if isinstance(node, dict)]
    normalized_skills = collect_normalized_skills(normalized) if normalized else []
    added = Counter()

    for section in seed.get("sections", []):
        for unit in section.get("units", []):
            for skill in unit.get("skills", []):
                normalized_skill = best_normalized_skill(skill.get("title", ""), normalized_skills)
                facts = merged_facts(skill, figma_texts if not normalized_skill else [])
                if normalized_skill:
                    merge_skill_content(skill, normalized_skill)
                    facts = [clean_text(item) for item in skill.get("facts", []) if clean_text(item)]
                challenges = skill.setdefault("challenges", [])
                apply_interactive_publish_policy(challenges)
                existing = {challenge.get("type") for challenge in challenges}
                signatures = {challenge_signature(challenge) for challenge in challenges}

                candidates: list[dict[str, Any]] = []
                if "theory" not in existing:
                    candidates.append(theory_challenge(skill, facts))
                if "single_choice" not in existing:
                    candidates.append(single_choice_challenge(skill, facts, all_facts))
                if "true_false" not in existing:
                    candidates.append(true_false_challenge(skill, facts))
                if "fill_in_blank" not in existing:
                    fill = fill_blank_challenge(skill, facts)
                    if fill is not None:
                        candidates.append(fill)
                if "match_pairs" not in existing:
                    pairs = explicit_pairs(skill, facts)
                    if pairs:
                        candidates.append(match_pairs_challenge(pairs))
                if not has_result_challenge(challenges):
                    results = [clean_text(item) for item in skill.get("results", []) if clean_text(item)]
                    if results:
                        candidates.append(results_challenge(skill, results, all_facts))
                if normalized_skill:
                    if "map_point" not in existing:
                        for item in skill.get("map_points", [])[:1]:
                            candidates.append(map_point_challenge(item))
                    if "map_area" not in existing:
                        for item in skill.get("map_areas", [])[:1]:
                            candidates.append(map_area_challenge(item))
                    if "match_photos" not in existing:
                        for item in skill.get("photo_matches", [])[:1]:
                            candidates.append(match_photos_challenge(item))
                if "map_point" not in existing and not any(candidate.get("type") == "map_point" for candidate in candidates):
                    candidates.append(placeholder_map_point_challenge(skill, facts))
                if "map_area" not in existing and not any(candidate.get("type") == "map_area" for candidate in candidates):
                    candidates.append(placeholder_map_area_challenge(skill, facts))
                if "match_photos" not in existing and not any(candidate.get("type") == "match_photos" for candidate in candidates):
                    candidates.append(placeholder_match_photos_challenge(skill, facts))

                for candidate in candidates:
                    sig = challenge_signature(candidate)
                    if sig in signatures:
                        continue
                    candidate["position"] = len(challenges) + 1
                    challenges.append(candidate)
                    signatures.add(sig)
                    added[candidate["type"]] += 1

                normalize_positions(challenges)
    return dict(added)


def apply_interactive_publish_policy(challenges: list[dict[str, Any]]) -> None:
    for challenge in challenges:
        if challenge.get("type") not in {"map_point", "map_area", "match_photos"}:
            continue
        tags = ensure_tags(challenge)
        if challenge.get("type") == "match_photos":
            if has_real_match_photos(challenge):
                if challenge.get("status") in {"", None, "draft"} and "placeholder" not in tags and "needs_review" not in tags:
                    challenge["status"] = "published"
                continue
            for tag in ("placeholder", "needs_review"):
                if tag not in tags:
                    tags.append(tag)
            challenge["status"] = "draft"
            continue
        if "placeholder" not in tags and ("needs_review" in tags or not has_real_interactive_asset(challenge)):
            tags.append("placeholder")
        if challenge.get("status") in {"", None, "draft"}:
            challenge["status"] = "published"


def has_real_interactive_asset(challenge: dict[str, Any]) -> bool:
    if challenge.get("type") != "match_photos":
        return "needs_review" not in challenge.get("tags", [])
    return has_real_match_photos(challenge)


def has_real_match_photos(challenge: dict[str, Any]) -> bool:
    options = challenge.get("options")
    photos = options.get("photos", []) if isinstance(options, dict) else []
    return bool(photos) and all(isinstance(photo, dict) and is_real_media_url(photo.get("image_url")) for photo in photos)


def is_real_media_url(value: Any) -> bool:
    url = clean_text(value)
    return url.startswith("http://") or url.startswith("https://") or url.startswith("/media/")


def ensure_tags(challenge: dict[str, Any]) -> list[str]:
    tags = challenge.get("tags")
    if not isinstance(tags, list):
        tags = []
        challenge["tags"] = tags
    normalized: list[str] = []
    for tag in tags:
        text = clean_text(tag)
        if text and text not in normalized:
            normalized.append(text)
    challenge["tags"] = normalized
    return normalized


def merge_normalized_course(seed: dict[str, Any], normalized: dict[str, Any]) -> None:
    sections = seed.setdefault("sections", [])
    for normalized_section in normalized.get("sections", []):
        theme = clean_text(normalized_section.get("theme", ""))
        if not theme:
            continue
        section = find_by_title(sections, "theme", theme)
        if section is None:
            section = {
                "theme": theme,
                "description": clean_text(normalized_section.get("description", "")),
                "units": [],
            }
            sections.append(section)
        elif not clean_text(section.get("description", "")) and clean_text(normalized_section.get("description", "")):
            section["description"] = clean_text(normalized_section.get("description", ""))

        units = section.setdefault("units", [])
        for normalized_unit in normalized_section.get("units", []):
            unit_title = clean_text(normalized_unit.get("title", ""))
            if not unit_title:
                continue
            unit = find_by_title(units, "title", unit_title)
            if unit is None:
                unit = {"title": unit_title, "skills": []}
                units.append(unit)
            skills = unit.setdefault("skills", [])
            for normalized_skill in normalized_unit.get("skills", []):
                skill_title = clean_text(normalized_skill.get("title", ""))
                if not skill_title:
                    continue
                skill = find_by_title(skills, "title", skill_title)
                if skill is None:
                    skill = {
                        "title": skill_title,
                        "facts": [],
                        "challenges": [],
                    }
                    skills.append(skill)
                merge_skill_content(skill, normalized_skill)


def find_by_title(items: list[dict[str, Any]], field: str, title: str) -> dict[str, Any] | None:
    title_norm = norm(title)
    for item in items:
        if norm(item.get(field, "")) == title_norm:
            return item
    return None


def merge_skill_content(skill: dict[str, Any], source: dict[str, Any]) -> None:
    for field in ("facts", "results", "map_points", "map_areas", "photo_matches"):
        existing = skill.setdefault(field, [])
        if not isinstance(existing, list):
            existing = []
            skill[field] = existing
        for item in source.get(field, []):
            if not contains_equivalent(existing, item):
                existing.append(item)

    existing_pairs = skill.setdefault("cause_effect", [])
    if not isinstance(existing_pairs, list):
        existing_pairs = []
        skill["cause_effect"] = existing_pairs
    for item in source.get("cause_effect", []):
        if isinstance(item, dict) and item.get("cause") and item.get("effect") and not contains_equivalent(existing_pairs, item):
            existing_pairs.append(item)


def contains_equivalent(items: list[Any], candidate: Any) -> bool:
    candidate_key = comparable_key(candidate)
    return any(comparable_key(item) == candidate_key for item in items)


def comparable_key(value: Any) -> str:
    if isinstance(value, dict):
        return json.dumps(value, ensure_ascii=False, sort_keys=True)
    return norm(value)


def collect_normalized_skills(normalized: dict[str, Any] | None) -> list[dict[str, Any]]:
    if not normalized:
        return []
    out: list[dict[str, Any]] = []
    for section in normalized.get("sections", []):
        for unit in section.get("units", []):
            for skill in unit.get("skills", []):
                enriched = dict(skill)
                enriched["_section"] = section.get("theme", "")
                enriched["_unit"] = unit.get("title", "")
                out.append(enriched)
    return out


def best_normalized_skill(title: str, skills: list[dict[str, Any]]) -> dict[str, Any] | None:
    title_tokens = token_set(title)
    if not title_tokens:
        return None
    best: tuple[float, dict[str, Any] | None] = (0, None)
    for skill in skills:
        candidate_text = " ".join([skill.get("title", ""), skill.get("_unit", ""), skill.get("_section", "")])
        candidate_tokens = token_set(candidate_text)
        if not candidate_tokens:
            continue
        overlap = len(title_tokens & candidate_tokens)
        score = overlap / max(len(title_tokens), 1)
        if score > best[0]:
            best = (score, skill)
    return best[1] if best[0] >= 0.34 else None


def token_set(value: str) -> set[str]:
    stop = {"и", "в", "на", "с", "о", "об", "по", "к", "xx", "xix"}
    return {item for item in re.findall(r"[а-яА-ЯёЁa-zA-Z0-9]+", norm(value)) if len(item) > 2 and item not in stop}


def merged_facts(skill: dict[str, Any], figma_texts: list[str]) -> list[str]:
    facts = [clean_text(item) for item in skill.get("facts", []) if clean_text(item)]
    title_norm = norm(skill.get("title", ""))
    for text in figma_texts:
        value = clean_text(text)
        if not value or len(value) < 12:
            continue
        if title_norm and (title_norm in norm(value) or norm(value) in title_norm):
            continue
        if any(norm(value) == norm(existing) for existing in facts):
            continue
        if looks_like_fact(value):
            facts.append(value)
    return facts


def theory_challenge(skill: dict[str, Any], facts: list[str]) -> dict[str, Any]:
    body = "\n".join(facts[:5]) if facts else f"Краткий материал по теме «{skill.get('title', '')}»."
    return base_challenge(
        "theory",
        "Изучи краткий материал",
        body=body,
        payload={"facts": facts[:5], "summary": f"Ключевые факты по теме «{skill.get('title', '')}»."},
        explanation="",
        tags=["seed", "theory", "figma_import"],
    )


def single_choice_challenge(skill: dict[str, Any], facts: list[str], all_facts: list[str]) -> dict[str, Any]:
    correct = best_fact(facts) or f"Тема связана с «{skill.get('title', '')}»."
    distractors = same_kind_distractors(correct, facts, all_facts)
    options = [{"id": option_id(i), "text": text} for i, text in enumerate([correct, *distractors[:2]])]
    return base_challenge(
        "single_choice",
        f"Какое утверждение точнее всего относится к теме «{skill.get('title', '')}»?",
        options=options,
        answers=["a"],
        explanation=correct,
        tags=["seed", "quiz", "figma_import"],
    )


def true_false_challenge(skill: dict[str, Any], facts: list[str]) -> dict[str, Any]:
    fact = facts[1] if len(facts) > 1 else (facts[0] if facts else f"Тема называется «{skill.get('title', '')}».")
    return base_challenge(
        "true_false",
        "Верно ли утверждение?",
        body=fact,
        options=[{"id": "true", "text": "Верно"}, {"id": "false", "text": "Неверно"}],
        answers=["true"],
        explanation=fact,
        tags=["seed", "true_false", "figma_import"],
    )


def fill_blank_challenge(skill: dict[str, Any], facts: list[str]) -> dict[str, Any] | None:
    for fact in facts:
        year = re.search(r"\b(18\d{2}|19\d{2})\b", fact)
        if year:
            return base_challenge(
                "fill_in_blank",
                "Заполни пропуск",
                payload={"text": fact[: year.start()] + "____" + fact[year.end() :], "placeholder": "____"},
                answers=[year.group(1)],
                explanation=fact,
                tags=["seed", "date", "figma_import"],
            )
    title = str(skill.get("title", ""))
    for term in ["большевистское", "меньшевистского", "самодержавие", "Государственная дума", "РСДРП"]:
        for fact in facts:
            if term.lower() in fact.lower():
                return base_challenge(
                    "fill_in_blank",
                    "Заполни пропуск",
                    payload={"text": fact.replace(term, "____", 1), "placeholder": "____"},
                    answers=[term],
                    explanation=fact,
                    tags=["seed", "term", "figma_import"],
                )
    if title:
        head = title.split()[0]
        return base_challenge(
            "fill_in_blank",
            "Заполни пропуск",
            payload={"text": f"Тема урока: ____ {' '.join(title.split()[1:])}".strip(), "placeholder": "____"},
            answers=[head],
            explanation=f"Тема урока: «{title}».",
            tags=["seed", "term", "figma_import", "needing_review"],
            status="draft",
        )
    return None


def explicit_pairs(skill: dict[str, Any], facts: list[str]) -> list[tuple[str, str]]:
    raw_pairs: list[tuple[str, str]] = []
    for item in skill.get("cause_effect", []):
        if item.get("cause") and item.get("effect"):
            raw_pairs.append((clean_text(item["cause"]), clean_text(item["effect"])))
    if len(raw_pairs) >= 2:
        return raw_pairs[:4]

    dated = []
    for fact in facts:
        match = re.search(r"\b(18\d{2}|19\d{2})\b", fact)
        if match:
            dated.append((match.group(1), fact))
    unique: list[tuple[str, str]] = []
    seen = set()
    for year, fact in dated:
        if year in seen:
            continue
        seen.add(year)
        unique.append((year, fact))
    return unique[:4] if len(unique) >= 2 else []


def match_pairs_challenge(pairs: list[tuple[str, str]]) -> dict[str, Any]:
    left = [{"id": f"l{i+1}", "text": left_text} for i, (left_text, _) in enumerate(pairs)]
    right = [{"id": f"r{i+1}", "text": right_text} for i, (_, right_text) in enumerate(pairs)]
    needs_review = has_duplicate_normalized_text([item["text"] for item in left]) or has_duplicate_normalized_text([item["text"] for item in right])
    return base_challenge(
        "match_pairs",
        "Соотнеси элементы",
        options={"left": left, "right": right},
        answers=[{"left_id": f"l{i+1}", "right_id": f"r{i+1}"} for i in range(len(pairs))],
        explanation="Пары связаны с материалом темы.",
        difficulty="medium",
        tags=["seed", "match_pairs", "figma_import", *review_tags(needs_review)],
        status="draft" if needs_review else "published",
    )


def has_result_challenge(challenges: list[dict[str, Any]]) -> bool:
    return any("results" in challenge.get("tags", []) for challenge in challenges)


def results_challenge(skill: dict[str, Any], results: list[str], all_facts: list[str]) -> dict[str, Any]:
    correct = results[0]
    distractors = [fact for fact in same_kind_distractors(correct, results, all_facts) if norm(fact) != norm(correct)]
    while len(distractors) < 2:
        fallback = f"Итог не относится к теме «{skill.get('title', '')}»."
        if all(norm(item) != norm(fallback) for item in [correct, *distractors]):
            distractors.append(fallback)
        else:
            distractors.append(f"Лишний вариант {len(distractors) + 1}")
    options = [{"id": option_id(i), "text": text} for i, text in enumerate([correct, *distractors[:2]])]
    return base_challenge(
        "single_choice",
        f"Какой итог относится к теме «{skill.get('title', '')}»?",
        options=options,
        answers=["a"],
        explanation=correct,
        tags=["seed", "quiz", "results", "figma_import"],
    )


def map_point_challenge(item: dict[str, Any]) -> dict[str, Any]:
    title = str(item.get("title", "")).strip() or "географический объект"
    needs_review = bool(item.get("needs_review"))
    lat = float(item.get("lat", 0))
    lng = float(item.get("lng", 0))
    radius = float(item.get("radius_m", 30000))
    return base_challenge(
        "map_point",
        f"Укажи {title} на карте",
        body=str(item.get("context", "")).strip(),
        payload=map_payload(lat, lng, zoom=6),
        answers={"lat": lat, "lng": lng, "radius_m": radius},
        explanation=str(item.get("context", "")).strip() or f"{title} упоминается в материале темы.",
        tags=["seed", "map", *review_tags(needs_review), *placeholder_tags(needs_review)],
        status="published",
    )


def map_area_challenge(item: dict[str, Any]) -> dict[str, Any]:
    title = str(item.get("title", "")).strip() or "область"
    needs_review = bool(item.get("needs_review"))
    center = item.get("center") if isinstance(item.get("center"), dict) else {}
    lat = float(center.get("lat", 0))
    lng = float(center.get("lng", 0))
    return base_challenge(
        "map_area",
        f"Обведи примерную область: {title}",
        body=str(item.get("context", "")).strip(),
        payload=map_payload(lat, lng, zoom=float(item.get("zoom", 5))),
        answers={
            "center": {"lat": lat, "lng": lng},
            "area_m2": float(item.get("area_m2", 1)),
            "center_radius_m": float(item.get("center_radius_m", 60000)),
            "area_tolerance": float(item.get("area_tolerance", 0.8)),
        },
        explanation=str(item.get("context", "")).strip() or f"{title} дана как примерная историческая область.",
        difficulty="medium",
        tags=["seed", "map", *review_tags(needs_review), *placeholder_tags(needs_review)],
        status="published",
    )


def match_photos_challenge(item: dict[str, Any]) -> dict[str, Any]:
    needs_review = bool(item.get("needs_review"))
    source_items = item.get("items") if isinstance(item.get("items"), list) else []
    photos = []
    labels = []
    answers = []
    explanations = []
    for index, photo in enumerate(source_items[:4], start=1):
        if not isinstance(photo, dict):
            continue
        photo_id = f"p{index}"
        label_id = f"l{index}"
        alt = str(photo.get("alt", "")).strip() or f"Фото {index}"
        label = str(photo.get("label", "")).strip() or alt
        photos.append({"id": photo_id, "image_url": str(photo.get("image_url", "")).strip(), "alt": alt})
        labels.append({"id": label_id, "text": label})
        answers.append({"photo_id": photo_id, "label_id": label_id})
        explanations.append(f"{alt}: {label}")
    return base_challenge(
        "match_photos",
        str(item.get("prompt", "")).strip() or "Соотнеси изображения с описаниями",
        body="Выбери, какое изображение относится к каждому историческому объекту.",
        payload={},
        options={"photos": photos, "labels": labels},
        answers=answers,
        explanation="; ".join(explanations),
        difficulty="medium",
        tags=["seed", "photos", *review_tags(needs_review), *placeholder_tags(needs_review)],
        status="published",
    )


def placeholder_map_point_challenge(skill: dict[str, Any], facts: list[str]) -> dict[str, Any]:
    place = placeholder_place(skill, facts)
    context = placeholder_context(skill, facts)
    return base_challenge(
        "map_point",
        f"Укажи {place['title']} на карте",
        body=context,
        payload=map_payload(place["lat"], place["lng"], zoom=place["zoom"]),
        answers={"lat": place["lat"], "lng": place["lng"], "radius_m": place["radius_m"]},
        explanation=f"{place['title']} используется как временная точка для интерактивной карты. Координаты можно уточнить в seed/admin.",
        tags=["seed", "map", "placeholder", "needs_review", "figma_import"],
        status="published",
    )


def placeholder_map_area_challenge(skill: dict[str, Any], facts: list[str]) -> dict[str, Any]:
    place = placeholder_place(skill, facts)
    context = placeholder_context(skill, facts)
    return base_challenge(
        "map_area",
        f"Обведи примерную область: {place['title']}",
        body=context,
        payload=map_payload(place["lat"], place["lng"], zoom=max(place["zoom"] - 1, 4)),
        answers={
            "center": {"lat": place["lat"], "lng": place["lng"]},
            "area_m2": 2_500_000_000,
            "center_radius_m": 120_000,
            "area_tolerance": 1.0,
        },
        explanation=f"Это временная область для проверки механики карты по теме «{skill.get('title', '')}».",
        difficulty="medium",
        tags=["seed", "map", "placeholder", "needs_review", "figma_import"],
        status="published",
    )


def placeholder_match_photos_challenge(skill: dict[str, Any], facts: list[str]) -> dict[str, Any]:
    labels = placeholder_photo_labels(skill, facts)
    photos = [
        {"id": f"p{index}", "image_url": "", "alt": label}
        for index, label in enumerate(labels, start=1)
    ]
    options = {
        "photos": photos,
        "labels": [{"id": f"l{index}", "text": label} for index, label in enumerate(labels, start=1)],
    }
    return base_challenge(
        "match_photos",
        f"Соотнеси фото-заглушки по теме «{skill.get('title', '')}»",
        body="Изображения можно заменить позже; пока ориентируйся на alt-текст.",
        payload={},
        options=options,
        answers=[{"photo_id": f"p{index}", "label_id": f"l{index}"} for index in range(1, len(labels) + 1)],
        explanation="Временное фото-задание: реальные изображения нужно добавить через seed/admin.",
        difficulty="medium",
        tags=["seed", "photos", "placeholder", "needs_review", "figma_import"],
        status="published",
    )


def placeholder_place(skill: dict[str, Any], facts: list[str]) -> dict[str, Any]:
    haystack = norm(" ".join([str(skill.get("title", "")), *facts]))
    places = [
        ("санкт-петербург", "Санкт-Петербург", 59.9343, 30.3351, 7.0, 60_000),
        ("петербург", "Санкт-Петербург", 59.9343, 30.3351, 7.0, 60_000),
        ("петроград", "Петроград", 59.9343, 30.3351, 7.0, 60_000),
        ("москва", "Москва", 55.7558, 37.6173, 7.0, 60_000),
        ("минск", "Минск", 53.9006, 27.5590, 7.0, 50_000),
        ("симбирск", "Симбирск", 54.3142, 48.4031, 7.0, 50_000),
        ("ульяновск", "Ульяновск", 54.3142, 48.4031, 7.0, 50_000),
        ("алма-ата", "Алма-Ата", 43.2389, 76.8897, 7.0, 70_000),
        ("казахстан", "Казахстан", 48.0196, 66.9237, 4.0, 500_000),
        ("германия", "Германия", 51.1657, 10.4515, 5.0, 350_000),
    ]
    for needle, title, lat, lng, zoom, radius in places:
        if needle in haystack:
            return {"title": title, "lat": lat, "lng": lng, "zoom": zoom, "radius_m": radius}
    return {"title": "Москва", "lat": 55.7558, "lng": 37.6173, "zoom": 6.0, "radius_m": 80_000}


def placeholder_context(skill: dict[str, Any], facts: list[str]) -> str:
    fact = best_fact(facts)
    if fact:
        return fact
    return f"Временная интерактивная карта по теме «{skill.get('title', '')}»."


def placeholder_photo_labels(skill: dict[str, Any], facts: list[str]) -> list[str]:
    labels: list[str] = []
    title = clean_text(skill.get("title", ""))
    if title:
        labels.append(title)
    for fact in facts:
        text = clean_text(fact)
        if len(text) > 70:
            text = text[:67].rstrip() + "..."
        if text and norm(text) not in {norm(item) for item in labels}:
            labels.append(text)
        if len(labels) >= 2:
            break
    while len(labels) < 2:
        labels.append(f"Фото-заглушка {len(labels) + 1}")
    return labels[:2]


def map_payload(lat: float, lng: float, *, zoom: float) -> dict[str, Any]:
    return {
        "center": {"lat": lat, "lng": lng},
        "zoom": zoom,
        "tile_url_template": "https://tile.openstreetmap.org/{z}/{x}/{y}.png",
        "attribution": "© OpenStreetMap contributors",
    }


def review_tags(needs_review: bool) -> list[str]:
    return ["needs_review"] if needs_review else []


def placeholder_tags(needs_review: bool) -> list[str]:
    return ["placeholder"] if needs_review else []


def has_duplicate_normalized_text(items: list[str]) -> bool:
    normalized = [norm(item) for item in items if norm(item)]
    return len(normalized) != len(set(normalized))


def base_challenge(
    challenge_type: str,
    prompt: str,
    *,
    body: str = "",
    payload: Any | None = None,
    options: Any | None = None,
    answers: Any | None = None,
    explanation: str = "",
    difficulty: str = "easy",
    tags: list[str] | None = None,
    status: str = "published",
) -> dict[str, Any]:
    return {
        "type": challenge_type,
        "difficulty": difficulty,
        "tags": tags or ["seed", "figma_import"],
        "level": 1,
        "lesson_count": 1,
        "prompt": prompt,
        "body": body,
        "payload": payload if payload is not None else {},
        "options": options if options is not None else [],
        "answers": answers if answers is not None else [],
        "explanation": explanation,
        "position": 0,
        "status": status,
    }


def validate_seed(seed: dict[str, Any]) -> None:
    errors: list[str] = []
    for section in seed.get("sections", []):
        for unit in section.get("units", []):
            for skill in unit.get("skills", []):
                title = skill.get("title", "")
                challenges = skill.get("challenges", [])
                if not any(challenge.get("type") == "theory" for challenge in challenges):
                    errors.append(f"{title}: missing theory challenge")
                positions = [challenge.get("position") for challenge in challenges]
                if positions != list(range(1, len(challenges) + 1)):
                    errors.append(f"{title}: positions are not sequential")
                for challenge in challenges:
                    validate_challenge(title, challenge, errors)
    if errors:
        raise SystemExit("Seed validation failed:\n" + "\n".join(f"- {error}" for error in errors[:50]))


def validate_challenge(skill_title: str, challenge: dict[str, Any], errors: list[str]) -> None:
    required = ["type", "difficulty", "tags", "level", "lesson_count", "prompt", "body", "payload", "options", "answers", "explanation", "position", "status"]
    for field in required:
        if field not in challenge:
            errors.append(f"{skill_title}: challenge missing {field}")
    if challenge.get("type") not in SUPPORTED_TYPES:
        errors.append(f"{skill_title}: unsupported challenge type {challenge.get('type')}")
    options = challenge.get("options")
    answers = challenge.get("answers")
    ctype = challenge.get("type")
    if ctype in {"single_choice", "multiple_choice", "true_false", "image_question"} and isinstance(options, list):
        ids = {item.get("id") for item in options if isinstance(item, dict)}
        for answer in answers if isinstance(answers, list) else []:
            if isinstance(answer, str) and answer not in ids:
                errors.append(f"{skill_title}: answer {answer!r} has no option")
    if ctype == "match_pairs" and isinstance(options, dict) and isinstance(answers, list):
        left_ids = {item.get("id") for item in options.get("left", []) if isinstance(item, dict)}
        right_ids = {item.get("id") for item in options.get("right", []) if isinstance(item, dict)}
        for answer in answers:
            if not isinstance(answer, dict) or answer.get("left_id") not in left_ids or answer.get("right_id") not in right_ids:
                errors.append(f"{skill_title}: invalid match_pairs answer")
    if ctype == "map_point":
        center = challenge.get("payload", {}).get("center") if isinstance(challenge.get("payload"), dict) else None
        answer = challenge.get("answers")
        if not has_lat_lng(center) or "zoom" not in challenge.get("payload", {}) or not has_lat_lng(answer) or not number_like(answer.get("radius_m") if isinstance(answer, dict) else None):
            errors.append(f"{skill_title}: invalid map_point payload/answers")
    if ctype == "map_area":
        payload = challenge.get("payload") if isinstance(challenge.get("payload"), dict) else {}
        answer = challenge.get("answers") if isinstance(challenge.get("answers"), dict) else {}
        if not has_lat_lng(payload.get("center")) or not has_lat_lng(answer.get("center")):
            errors.append(f"{skill_title}: invalid map_area centers")
        for field in ("area_m2", "center_radius_m", "area_tolerance"):
            if not number_like(answer.get(field)):
                errors.append(f"{skill_title}: invalid map_area {field}")
    if ctype == "match_photos":
        if not validate_match_photos_options_answers(options, answers):
            errors.append(f"{skill_title}: invalid match_photos options/answers")


def has_lat_lng(value: Any) -> bool:
    return isinstance(value, dict) and number_like(value.get("lat")) and number_like(value.get("lng"))


def number_like(value: Any) -> bool:
    return isinstance(value, (int, float)) and not isinstance(value, bool)


def validate_match_photos_options_answers(options: Any, answers: Any) -> bool:
    if not isinstance(options, dict) or not isinstance(answers, list):
        return False
    photos = options.get("photos")
    labels = options.get("labels")
    if not isinstance(photos, list) or not isinstance(labels, list) or not photos or not labels:
        return False
    photo_ids = set()
    for photo in photos:
        if not isinstance(photo, dict) or not photo.get("id") or not photo.get("alt"):
            return False
        photo_ids.add(photo["id"])
    label_ids = set()
    for label in labels:
        if not isinstance(label, dict) or not label.get("id") or not label.get("text"):
            return False
        label_ids.add(label["id"])
    for answer in answers:
        if not isinstance(answer, dict) or answer.get("photo_id") not in photo_ids or answer.get("label_id") not in label_ids:
            return False
    return True


def collect_all_facts(seed: dict[str, Any]) -> list[str]:
    facts: list[str] = []
    for section in seed.get("sections", []):
        for unit in section.get("units", []):
            for skill in unit.get("skills", []):
                facts.extend(clean_text(item) for item in skill.get("facts", []) if clean_text(item))
    return facts


def best_fact(facts: list[str]) -> str:
    for fact in facts:
        if len(fact) > 25:
            return fact
    return facts[0] if facts else ""


def same_kind_distractors(correct: str, local_facts: list[str], all_facts: list[str]) -> list[str]:
    has_year = bool(re.search(r"\b(18\d{2}|19\d{2})\b", correct))
    pool = [fact for fact in [*local_facts, *all_facts] if fact and norm(fact) != norm(correct)]
    if has_year:
        dated = [fact for fact in pool if re.search(r"\b(18\d{2}|19\d{2})\b", fact)]
        if len(dated) >= 2:
            return unique(dated)[:2]
    return unique(pool)[:2]


def challenge_signature(challenge: dict[str, Any]) -> str:
    return "|".join([str(challenge.get("type", "")), norm(challenge.get("prompt", "")), norm(challenge.get("body", "")), norm(challenge.get("explanation", ""))])


def normalize_positions(challenges: list[dict[str, Any]]) -> None:
    challenges.sort(key=lambda item: int(item.get("position") or 10_000))
    for index, challenge in enumerate(challenges, start=1):
        challenge["position"] = index


def count_challenges(seed: dict[str, Any]) -> int:
    return sum(len(skill.get("challenges", [])) for section in seed.get("sections", []) for unit in section.get("units", []) for skill in unit.get("skills", []))


def match_photos_stats(seed: dict[str, Any]) -> dict[str, int]:
    stats = Counter()
    for section in seed.get("sections", []):
        for unit in section.get("units", []):
            for skill in unit.get("skills", []):
                for challenge in skill.get("challenges", []):
                    if challenge.get("type") != "match_photos":
                        continue
                    stats["total"] += 1
                    status = str(challenge.get("status", ""))
                    if status:
                        stats[status] += 1
                    tags = challenge.get("tags") if isinstance(challenge.get("tags"), list) else []
                    if "placeholder" in tags:
                        stats["placeholder"] += 1
                    if "needs_review" in tags:
                        stats["needs_review"] += 1
                    if has_real_match_photos(challenge):
                        stats["with_real_images"] += 1
                        if status == "published":
                            stats["published_real"] += 1
                    else:
                        stats["with_empty_images"] += 1
                        if status == "draft":
                            stats["draft_placeholders"] += 1
                        if status == "published":
                            stats["published_empty"] += 1
    return dict(stats)


def looks_like_fact(text: str) -> bool:
    return len(text) > 20 and text.endswith((".", "!", "?")) and not text.startswith(("http://", "https://"))


def option_id(index: int) -> str:
    return chr(ord("a") + index)


def clean_text(value: Any) -> str:
    return re.sub(r"\s+", " ", str(value).strip())


def norm(value: Any) -> str:
    return clean_text(value).casefold().replace("ё", "е")


def unique(items: list[str]) -> list[str]:
    out: list[str] = []
    seen = set()
    for item in items:
        key = norm(item)
        if key and key not in seen:
            seen.add(key)
            out.append(item)
    return out


if __name__ == "__main__":
    raise SystemExit(main())
