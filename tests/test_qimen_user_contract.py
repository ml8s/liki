from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
CARD = ROOT / "skills/liki-divination/app/qimen-snapshot.md"
QUARTER = ROOT / "skills/liki-divination/domains/qimen/quarter.md"
YONGSHEN = ROOT / "skills/liki-divination/domains/qimen/yongshen.md"
LOST_PROPERTY = ROOT / "skills/liki-divination/domains/qimen/lost-property.md"
MISSING_PERSON = ROOT / "skills/liki-divination/domains/qimen/missing-person.md"
CAPTURE_ESCAPE = ROOT / "skills/liki-divination/domains/qimen/capture-escape.md"


def test_qimen_card_keeps_normal_users_on_default_chart() -> None:
    text = CARD.read_text(encoding="utf-8")
    assert "默认 `scope=hour` / `school=zhuanpan` / `dingju_method=chaibu`" in text
    assert "用户没有明确要求，不传方法参数" in text
    assert "不能从“更细”“传统”“准确”推断高级方法" in text


def test_qimen_card_documents_user_phrases_and_defaults() -> None:
    text = CARD.read_text(encoding="utf-8")
    for phrase in (
        "用置闰盘看现在",
        "用洛书飞盘看这件事",
        "用十分钟刻家看看",
        "用十二分钟十分局看看",
        "用金函玉镜看今天",
    ):
        assert phrase in text


def test_qimen_card_guides_matter_selection() -> None:
    text = CARD.read_text(encoding="utf-8")
    assert "普通问事传 `matter`" in text
    assert "`matter` 与 `yong_shen` 互斥" in text
    assert "高级用户可直接传 `yong_shen`" in text


def test_quarter_domain_keeps_public_axes_domain_named() -> None:
    text = QUARTER.read_text(encoding="utf-8")
    assert "quarter_rule=ten_minute_sanyuan" in text
    assert "quarter_rule=twelve_minute_ten_division" in text
    assert "base_dingju_method" in text
    assert "dun_source" in text
    assert "hour_boundary" in text
    assert "Horosa" not in text


def test_yongshen_selection_stays_table_driven() -> None:
    text = YONGSHEN.read_text(encoding="utf-8")
    assert "表外问事先映射到表内问事" in text
    assert "无法映射时说明无封闭用神规则" in text
    assert "不得按五行临场类象" in text
    assert "按五行类象选择最接近" not in text


def test_yongshen_matter_mapping_is_python_table() -> None:
    text = YONGSHEN.read_text(encoding="utf-8")
    assert "`tools/data/qimen_matters.csv` 是事象到用神的唯一事实源" in text
    assert "本页表格仅作展示" in text


def test_qimen_engine_and_python_layers_are_decoupled() -> None:
    text = YONGSHEN.read_text(encoding="utf-8")
    assert "engine 只接收 `yong_shen`" in text
    assert "`matter` 不进入 `qimen.chart`" in text


def test_lost_property_stays_in_interpretation_layer() -> None:
    text = LOST_PROPERTY.read_text(encoding="utf-8")
    assert "不传 `matter` 或 `yong_shen`" in text
    assert "snapshot.special.assertions" in text
    assert "反吟为复得候选" in text
    assert "时干落空亡为难复得候选" in text
    assert "乘旺相气为复得候选" in text
    assert "`palace_wang_shuai` / `shi_gan_gong_wang_shuai` 输出" in text
    assert "墓、绝未入表" in text
    assert "《奇门遁甲元灵经·卷十八·占失物》补方向与内外远近" in text
    assert "中宫无内外 / 方向" in text


def test_missing_person_stays_conservative_and_interpretation_layer() -> None:
    text = MISSING_PERSON.read_text(encoding="utf-8")
    assert "`qimen_snapshot(matter=missing_person)`" in text
    assert "snapshot.special.assertions" in text
    assert "不排序" in text
    assert "不推出必然回归或必然失踪" in text
    assert "六合落宫星旺 / 相且临景、死、惊、伤四门时" in text
    assert "`qimen_snapshot(matter=missing_person, rule=missing_person)`" in CARD.read_text(
        encoding="utf-8"
    )


def test_capture_escape_stays_conservative_and_interpretation_layer() -> None:
    text = CAPTURE_ESCAPE.read_text(encoding="utf-8")
    assert "不传 `matter` 或 `yong_shen`" in text
    assert "snapshot.special.assertions" in text
    assert "不排序" in text
    assert "不做执法建议或必然结论" in text
    assert "行人年命、官府差人等另占口径未入表" in text
    assert "`qimen_snapshot(rule=thief_capture/capture_escape/thief_profile)`" in CARD.read_text(encoding="utf-8")


def test_thief_capture_documents_geng_branches() -> None:
    text = (ROOT / "skills/liki-divination/domains/qimen/thief-capture.md").read_text(
        encoding="utf-8"
    )
    assert "庚格已入表" in text
    assert "年格、月格、日格、时格" in text
    assert "无庚格输出难获候选" in text
    assert "杜门落宫有庚格输出杜门捕获候选" in text
    assert "当前尚未入表" not in text


def test_thief_profile_is_table_driven_route() -> None:
    text = (ROOT / "skills/liki-divination/domains/qimen/thief-capture.md").read_text(
        encoding="utf-8"
    )
    assert "snapshot.special.assertions" in text
    assert "贵人 / 小人" in text
    assert "内盘亲近人或外盘外人" in text
    assert "`qimen_snapshot(rule=thief_capture/capture_escape/thief_profile)`" in CARD.read_text(
        encoding="utf-8"
    )


def test_yongshen_documents_missing_person_mapping() -> None:
    text = YONGSHEN.read_text(encoding="utf-8")
    assert "| 走失人口 | 六合 |" in text
    assert "《奇门遁甲元灵经·占走失》" in text
    assert "时干宫为失主，六合宫为走失者" in text


def test_pattern_interpretation_stays_table_driven() -> None:
    text = YONGSHEN.read_text(encoding="utf-8")
    assert "`patterns` 只返回格局表已收录且满足同宫条件的格局" in text
    assert "不承诺全量格局识别" in text
    assert "不得临场补格局名" in text


def test_yingqi_dates_come_from_engine() -> None:
    text = (ROOT / "skills/liki-divination/domains/qimen/yingqi.md").read_text(
        encoding="utf-8"
    )
    assert "`candidates[].dates[]`" in text
    assert "日期窗口由引擎计算" in text
    assert "LLM 不得补算日期" in text
