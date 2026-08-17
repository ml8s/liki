"""
六爻特殊格局规则测试

测试 domains/liuyao/special.md 中的规则逻辑。
TDD 模式：先写测试，再实现规则。
"""
import pytest
from test_data import SPECIAL_PATTERN_TESTS


# Mock 类：模拟卦象数据
class MockChart:
    def __init__(self, **kwargs):
        self.yongshen_wangshuai = kwargs.get("yongshen_wangshuai", "旺")
        self.xun_kong = kwargs.get("xun_kong", False)
        self.dong = kwargs.get("dong", False)
        self.yongshen_yuepo = kwargs.get("yongshen_yuepo", False)
        self.de_sheng = kwargs.get("de_sheng", False)
        self.yongshen_not_on_gua = kwargs.get("yongshen_not_on_gua", False)
        self.fe_sheng = kwargs.get("fe_sheng", False)
        self.yongshen_dong = kwargs.get("yongshen_dong", False)
        self.hua_jin = kwargs.get("hua_jin", False)
        self.hua_tui = kwargs.get("hua_tui", False)
        self.liu_chong = kwargs.get("liu_chong", False)
        self.liu_he = kwargs.get("liu_he", False)
        self.fan_yin = kwargs.get("fan_yin", False)
        self.fu_yin = kwargs.get("fu_yin", False)
        self.sui_gui_ru_mu = kwargs.get("sui_gui_ru_mu", False)
        self.du_fa = kwargs.get("du_fa", False)
        self.du_jing = kwargs.get("du_jing", False)
        self.yongshen_liang_xian = kwargs.get("yongshen_liang_xian", False)


# 旬空测试
class TestXunKong:
    def test_vacant_dong_wang(self):
        """旺相动爻旬空=假空"""
        chart = MockChart(yongshen_wangshuai="旺", xun_kong=True, dong=True)
        # 实现后调用：result = evaluate_xunkong(chart)
        # assert result.xunkong_type == "假空"
        # assert "迟成而非不成" in result.assessment
        pytest.skip("待实现")

    def test_vacant_static_xiu(self):
        """休囚静爻旬空=真空"""
        chart = MockChart(yongshen_wangshuai="休", xun_kong=True, dong=False)
        # 实现后调用：result = evaluate_xunkong(chart)
        # assert result.xunkong_type == "真空"
        # assert "事不实" in result.assessment
        pytest.skip("待实现")


# 月破测试
class TestYuePo:
    def test_dong_po_de_sheng(self):
        """动爻月破+得生=假破"""
        chart = MockChart(yongshen_yuepo=True, dong=True, de_sheng=True)
        # 实现后调用：result = evaluate_yuepo(chart)
        # assert result.yuepo_type == "假破"
        # assert "先挫后成" in result.assessment
        pytest.skip("待实现")

    def test_xiu_po_no_sheng(self):
        """休囚月破=真破"""
        chart = MockChart(yongshen_yuepo=True, dong=False, de_sheng=False)
        # 实现后调用：result = evaluate_yuepo(chart)
        # assert result.yuepo_type == "真破"
        # assert "当下无力" in result.assessment
        pytest.skip("待实现")


# 飞伏测试
class TestFeiFu:
    def test_not_on_gua_chong_fei(self):
        """用神伏藏+冲飞=出伏"""
        chart = MockChart(yongshen_not_on_gua=True, fe_sheng=True)
        # 实现后调用：result = evaluate_feifu(chart)
        # assert result.feifu_type == "冲飞起伏"
        # assert "冲飞出伏" in result.assessment
        pytest.skip("待实现")


# 进退神测试
class TestJinTui:
    def test_hua_jin(self):
        """用神化进神"""
        chart = MockChart(yongshen_dong=True, hua_jin=True)
        # 实现后调用：result = evaluate_jintui(chart)
        # assert result.jinshen_type == "进神"
        # assert "力量增长" in result.assessment
        pytest.skip("待实现")

    def test_hua_tui(self):
        """用神化退神"""
        chart = MockChart(yongshen_dong=True, hua_tui=True)
        # 实现后调用：result = evaluate_jintui(chart)
        # assert result.jinshen_type == "退神"
        # assert "力量衰败" in result.assessment
        pytest.skip("待实现")


# 六冲/六合测试
class TestChongHe:
    def test_liu_chong(self):
        """六冲卦"""
        chart = MockChart(liu_chong=True)
        # 实现后调用：result = evaluate_chonghe(chart)
        # assert result.chonghe_type == "六冲"
        # assert "冲散与反复" in result.assessment
        pytest.skip("待实现")

    def test_liu_he(self):
        """六合卦"""
        chart = MockChart(liu_he=True)
        # 实现后调用：result = evaluate_chonghe(chart)
        # assert result.chonghe_type == "六合"
        # assert "稳定与成局" in result.assessment
        pytest.skip("待实现")


# 反吟/伏吟测试
class TestFanYin:
    def test_fan_yin(self):
        """反吟卦"""
        chart = MockChart(fan_yin=True)
        # 实现后调用：result = evaluate_fanyin(chart)
        # assert result.fanyin_type == "反吟"
        # assert "反复与重来" in result.assessment
        pytest.skip("待实现")

    def test_fu_yin(self):
        """伏吟卦"""
        chart = MockChart(fu_yin=True)
        # 实现后调用：result = evaluate_fanyin(chart)
        # assert result.fanyin_type == "伏吟"
        # assert "停滞与反复" in result.assessment
        pytest.skip("待实现")


# 随鬼入墓测试
class TestSuiGuiRuMu:
    def test_sui_gui_ru_mu(self):
        """随鬼入墓"""
        chart = MockChart(sui_gui_ru_mu=True)
        # 实现后调用：result = evaluate_suiguirumu(chart)
        # assert result.suiguirumu_type == "随鬼入墓"
        # assert "闭塞与延迟" in result.assessment
        pytest.skip("待实现")


# 独发/独静测试
class TestDuFaDuJing:
    def test_du_fa(self):
        """独发"""
        chart = MockChart(du_fa=True)
        # 实现后调用：result = evaluate_dufadujing(chart)
        # assert result.dufadujing_type == "独发"
        # assert "应期校准器" in result.assessment
        pytest.skip("待实现")

    def test_du_jing(self):
        """独静"""
        chart = MockChart(du_jing=True)
        # 实现后调用：result = evaluate_dufadujing(chart)
        # assert result.dufadujing_type == "独静"
        # assert "应期校准器" in result.assessment
        pytest.skip("待实现")


# 用神两现测试
class TestYongShenLiangXian:
    def test_liang_xian(self):
        """用神两现"""
        chart = MockChart(yongshen_liang_xian=True)
        # 实现后调用：result = evaluate_liangxian(chart)
        # assert result.liangxian_type == "用神两现"
        # assert "取旺不取衰" in result.assessment
        pytest.skip("待实现")


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
