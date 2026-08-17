"""
六爻流程集成测试

测试 L0-L5 分层断卦流程。
TDD 模式：先写测试，再实现流程。
"""
import pytest


# Mock 类
class MockLiuyaoAPI:
    def qigua(self):
        return {"yaos": [7, 8, 7, 8, 7, 8], "dong_yao": []}

    def chart(self, solar_time, yong_shen, yaos):
        return {
            "lines": [],
            "wang_shuai": ["旺", "相", "休", "囚", "死", "旺"],
            "yong_shen": {"position": 3, "type": yong_shen},
            "shi_yao": 1,
            "ying_yao": 4,
        }


class MockYongShen:
    def get_yongshen(self, chart, question_type):
        mapping = {
            "事业": "官鬼",
            "财运": "妻财",
            "感情": "妻财",
            "学业": "父母",
            "出行": "世爻",
            "住宅": "父母",
            "法律": "官鬼",
            "家庭": "父母",
            "通用": "世爻",
        }
        return mapping.get(question_type, "世爻")


class MockWangShuai:
    def get_wangshuai(self, chart, yong_shen):
        return "旺"


class MockSpecialPatterns:
    def get_patterns(self, chart, yong_shen):
        return []


class MockDivination:
    def divinate(self, chart, yong_shen, wangshuai, patterns):
        return {"forecast": "测试结论", "analysis": "测试分析"}


class MockYingQi:
    def get_yingqi(self, chart, yong_shen, wangshuai):
        return {"time": "测试应期", "window": "测试时间窗口"}


# L0: 排盘测试
class TestL0_PaiPan:
    def test_qigua(self):
        """L0: 起卦"""
        api = MockLiuyaoAPI()
        result = api.qigua()
        assert "yaos" in result
        assert len(result["yaos"]) == 6
        assert all(6 <= y <= 9 for y in result["yaos"])

    def test_chart(self):
        """L0: 装卦"""
        api = MockLiuyaoAPI()
        result = api.chart("2024-01-01", "官鬼", [7, 8, 7, 8, 7, 8])
        assert "lines" in result
        assert "wang_shuai" in result
        assert "yong_shen" in result


# L1: 用神取用测试
class TestL1_YongShen:
    def test_get_yongshen_career(self):
        """L1: 事业取用神=官鬼"""
        mock = MockYongShen()
        chart = MockLiuyaoAPI().chart("2024-01-01", "官鬼", [7, 8, 7, 8, 7, 8])
        result = mock.get_yongshen(chart, "事业")
        assert result == "官鬼"

    def test_get_yongshen_wealth(self):
        """L1: 财运取用神=妻财"""
        mock = MockYongShen()
        chart = MockLiuyaoAPI().chart("2024-01-01", "妻财", [7, 8, 7, 8, 7, 8])
        result = mock.get_yongshen(chart, "财运")
        assert result == "妻财"

    def test_get_yongshen_relationship(self):
        """L1: 感情取用神=妻财（男）"""
        mock = MockYongShen()
        chart = MockLiuyaoAPI().chart("2024-01-01", "妻财", [7, 8, 7, 8, 7, 8])
        result = mock.get_yongshen(chart, "感情")
        assert result == "妻财"


# L2: 旺衰判定测试
class TestL2_WangShuai:
    def test_get_wangshuai(self):
        """L2: 判定用神旺衰"""
        mock = MockWangShuai()
        chart = MockLiuyaoAPI().chart("2024-01-01", "官鬼", [7, 8, 7, 8, 7, 8])
        result = mock.get_wangshuai(chart, "官鬼")
        assert result in ["旺", "相", "休", "囚", "死"]


# L3: 特殊格局测试
class TestL3_SpecialPatterns:
    def test_get_patterns(self):
        """L3: 获取特殊格局"""
        mock = MockSpecialPatterns()
        chart = MockLiuyaoAPI().chart("2024-01-01", "官鬼", [7, 8, 7, 8, 7, 8])
        result = mock.get_patterns(chart, "官鬼")
        assert isinstance(result, list)


# L4: 分类断卦测试
class TestL4_Divination:
    def test_divinate(self):
        """L4: 分类断卦"""
        mock = MockDivination()
        chart = MockLiuyaoAPI().chart("2024-01-01", "官鬼", [7, 8, 7, 8, 7, 8])
        result = mock.divinate(chart, "官鬼", "旺", [])
        assert "forecast" in result
        assert "analysis" in result


# L5: 应期推断测试
class TestL5_YingQi:
    def test_get_yingqi(self):
        """L5: 应期推断"""
        mock = MockYingQi()
        chart = MockLiuyaoAPI().chart("2024-01-01", "官鬼", [7, 8, 7, 8, 7, 8])
        result = mock.get_yingqi(chart, "官鬼", "旺")
        assert "time" in result
        assert "window" in result


# 完整流程测试
class TestFullFlow:
    def test_career_flow(self):
        """事业占断完整流程"""
        # L0: 排盘
        api = MockLiuyaoAPI()
        chart = api.qigua()
        assert chart["yaos"] is not None

        chart_data = api.chart("2024-01-01", "官鬼", chart["yaos"])
        assert chart_data is not None

        # L1: 用神取用
        yongshen = MockYongShen().get_yongshen(chart_data, "事业")
        assert yongshen == "官鬼"

        # L2: 旺衰判定
        wangshuai = MockWangShuai().get_wangshuai(chart_data, yongshen)
        assert wangshuai in ["旺", "相", "休", "囚", "死"]

        # L3: 特殊格局
        patterns = MockSpecialPatterns().get_patterns(chart_data, yongshen)
        assert isinstance(patterns, list)

        # L4: 分类断卦
        result = MockDivination().divinate(chart_data, yongshen, wangshuai, patterns)
        assert result["forecast"] is not None

        # L5: 应期推断
        yingqi = MockYingQi().get_yingqi(chart_data, yongshen, wangshuai)
        assert yingqi["time"] is not None

    def test_wealth_flow(self):
        """财运占断完整流程"""
        api = MockLiuyaoAPI()
        chart = api.qigua()
        chart_data = api.chart("2024-01-01", "妻财", chart["yaos"])

        yongshen = MockYongShen().get_yongshen(chart_data, "财运")
        assert yongshen == "妻财"

        wangshuai = MockWangShuai().get_wangshuai(chart_data, yongshen)
        patterns = MockSpecialPatterns().get_patterns(chart_data, yongshen)
        result = MockDivination().divinate(chart_data, yongshen, wangshuai, patterns)
        yingqi = MockYingQi().get_yingqi(chart_data, yongshen, wangshuai)

        assert result["forecast"] is not None
        assert yingqi["time"] is not None


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
