"""领域快照测试：完整 pan 生成 snap，并保留基础因子与领域事实。

断言快照结构自洽（{八字, 紫微, context} + 因子产生），并经 _base_ctx_from_pan 保证 shishen/wuxing 直读。
"""
import _helpers  # noqa: F401 —— 注入 tools 路径
import factors


def _cmp_bazi_pan() -> dict:
    return {
        "chart": {
            "nian": {"gan": "庚", "zhi": "午"}, "yue": {"gan": "壬", "zhi": "午"},
            "ri": {"gan": "甲", "zhi": "子"}, "shi": {"gan": "庚", "zhi": "午"},
            "da_yun": {"steps": [{"name": "甲午", "start_year": 2000, "end_year": 2010, "shi_shen": "偏印"}],
                       "current_step_index": 0},
        },
        "full": {
            "nian": {"gan": "庚", "zhi": "午", "shi_shens": [
                {"shi_shen": "七杀", "gan": "庚", "source": "gan"},
                {"shi_shen": "伤官", "gan": "丁", "source": "main_qi"}],
                "cang_gan": {"main": "丁", "mid": "己"}, "shen_sha": [{"name": "桃花"}]},
            "yue": {"gan": "壬", "zhi": "午", "shi_shens": [
                {"shi_shen": "偏印", "gan": "壬", "source": "gan"},
                {"shi_shen": "伤官", "gan": "丁", "source": "main_qi"}],
                "cang_gan": {"main": "丁", "mid": "己"}},
            "ri": {"gan": "甲", "zhi": "子", "shi_shens": [], "cang_gan": {"main": "癸"}},
            "shi": {"gan": "庚", "zhi": "午", "shi_shens": [{"shi_shen": "七杀", "gan": "庚", "source": "gan"}],
                    "cang_gan": {"main": "丁"}},
            "gan_he": [{"gan_a": "甲", "gan_b": "己"}],
            "zhi_liu_he": [], "liu_chong": [{"zhi_a": "子", "zhi_b": "午"}],
            "liu_hai": [], "san_he": [], "san_hui": [], "liu_xing": [],
            "chang_sheng": [{"index": "子", "name": "沐浴"}],
        },
        "yongshen": {"fu_yi": {"wuxing_count": {"木": 1, "火": 2}, "wang_shuai": {"木": "旺", "火": "相", "土": "死", "金": "囚", "水": "休"},
            "yong": "木", "xi": "水", "ji": "金", "qiangruo": "身强"}, "tiao_hou": {"season": "冬"}, "ge_ju": {"ge_ju": "正印格"}},
        "ziwei": {"gong_wei": [{"index": "命宫", "name": "命宫", "xing_yao": [{"name": "七杀", "xing": "七杀", "is_major": True}]}],
            "si_hua": {"七杀": "禄"}, "patterns": [{"name": "杀破狼"}]},
        "gender": "male",
    }


def test_evaluate_snap_from_pan_produces_complete_snap():
    pan = _cmp_bazi_pan()
    new = factors.evaluate_snap_from_pan(pan)
    # 结构：八字/紫微 双盘因子快照 + context
    assert "八字" in new and "紫微" in new and "context" in new
    assert new["context"] == {"性别": "male"}
    # 因子被产出（八字/紫微各有若干因子，非空）
    assert len(new["八字"]) > 0
    assert len(new["紫微"]) > 0
    # 基础字段已直读：宫含命宫(=含紫微) 应在 snap 被解释为具体因子，此处验证 shishen 直读正常
    ctx = factors._base_ctx_from_pan(pan)
    assert ctx["ri_gan"] == "甲"
    assert ctx["palace_ri"] == {"zhi": "子"}
    assert "七杀" in ctx["shishen"]  # 年干庚=甲日主之七杀(金克木)，应聚合出该十神键


def test_snap_embeds_stable_domain_facts():
    """P1/领域驱动：pan 的领域事实应完整透传进 snap（藏干/大运/宫位/solar lunar 等），供断语与未来扩展。"""
    pan = _cmp_bazi_pan()
    pan["full"]["nian"]["cang_gan"] = {"main": "丁", "mid": "己"}
    pan["full"]["yue"]["cang_gan"] = {"main": "丁", "mid": "己"}
    pan["full"]["ri"]["cang_gan"] = {"main": "癸"}
    pan["full"]["shi"]["cang_gan"] = {"main": "丁"}
    pan["full"]["ri"]["na_yin"] = "海中金"
    pan["full"]["san_yuan"] = {"胎元": "丙子", "命宫": "甲午", "身宫": "甲午"}
    pan["full"]["xun_kong"] = "甲申旬空午未"
    pan["full"]["san_qi_name"] = "天乙贵人"
    pan["chart"]["da_yun"] = {"direction": "顺排", "current_step_index": 0,
        "steps": [{"name": "甲午", "gan": "甲", "zhi": "午", "wuxing": "木", "shi_shen": "偏印",
                   "start_year": 2000, "end_year": 2010}]}
    pan["ziwei"]["ju_shu"] = "火六局"
    pan["ziwei"]["ming_zhu"] = "贪狼"
    pan["ziwei"]["shen_zhu"] = "天梁"
    pan["ziwei"]["kong_gong"] = [{"gong_name": "兄弟", "jie_xing": ["天机"]}]
    pan["ziwei"]["nian_gan"] = "庚"
    pan["ziwei"]["ziwei_pos"] = "命宫"
    pan["solar"] = "1990-05-20T12:00:00"
    pan["lunar"] = "1990年四月廿六"

    snap = factors.evaluate_snap_from_pan(pan)
    bz = snap["八字"]; zw = snap["紫微"]; ctx = snap["context"]
    assert bz["日主"]
    assert bz["ri柱纳音"] == "海中金"
    assert bz["ri柱藏干"] == {"main": "癸"}
    assert bz["nian柱藏干"] == {"main": "丁", "mid": "己"}
    assert bz["三元"] == {"胎元": "丙子", "命宫": "甲午", "身宫": "甲午"}
    assert bz["旬空"] == "甲申旬空午未"
    assert bz["三奇贵人"] == "天乙贵人"
    # 大运领域结构完整透传
    assert bz["大运"]["direction"] == "顺排"
    assert bz["大运"]["steps"][0]["gan"] == "甲"
    assert bz["大运"]["steps"][0]["wuxing"] == "木"
    # 紫微：宫位结构 + 领域字段
    assert zw["宫位"][0]["xing_yao"][0]["xing"] == "七杀"
    assert zw["局数"] == "火六局"
    assert zw["命主"] == "贪狼"
    assert zw["身主"] == "天梁"
    assert zw["年干"] == "庚"
    assert zw["紫微星位"] == "命宫"
    assert zw["空宫"][0]["gong_name"] == "兄弟"
    # 出生时间进 context
    assert ctx["公历出生"] == "1990-05-20T12:00:00"
    assert ctx["农历出生"] == "1990年四月廿六"
