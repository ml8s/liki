"""本命事实投影契约：reserved domain facts 的结构与只读投影。"""

from app.natal.tools.natal_projection import load_contract, project_natal_facts


def _rich_pan() -> dict:
    return {
        "bazi": {
            pillar: {
                "gan": "甲", "zhi": "子",
                "na_yin": f"{pillar}纳音", "cang_gan": {"main": "甲"},
                "shi_shens": [{"shi_shen": "比肩", "gan": "甲", "source": "gan"}],
                "self_sitting": "沐浴", "shen_sha": [{"name": "天乙贵人"}],
                "day_master_trend": f"{pillar}地势",
                "xun": f"{pillar}旬", "xun_kong": f"{pillar}旬空支",
                "is_void": False, "is_self_he": True,
                "is_kui_gang": False, "self_he_name": "自合",
            } for pillar in ("nian", "yue", "ri", "shi")
        } | {
            "birth_year": 1990,
            "_yong_shen_wrapper": {
                "fu_yi": {"fu_yi_model_id": "support_control_with_day_master_strength"},
                "tiao_hou": {"tiao_hou_model_id": "qiongtong_primary_secondary_table"},
                "ge_ju": {"model": "ziping_month_command_candidate"},
            },
            "chang_sheng": [{"pillar": "nian", "stage": "长生"}],
            "zodiac": "鼠",
            "san_yuan": {"胎元": "丙子"}, "tai_xi": "辛丑",
            "shen_sha_school": {
                "dual_reference": ["天乙贵人", "桃花"],
                "policy": "union_of_year_and_day_references",
            },
            "lu_roots": [{"shi_shen": "正印", "gan": "癸", "zhi": "子"}],
            "guan_xi": [
                {"relation_kind": "gan_he", "group": "甲己", "members": [
                    {"pillar": "nian", "gan": "甲"}, {"pillar": "ri", "gan": "己"}
                ]},
                {"relation_kind": "zhi_liu_he", "group": "子丑", "members": [
                    {"pillar": "nian", "zhi": "子"}, {"pillar": "shi", "zhi": "丑"}
                ]},
            ],
            "ten_god_states": [{"shi_shen": "正印", "wuxing": "水", "transparent": True, "hidden": False, "count": 1, "rooted": True, "timely": True, "strength": "strong"}],
            "element_states": [{"wuxing": "水", "season": "旺", "season_strength": "strong", "strength": "strong", "transparent": True, "rooted": True, "controls": "火", "controlled_by": "土", "controller_strength": "weak", "reasons": ["得令"]}],
            "atomic_facts": {
                "day_master_element": "木", "day_master_stem": "甲", "day_branch": "子",
                "month_longevity": "沐浴", "pattern_god_transparent": False,
            },
            "fu_yi": {"qiang_ruo": "身强", "yong": "木", "xi": "水", "ji": "金", "wuxing_count": {}, "wang_shuai": {}, "fu_yi_model_id": "test", "basis": {}},
            "tiao_hou": {"primary_wuxing": "火", "season": "春", "tiao_hou_model_id": "qiongtong_primary_secondary_table", "primary": {"stem": "丙", "transparent": True, "hidden": False, "occurrences": [], "relation_facts": []}},
            "ge_ju": {"ge_ju": "正印格", "yong_fa": "顺用", "pattern_god_source": "main_qi", "structure": {}},
            "guan_sha": {"qu_qing": False}, "cai_ku": {"xian": False, "xing_ru_ku": False},
            "zi_shi_rule": "晚子时换日口径",
            "ri_xun": "甲戌旬", "ri_xun_kong": "申酉",
            "san_qi_name": "三奇", "gong_jia": ["拱夹"], "na_yin_rel": [{"rel": "相生"}],
            "da_yun": {"steps": [], "current_step_index": 0},
        },
        "chart": {"birth_year": 1990},
        "ziwei": {
            "shi_er_gong": [{"ming": "命宫", "xing_yao": [{"ming": "七杀"}]}],
            "ju_shu": "fire_6", "ju_shu_name": "火六局",
            "kou_jing": {"leap_month": "iztro_compatible", "source": "test"},
            "kong_gong": [{"kong_gong": "兄弟", "jie_gong": "子女"}],
            "ming_zhu": "贪狼", "shen_zhu": "天梁", "ming_gong": "命宫", "shen_gong": "身宫",
            "si_hua": {"七杀": "禄"}, "ge_ju": [{"name": "七杀朝斗"}],
            "san_fang_si_zheng": {"ming_gong": ["命宫", "财帛", "官禄", "迁移"]},
            "nian_gan": "庚", "nian_zhi": "午",
            "shi_zhi": "子", "ziwei_pos": "命宫", "birth_year": 1990,
            "lunar_month": 4, "lunar_day": 26, "birth_lunar_month": 4,
            "birth_is_leap": False, "gender": "male",
        },
        "ziwei_daxian": [{"gong": "命宫", "name": "甲子", "start_year": 1990, "end_year": 1999, "qi_sui": 1, "zhi_sui": 10}],
    }


def test_projected_domain_facts_match_contract():
    facts = project_natal_facts(_rich_pan())
    contract = load_contract()
    assert set(facts["八字"]) == set(contract["八字"])
    assert set(facts["紫微"]) == set(contract["紫微"])


def test_empty_pan_projects_empty_domain_facts():
    assert project_natal_facts({}) == {"八字": {}, "紫微": {}}


def test_domain_projection_does_not_mutate_pan():
    natal_chart = _rich_pan()
    project_natal_facts(natal_chart)
    assert "_snap" not in natal_chart and "_ctx" not in natal_chart


def test_fields_are_declared_when_present():
    facts = project_natal_facts(_rich_pan())
    assert facts["八字"]["大运"]["current_step_index"] == 0
    assert facts["八字"]["tiao_hou"]["tiao_hou_model_id"] == "qiongtong_primary_secondary_table"
    assert facts["八字"]["命局关系集合"][0]["group"] == "甲己"
    assert "命理原子事实" not in facts["八字"]
    assert facts["紫微"]["局数"] == "fire_6"
    assert facts["紫微"]["十二宫"][0]["ming"] == "命宫"
    assert facts["紫微"]["格局候选"][0]["name"] == "七杀朝斗"
    assert facts["八字"]["十二长生"][0]["pillar"] == "nian"
