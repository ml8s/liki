"""Build the factor-matching index from canonical Ziwei palaces.

This is counsel's internal projection: the canonical chart remains
subject + bazi + shi_er_gong; engine does not publish a duplicate flat fact list.
"""


def _star_groups(star: dict) -> list[str]:
    groups: list[str] = []
    for lei in star.get("lei") or []:
        if lei == "zhu_xing":
            groups.append("紫微主星")
        elif lei == "liu_ji":
            groups.append("紫微六吉星")
        elif lei == "wen_xing":
            groups.append("紫微文星")
        elif lei == "sha_xing":
            groups.append("煞星")
    return groups


def build_palace_facts(ziwei: dict) -> list[dict]:
    facts: list[dict] = []

    def add(palace: str, kind: str, target: str, star: str = "", value: str = "") -> None:
        facts.append({
            "palace": palace,
            "kind": kind,
            "target": target,
            "star": star,
            "value": value,
        })

    for palace in ziwei.get("shi_er_gong") or []:
        name = palace.get("ming") or ""
        major_count = 0
        for star in palace.get("xing_yao") or []:
            groups = _star_groups(star)
            if "紫微主星" in groups:
                major_count += 1
            star_name = star.get("ming") or ""
            add(name, "star", star_name, star_name)
            for group in groups:
                add(name, "star", group, star_name)
            si_hua = star.get("si_hua")
            if si_hua:
                add(name, "si_hua", si_hua, star_name)
            brightness = star.get("liang_du")
            if brightness:
                add(name, "brightness", brightness, star_name, brightness)
                if "紫微主星" in groups:
                    add(name, "brightness", "紫微主星", star_name, brightness)
                if brightness in {"庙", "旺"}:
                    add(name, "brightness", "庙旺", star_name, brightness)
                elif brightness == "陷":
                    add(name, "brightness", "落陷", star_name, brightness)

        if major_count == 0:
            add(name, "special", "无主星", "", "true")
        elif major_count == 1:
            for star in palace.get("xing_yao") or []:
                if "紫微主星" in _star_groups(star):
                    add(name, "special", "唯一主星", star.get("ming") or "", "true")

        for star_name in palace.get("za_yao") or []:
            add(name, "star", star_name, star_name)

    return facts
