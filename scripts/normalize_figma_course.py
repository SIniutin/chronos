#!/usr/bin/env python3
"""Normalize extracted Figma text nodes into a course-like JSON shape."""

from __future__ import annotations

import argparse
import json
import re
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any


ROOT = Path(__file__).resolve().parents[1]
DEFAULT_RAW = ROOT / "backend" / "seeds" / "figma_course_raw.json"
DEFAULT_OUT = ROOT / "backend" / "seeds" / "normalized_course.json"


@dataclass
class Node:
    index: int
    text: str
    path: list[str]
    x: float
    y: float
    width: float
    height: float
    is_heading: bool = False
    is_section: bool = False


@dataclass
class Topic:
    title: str
    node: Node
    content: list[Node] = field(default_factory=list)


PLACE_COORDS = {
    "Санкт-Петербург": {"lat": 59.9343, "lng": 30.3351, "radius_m": 30000},
    "Петербург": {"lat": 59.9343, "lng": 30.3351, "radius_m": 30000},
    "Петроград": {"lat": 59.9343, "lng": 30.3351, "radius_m": 30000},
    "Москва": {"lat": 55.7558, "lng": 37.6173, "radius_m": 25000},
    "Минск": {"lat": 53.9006, "lng": 27.5590, "radius_m": 25000},
    "Симбирск": {"lat": 54.3187, "lng": 48.3978, "radius_m": 25000},
    "Ульяновск": {"lat": 54.3187, "lng": 48.3978, "radius_m": 25000},
    "Сараево": {"lat": 43.8563, "lng": 18.4131, "radius_m": 25000},
    "Брест-Литовск": {"lat": 52.0976, "lng": 23.7341, "radius_m": 25000},
    "Брест": {"lat": 52.0976, "lng": 23.7341, "radius_m": 25000},
    "Кронштадт": {"lat": 59.9917, "lng": 29.7778, "radius_m": 15000},
    "Тбилиси": {"lat": 41.7151, "lng": 44.8271, "radius_m": 25000},
    "Ленинград": {"lat": 59.9343, "lng": 30.3351, "radius_m": 30000},
}

AREA_COORDS = {
    "Балканы": {"lat": 43.8, "lng": 21.0, "area_m2": 550_000_000_000, "center_radius_m": 350000, "zoom": 5},
    "Восточная Пруссия": {"lat": 54.7, "lng": 21.5, "area_m2": 37_000_000_000, "center_radius_m": 150000, "zoom": 6},
    "Кавказ": {"lat": 42.4, "lng": 44.0, "area_m2": 440_000_000_000, "center_radius_m": 350000, "zoom": 5},
    "Дальний Восток": {"lat": 50.0, "lng": 135.0, "area_m2": 6_000_000_000_000, "center_radius_m": 1200000, "zoom": 4},
}

PERSON_PHOTOS = {
    "Сергей Юльевич Витте": ("С. Ю. Витте", "Министр финансов, связанный с денежной реформой"),
    "С. Ю. Витте": ("С. Ю. Витте", "Министр финансов, связанный с денежной реформой"),
    "Николай II": ("Николай II", "Последний российский император"),
    "В. И. Ленин": ("В. И. Ленин", "Лидер большевистского крыла российской социал-демократии"),
    "Владимир Ильич Ульянов": ("В. И. Ленин", "Революционер, связанный с большевистским движением"),
    "Г. В. Плеханов": ("Г. В. Плеханов", "Один из лидеров меньшевистского течения"),
    "П. А. Столыпин": ("П. А. Столыпин", "Председатель Совета министров и автор аграрного курса"),
}

def main() -> int:
    parser = argparse.ArgumentParser(description="Normalize backend/seeds/figma_course_raw.json into normalized_course.json")
    parser.add_argument("--raw", default=str(DEFAULT_RAW), help="Path to figma_course_raw.json")
    parser.add_argument("--out", default=str(DEFAULT_OUT), help="Path to write normalized_course.json")
    args = parser.parse_args()

    raw_path = Path(args.raw)
    out_path = Path(args.out)
    raw = json.loads(raw_path.read_text(encoding="utf-8"))
    nodes = parse_nodes(raw)
    normalized = normalize(nodes)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(normalized, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"normalized sections={len(normalized['sections'])} ambiguous={len(normalized['ambiguous_items'])} -> {out_path}")
    return 0


def parse_nodes(raw: dict[str, Any]) -> list[Node]:
    nodes = []
    for index, item in enumerate(raw.get("text_nodes", [])):
        text = clean(item.get("text", ""))
        if not text:
            continue
        node = Node(
            index=index,
            text=text,
            path=[str(part) for part in item.get("path", [])],
            x=float(item.get("x") or 0),
            y=float(item.get("y") or 0),
            width=float(item.get("width") or 0),
            height=float(item.get("height") or 0),
        )
        node.is_heading = looks_like_heading(node)
        node.is_section = looks_like_section_heading(node)
        nodes.append(node)
    nodes.sort(key=lambda node: (node.x, node.y, node.index))
    return nodes


def normalize(nodes: list[Node]) -> dict[str, Any]:
    headings = [node for node in nodes if node.is_heading]
    section_nodes = [node for node in headings if node.is_section]
    if not section_nodes:
        fallback = Node(index=-1, text="Курс из Figma", path=[], x=0, y=0, width=0, height=0, is_heading=True, is_section=True)
        section_nodes = [fallback]

    section_topics = [Topic(title=node.text, node=node) for node in section_nodes]
    unit_topics: dict[int, list[Topic]] = {section.node.index: [] for section in section_topics}
    ambiguous: list[dict[str, Any]] = []

    for heading in headings:
        if heading.is_section:
            continue
        section = nearest_section(heading, section_topics)
        unit_topics.setdefault(section.node.index, []).append(Topic(title=heading.text, node=heading))

    all_topics = [*section_topics, *(topic for topics in unit_topics.values() for topic in topics)]
    for node in nodes:
        if node.is_heading:
            continue
        topic, score = nearest_topic(node, all_topics)
        if topic is None or score > 5.0:
            ambiguous.append(ambiguous_node(node, "no nearby heading"))
            continue
        topic.content.append(node)

    sections = []
    for section in section_topics:
        units = []
        topics = unit_topics.get(section.node.index, [])
        if section.content:
            topics = [Topic(title=section.title, node=section.node, content=section.content), *topics]
        for topic in sorted(topics, key=lambda item: (item.node.y, item.node.x, item.node.index)):
            skill, topic_ambiguous = topic_to_skill(topic)
            ambiguous.extend(topic_ambiguous)
            units.append({"title": topic.title, "skills": [skill]})
        if not units:
            ambiguous.append(ambiguous_node(section.node, "section has no normalized units"))
            continue
        sections.append(
            {
                "theme": section.title,
                "description": description_for_section(section, units),
                "units": units,
            }
        )

    return {"sections": sections, "ambiguous_items": ambiguous}


def topic_to_skill(topic: Topic) -> tuple[dict[str, Any], list[dict[str, Any]]]:
    facts: list[str] = []
    results: list[str] = []
    cause_effect: list[dict[str, str]] = []
    ambiguous: list[dict[str, Any]] = []

    for node in sorted(topic.content, key=lambda item: (item.y, item.x, item.index)):
        parsed = parse_content_block(node.text)
        facts.extend(parsed["facts"])
        results.extend(parsed["results"])
        cause_effect.extend(parsed["cause_effect"])
        for value in parsed["ambiguous"]:
            ambiguous.append({"text": value, "reason": "content parser could not classify line", "source_node": node.index, "path": node.path})

    facts = unique(facts)
    results = unique(results)
    cause_effect = unique_pairs(cause_effect)
    map_points = extract_map_points(topic.title, facts)
    map_areas = extract_map_areas(topic.title, facts)
    photo_matches = extract_photo_matches(topic.title, facts)
    if not facts and not results and not cause_effect:
        ambiguous.append(ambiguous_node(topic.node, "heading has no classified content"))

    return {
        "title": topic.title,
        "facts": facts,
        "cause_effect": cause_effect,
        "results": results,
        "map_points": map_points,
        "map_areas": map_areas,
        "photo_matches": photo_matches,
    }, ambiguous


def extract_map_points(title: str, facts: list[str]) -> list[dict[str, Any]]:
    text = " ".join([title, *facts])
    out: list[dict[str, Any]] = []
    for place, coords in PLACE_COORDS.items():
        if re.search(rf"\b{re.escape(place)}\b", text, re.IGNORECASE):
            out.append(
                {
                    "title": place,
                    "context": context_for_place(place, facts),
                    "lat": coords["lat"],
                    "lng": coords["lng"],
                    "radius_m": coords["radius_m"],
                    "needs_review": True,
                }
            )
    return unique_map_items(out, "title")


def extract_map_areas(title: str, facts: list[str]) -> list[dict[str, Any]]:
    text = " ".join([title, *facts])
    out: list[dict[str, Any]] = []
    for area, coords in AREA_COORDS.items():
        if area.casefold() in text.casefold():
            out.append(
                {
                    "title": area,
                    "context": context_for_place(area, facts),
                    "center": {"lat": coords["lat"], "lng": coords["lng"]},
                    "zoom": coords["zoom"],
                    "area_m2": coords["area_m2"],
                    "center_radius_m": coords["center_radius_m"],
                    "area_tolerance": 0.8,
                    "needs_review": True,
                }
            )
    return unique_map_items(out, "title")


def extract_photo_matches(title: str, facts: list[str]) -> list[dict[str, Any]]:
    text = " ".join([title, *facts])
    items: list[dict[str, str]] = []
    for marker, (alt, label) in PERSON_PHOTOS.items():
        if marker.casefold() in text.casefold() and not any(item["alt"] == alt for item in items):
            slug = re.sub(r"[^a-z0-9_]+", "_", translit_slug(alt)).strip("_")
            items.append({"image_url": f"/images/history/{slug or 'photo'}.jpg", "alt": alt, "label": label})
    if len(items) < 2:
        return []
    return [
        {
            "prompt": "Соотнеси исторических деятелей с описаниями",
            "items": items[:4],
            "needs_review": True,
        }
    ]


def context_for_place(place: str, facts: list[str]) -> str:
    for fact in facts:
        if place.casefold() in fact.casefold():
            return fact
    return f"Географический объект, упомянутый в теме: {place}."


def unique_map_items(items: list[dict[str, Any]], key: str) -> list[dict[str, Any]]:
    out = []
    seen = set()
    for item in items:
        value = normalize_key(str(item.get(key, "")))
        if value and value not in seen:
            seen.add(value)
            out.append(item)
    return out


def translit_slug(value: str) -> str:
    mapping = {
        "а": "a",
        "б": "b",
        "в": "v",
        "г": "g",
        "д": "d",
        "е": "e",
        "ё": "e",
        "ж": "zh",
        "з": "z",
        "и": "i",
        "й": "i",
        "к": "k",
        "л": "l",
        "м": "m",
        "н": "n",
        "о": "o",
        "п": "p",
        "р": "r",
        "с": "s",
        "т": "t",
        "у": "u",
        "ф": "f",
        "х": "h",
        "ц": "c",
        "ч": "ch",
        "ш": "sh",
        "щ": "sh",
        "ы": "y",
        "э": "e",
        "ю": "yu",
        "я": "ya",
        "ь": "",
        "ъ": "",
    }
    return "".join(mapping.get(ch.casefold(), ch if ch.isascii() else "_") for ch in value).lower()


def parse_content_block(text: str) -> dict[str, Any]:
    lines = split_lines(text)
    facts: list[str] = []
    results: list[str] = []
    cause_effect: list[dict[str, str]] = []
    ambiguous: list[str] = []

    if not lines:
        return {"facts": facts, "results": results, "cause_effect": cause_effect, "ambiguous": ambiguous}

    lower_first = lines[0].casefold()
    if lower_first.startswith(("результат", "итог")):
        results.extend(clean_item(line) for line in lines[1:] if clean_item(line))
        return {"facts": facts, "results": results, "cause_effect": cause_effect, "ambiguous": ambiguous}

    if is_cause_block(lines[0]):
        effect = effect_from_question(lines[0])
        for line in lines[1:]:
            item = clean_item(line)
            if item and not item.startswith("("):
                cause_effect.append({"cause": item.rstrip(";."), "effect": effect})
        return {"facts": facts, "results": results, "cause_effect": cause_effect, "ambiguous": ambiguous}

    current_label = ""
    for line in lines:
        item = clean_item(line)
        if not item:
            continue
        label = item.casefold().rstrip(":")
        if label in {"ключевые элементы", "особенности реформы", "правовые принципы реформы", "предпосылки"}:
            current_label = label
            continue
        if "результат" in label or "итог" in label:
            current_label = "results"
            continue
        if "причин" in label or label.startswith("почему"):
            current_label = label
            continue

        if current_label == "results":
            results.append(item)
        elif current_label.startswith("почему") or "причин" in current_label:
            cause_effect.append({"cause": item.rstrip(";."), "effect": effect_from_question(current_label)})
        elif looks_like_fact(item):
            facts.extend(split_sentences(item))
        else:
            ambiguous.append(item)

    return {"facts": unique(facts), "results": unique(results), "cause_effect": unique_pairs(cause_effect), "ambiguous": ambiguous}


def looks_like_heading(node: Node) -> bool:
    text = node.text
    lines = split_lines(text)
    if len(lines) > 2 or len(text) > 140:
        return False
    if starts_like_sentence(text):
        return False
    if node.path and any(part.lower().startswith("frame") for part in node.path):
        return True
    if uppercase_ratio(text) > 0.55:
        return True
    if not re.search(r"[.!?]\s*$", text) and 2 <= len(text.split()) <= 10:
        return True
    return False


def looks_like_section_heading(node: Node) -> bool:
    text = node.text
    if uppercase_ratio(text) > 0.55:
        return True
    if node.path and any(part.lower().startswith("frame") for part in node.path) and node.width >= 250 and not starts_like_sentence(text):
        return True
    if node.y <= 900 and len(text) >= 18:
        return True
    return False


def nearest_section(node: Node, sections: list[Topic]) -> Topic:
    candidates = [section for section in sections if section.node.x <= node.x + 120]
    if candidates:
        return min(candidates, key=lambda section: (abs(node.x - section.node.x) / 1200) + (abs(node.y - section.node.y) / 3500))
    return min(sections, key=lambda section: abs(node.x - section.node.x) + abs(node.y - section.node.y) / 4)


def nearest_topic(node: Node, topics: list[Topic]) -> tuple[Topic | None, float]:
    if not topics:
        return None, 999.0
    best = None
    best_score = 999.0
    for topic in topics:
        heading = topic.node
        dy = node.y - heading.y
        score = abs(node.x - heading.x) / 750 + abs(dy) / 1700
        if dy < -160:
            score += 1.5
        if heading.is_section:
            score += 0.35
        same_path_bonus = common_suffix_len(node.path, heading.path) * 0.2
        score -= same_path_bonus
        if score < best_score:
            best = topic
            best_score = score
    return best, best_score


def description_for_section(section: Topic, units: list[dict[str, Any]]) -> str:
    for unit in units:
        skill = unit["skills"][0]
        if skill["facts"]:
            return skill["facts"][0]
    return f"Материалы раздела «{section.title}», извлечённые из Figma."


def split_lines(text: str) -> list[str]:
    normalized = text.replace("\u2028", "\n")
    return [clean(line) for line in normalized.split("\n") if clean(line)]


def split_sentences(text: str) -> list[str]:
    parts = re.split(r"(?<=[.!?])\s+(?=[А-ЯA-ZЁ0-9])", text)
    merged: list[str] = []
    for part in [clean(part) for part in parts if clean(part)]:
        if merged and re.search(r"\b([А-ЯA-ZЁ]|г|гг|в)\.$", merged[-1]):
            merged[-1] = f"{merged[-1]} {part}"
        else:
            merged.append(part)
    return merged


def clean_item(text: str) -> str:
    return clean(text).strip("•-–— ").strip()


def clean(text: Any) -> str:
    return re.sub(r"[ \t\r\f\v]+", " ", str(text).replace("\u00a0", " ")).strip()


def looks_like_fact(text: str) -> bool:
    if len(text) < 14:
        return False
    if text.endswith(":"):
        return False
    return bool(re.search(r"[А-Яа-яA-Za-zЁё]", text))


def is_cause_block(line: str) -> bool:
    lower = line.casefold()
    return lower.startswith("почему") or "причины" in lower


def effect_from_question(line: str) -> str:
    lower = clean_item(line).rstrip("?").casefold()
    replacements = [
        ("почему ", ""),
        ("возникали ", "возникновение "),
        ("возникли ", "возникновение "),
        ("реформа была нужна", "необходимость реформы"),
        ("реформу было трудно провести", "трудности проведения реформы"),
        ("причины ", ""),
    ]
    value = lower
    for old, new in replacements:
        value = value.replace(old, new)
    return value.strip(" :;") or "следствие по теме"


def uppercase_ratio(text: str) -> float:
    letters = [ch for ch in text if ch.isalpha()]
    if not letters:
        return 0
    return sum(1 for ch in letters if ch.upper() == ch and ch.lower() != ch) / len(letters)


def starts_like_sentence(text: str) -> bool:
    stripped = text.strip()
    if not stripped:
        return False
    return stripped[0].islower() and stripped.endswith((".", ":"))


def common_suffix_len(left: list[str], right: list[str]) -> int:
    total = 0
    for a, b in zip(reversed(left), reversed(right)):
        if a != b:
            break
        total += 1
    return total


def ambiguous_node(node: Node, reason: str) -> dict[str, Any]:
    return {"text": node.text, "reason": reason, "path": node.path, "x": node.x, "y": node.y}


def unique(items: list[str]) -> list[str]:
    out: list[str] = []
    seen = set()
    for item in items:
        key = normalize_key(item)
        if key and key not in seen:
            seen.add(key)
            out.append(item)
    return out


def unique_pairs(items: list[dict[str, str]]) -> list[dict[str, str]]:
    out: list[dict[str, str]] = []
    seen = set()
    for item in items:
        cause = clean_item(item.get("cause", ""))
        effect = clean_item(item.get("effect", ""))
        key = (normalize_key(cause), normalize_key(effect))
        if cause and effect and key not in seen:
            seen.add(key)
            out.append({"cause": cause, "effect": effect})
    return out


def normalize_key(value: str) -> str:
    return clean(value).casefold().replace("ё", "е")


if __name__ == "__main__":
    raise SystemExit(main())
