import json
from pathlib import Path

DATA = Path(__file__).resolve().parent.parent / "engine" / "internal" / "engine" / "city" / "cities_data.json"


def load():
    return json.loads(DATA.read_text(encoding="utf-8"))


def all_records():
    data = load()
    for level in ("cities", "counties", "world"):
        for record in data[level]:
            yield level, record


def test_city_data_has_unique_stable_ids():
    ids = [record["id"] for _, record in all_records()]
    assert len(ids) == len(set(ids))
    assert any(value.startswith("CN-") for value in ids)
    assert any(value.startswith("GN-") for value in ids)


def test_every_record_has_three_locales():
    for level, record in all_records():
        for locale in ("zh-Hans", "zh-Hant", "en"):
            assert record["names"][locale], (level, record["name"], locale)
            assert record["region_names"][locale], (level, record["name"], locale)
        assert not any("\u3400" <= char <= "\u9fff" for char in record["names"]["en"])
        assert not any("\u3400" <= char <= "\u9fff" for char in record["region_names"]["en"])


def test_known_places_are_localized():
    by_id = {record["id"]: record for _, record in all_records()}
    beijing = by_id["CN-110000"]
    assert beijing["names"] == {"zh-Hans": "北京市", "zh-Hant": "北京市", "en": "Beijing"}
    changsha = by_id["CN-430100"]
    assert changsha["names"]["zh-Hant"] == "長沙市"
    assert changsha["names"]["en"] == "Changsha"
    taipei = by_id["GN-1668341"]
    assert taipei["names"]["zh-Hant"] == "台北"
    assert taipei["names"]["en"] == "Taipei"
    assert taipei["region_names"]["zh-Hant"] == "中國台灣"


def test_search_aliases_cover_each_locale():
    for _, record in all_records():
        aliases = set(record.get("aliases", []))
        assert record["names"]["zh-Hans"] in aliases
        assert record["names"]["zh-Hant"] in aliases
        assert record["names"]["en"] in aliases
