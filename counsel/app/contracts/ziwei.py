"""Typed wire model for the canonical engine Ziwei chart."""


from typing import Annotated, Literal

from pydantic import BaseModel, ConfigDict, Field

from app.contracts.bazi import Gan, Gender, Zhi

Palace = Literal[
    "命宫", "兄弟", "夫妻", "子女", "财帛", "疾厄",
    "迁移", "仆役", "官禄", "田宅", "福德", "父母",
]
SiHua = Literal["禄", "权", "科", "忌"]
JuShuCode = Literal["water_2", "wood_3", "metal_4", "earth_5", "fire_6"]
StarLei = Literal["zhu_xing", "fu_xing", "liu_ji", "wen_xing", "sha_xing"]


class ZiweiKouJing(BaseModel):
    model_config = ConfigDict(extra="forbid")
    leap_month: Literal["first_fifteen_days_current_month_next_fifteen_days_following_month"] = Field(
        description="闰月取月口径。",
    )
    source: Literal["iztro_compatible"] = Field(description="排盘口径来源。")


class ZiweiStar(BaseModel):
    model_config = ConfigDict(extra="forbid")
    ming: Annotated[str, Field(min_length=1)] = Field(description="星曜名。")
    lei: Annotated[list[StarLei], Field(min_length=1)] = Field(description="星曜类别。")
    si_hua: SiHua | None = Field(default=None, description="四化。")
    liang_du: str | None = Field(default=None, description="亮度。")


class ZiweiGong(BaseModel):
    model_config = ConfigDict(extra="forbid")
    ming: Palace = Field(description="宫名。")
    gan: Gan = Field(description="宫干。")
    zhi: Zhi = Field(description="宫支。")
    shen_gong: bool | None = Field(default=None, description="是否身宫。")
    lai_yin_gong: bool | None = Field(default=None, description="是否来因宫。")
    xing_yao: Annotated[list[ZiweiStar], Field(min_length=0)] = Field(description="星曜。")
    xiao_xian: list[int] | None = Field(default=None, description="小限年龄。")
    chang_sheng: str | None = Field(default=None, description="长生十二宫。")
    bo_shi: str | None = Field(default=None, description="博士十二神。")
    jiang_qian: str | None = Field(default=None, description="将前十二神。")
    sui_qian: str | None = Field(default=None, description="岁前十二神。")
    za_yao: list[str] | None = Field(default=None, description="杂曜。")


class ZiweiKongGongJieXing(BaseModel):
    model_config = ConfigDict(extra="forbid")
    kong_gong: Palace = Field(description="无主星宫。")
    jie_gong: Palace = Field(description="借星来源宫。")


class ZiweiGeJu(BaseModel):
    model_config = ConfigDict(extra="forbid")
    name: Annotated[str, Field(min_length=1)] = Field(description="格局名。")
    description: str | None = Field(default=None, description="结构说明。")


class ZiweiSanFangSiZheng(BaseModel):
    model_config = ConfigDict(extra="forbid")
    ming_gong: Annotated[list[Palace], Field(min_length=4, max_length=4)] = Field(
        description="命宫三方四正宫位。",
    )


class ZiweiChart(BaseModel):
    """Canonical Ziwei chart; palaces are the only owner of star facts."""

    model_config = ConfigDict(extra="forbid")
    kou_jing: ZiweiKouJing = Field(description="排盘口径。")
    shi_er_gong: Annotated[list[ZiweiGong], Field(min_length=12, max_length=12)] = Field(
        description="从命宫顺布的十二宫。",
    )
    ju_shu: JuShuCode = Field(description="局数机器码。")
    ju_shu_name: str = Field(description="局数名。")
    kong_gong: list[ZiweiKongGongJieXing] | None = Field(default=None, description="空宫借星。")
    nian_gan: Gan = Field(description="年干。")
    nian_zhi: Zhi | None = Field(default=None, description="年支。")
    shi_zhi: Zhi = Field(description="时支。")
    birth_year: Annotated[int, Field(ge=1)] = Field(description="出生公历年。")
    lunar_month: int | None = Field(default=None, description="排盘调整农历月。")
    lunar_day: int | None = Field(default=None, description="排盘农历日。")
    birth_lunar_month: int | None = Field(default=None, description="出生原始农历月。")
    birth_is_leap: bool | None = Field(default=None, description="出生农历月是否闰月。")
    gender: Gender = Field(description="性别。")
    ming_zhu: str | None = Field(default=None, description="命主。")
    shen_zhu: str | None = Field(default=None, description="身主。")
    ge_ju: list[ZiweiGeJu] | None = Field(default=None, description="格局候选。")
    san_fang_si_zheng: ZiweiSanFangSiZheng = Field(description="三方四正宫位关系。")


__all__ = [
    "JuShuCode",
    "Palace",
    "SiHua",
    "StarLei",
    "ZiweiChart",
    "ZiweiGong",
    "ZiweiKongGongJieXing",
    "ZiweiKouJing",
    "ZiweiSanFangSiZheng",
    "ZiweiStar",
]
