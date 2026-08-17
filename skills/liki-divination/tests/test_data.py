# 六爻测试数据

# 卦例数据格式说明
# - id: 卦例编号
# - source: 来源（增删卜易 卷/章）
# - question: 占事类型
# - time: 起卦时间（农历月日）
# - main_hexagram: 主卦名
# - bian_hexagram: 变卦名
# - yaos: 六爻值（初爻到上爻，6/7/8/9）
# - expected: 预期结果

## 事业类卦例

CAREER_EXAMPLES = [
    {
        "id": "C001",
        "source": "《增删卜易》卷三",
        "question": "升迁",
        "time": "卯月丙寅日",
        "main_hexagram": "解",
        "bian_hexagram": "困",
        "yaos": [7, 8, 7, 8, 7, 8],
        "dong_yao": [1, 3, 5],
        "expected": {
            "yongshen": "官鬼",
            "wangshuai": "旺",
            "forecast": "可升",
            "yingqi": "应巳月"
        }
    },
    {
        "id": "C002",
        "source": "《增删卜易》卷七",
        "question": "求职",
        "time": "午月丁未日",
        "main_hexagram": "大壮",
        "bian_hexagram": "豫",
        "yaos": [9, 8, 7, 8, 9, 7],
        "dong_yao": [1, 5],
        "expected": {
            "yongshen": "官鬼",
            "wangshuai": "相",
            "forecast": "有机遇",
            "yingqi": "未日"
        }
    },
]

## 财运类卦例

WEALTH_EXAMPLES = [
    {
        "id": "W001",
        "source": "《增删卜易》卷九",
        "question": "开金银器皿铺",
        "time": "未月辛丑日",
        "main_hexagram": "火雷噬嗑",
        "bian_hexagram": "屯",
        "yaos": [8, 9, 7, 8, 9, 8],
        "dong_yao": [2, 5],
        "expected": {
            "yongshen": "妻财",
            "wangshuai": "旺",
            "forecast": "财源广进",
            "yingqi": "戌日开张"
        }
    },
    {
        "id": "W002",
        "source": "《增删卜易》卷九",
        "question": "求财",
        "time": "寅月庚戌日",
        "main_hexagram": "火水未济",
        "bian_hexagram": "山水蒙",
        "yaos": [9, 8, 7, 9, 8, 7],
        "dong_yao": [1, 4],
        "expected": {
            "yongshen": "妻财",
            "wangshuai": "休",
            "forecast": "待时",
            "yingqi": "寅日"
        }
    },
]

## 感情类卦例

RELATIONSHIP_EXAMPLES = [
    {
        "id": "R001",
        "source": "《增删卜易》卷十",
        "question": "婚姻",
        "time": "寅月辛卯日",
        "main_hexagram": "乾",
        "bian_hexagram": "离",
        "yaos": [7, 8, 7, 8, 7, 8],
        "dong_yao": [],
        "expected": {
            "yongshen": "妻财",
            "wangshuai": "旺",
            "forecast": "婚期将近",
            "yingqi": "卯月"
        }
    },
    {
        "id": "R002",
        "source": "《增删卜易》卷十",
        "question": "复合",
        "time": "午月丙子日",
        "main_hexagram": "大壮",
        "bian_hexagram": "豫",
        "yaos": [9, 8, 7, 8, 9, 7],
        "dong_yao": [1, 5],
        "expected": {
            "yongshen": "妻财",
            "wangshuai": "平",
            "forecast": "有反复",
            "yingqi": "子日"
        }
    },
]

## 学业类卦例

ACADEMIC_EXAMPLES = [
    {
        "id": "A001",
        "source": "《增删卜易》卷七",
        "question": "考试",
        "time": "申月癸巳日",
        "main_hexagram": "姤",
        "bian_hexagram": "中孚",
        "yaos": [9, 8, 7, 8, 9, 7],
        "dong_yao": [1, 5],
        "expected": {
            "yongshen": "父母",
            "wangshuai": "旺",
            "forecast": "成绩可期",
            "yingqi": "申日"
        }
    },
]

## 出行类卦例

TRAVEL_EXAMPLES = [
    {
        "id": "T001",
        "source": "《增删卜易》卷十",
        "question": "出行",
        "time": "卯月丙寅日",
        "main_hexagram": "坎",
        "bian_hexagram": "坤",
        "yaos": [8, 7, 9, 8, 7, 9],
        "dong_yao": [3, 6],
        "expected": {
            "yongshen": "世爻",
            "wangshuai": "旺",
            "forecast": "出行顺利",
            "yingqi": "寅日"
        }
    },
]

## 住宅类卦例

HOME_EXAMPLES = [
    {
        "id": "H001",
        "source": "《增删卜易》卷十二",
        "question": "买房",
        "time": "戌月壬申日",
        "main_hexagram": "艮",
        "bian_hexagram": "坤",
        "yaos": [8, 9, 7, 8, 7, 9],
        "dong_yao": [2, 6],
        "expected": {
            "yongshen": "父母",
            "wangshuai": "旺",
            "forecast": "房屋可买",
            "yingqi": "申日"
        }
    },
]

## 法律类卦例

LEGAL_EXAMPLES = [
    {
        "id": "L001",
        "source": "《增删卜易》卷八",
        "question": "诉讼",
        "time": "巳月戊辰日",
        "main_hexagram": "讼",
        "bian_hexagram": "否",
        "yaos": [7, 8, 9, 7, 8, 9],
        "dong_yao": [3, 6],
        "expected": {
            "yongshen": "官鬼",
            "wangshuai": "平",
            "forecast": "有转机",
            "yingqi": "辰日"
        }
    },
]

## 家庭类卦例

FAMILY_EXAMPLES = [
    {
        "id": "F001",
        "source": "《增删卜易》卷一",
        "question": "父母病",
        "time": "辰月丙申日",
        "main_hexagram": "既济",
        "bian_hexagram": "革",
        "yaos": [7, 8, 9, 7, 8, 7],
        "dong_yao": [3],
        "expected": {
            "yongshen": "父母",
            "wangshuai": "旺",
            "forecast": "有救",
            "yingqi": "酉时"
        }
    },
]

## 特殊格局测试数据

SPECIAL_PATTERN_TESTS = {
    "xunkong": [
        {
            "name": "旺相动爻旬空=假空",
            "yongshen_wangshuai": "旺",
            "xun_kong": True,
            "dong": True,
            "expected_type": "假空",
            "expected_assessment": "旺相动爻旬空，迟成而非不成"
        },
        {
            "name": "休囚静爻旬空=真空",
            "yongshen_wangshuai": "休",
            "xun_kong": True,
            "dong": False,
            "expected_type": "真空",
            "expected_assessment": "休囚静爻旬空，事不实"
        },
    ],
    "yuepo": [
        {
            "name": "动爻月破+得生=假破",
            "yongshen_yuepo": True,
            "dong": True,
            "de_sheng": True,
            "expected_type": "假破",
            "expected_assessment": "动爻月破但得生，先挫后成"
        },
        {
            "name": "休囚月破=真破",
            "yongshen_yuepo": True,
            "dong": False,
            "de_sheng": False,
            "expected_type": "真破",
            "expected_assessment": "休囚月破，当下无力"
        },
    ],
    "feifu": [
        {
            "name": "用神伏藏+冲飞=出伏",
            "yongshen_not_on_gua": True,
            "fe_sheng": True,
            "expected_type": "冲飞起伏",
            "expected_assessment": "冲飞出伏，待时可成"
        },
    ],
    "jinshen": [
        {
            "name": "用神化进神",
            "yongshen_dong": True,
            "hua_jin": True,
            "expected_type": "进神",
            "expected_assessment": "力量增长"
        },
        {
            "name": "用神化退神",
            "yongshen_dong": True,
            "hua_tui": True,
            "expected_type": "退神",
            "expected_assessment": "力量衰败"
        },
    ],
    "chonghe": [
        {
            "name": "六冲卦",
            "liu_chong": True,
            "expected_type": "六冲",
            "expected_assessment": "冲散与反复"
        },
        {
            "name": "六合卦",
            "liu_he": True,
            "expected_type": "六合",
            "expected_assessment": "稳定与成局"
        },
    ],
    "fanyin": [
        {
            "name": "反吟卦",
            "fan_yin": True,
            "expected_type": "反吟",
            "expected_assessment": "反复与重来"
        },
        {
            "name": "伏吟卦",
            "fu_yin": True,
            "expected_type": "伏吟",
            "expected_assessment": "停滞与反复"
        },
    ],
    "suigui": [
        {
            "name": "随鬼入墓",
            "sui_gui_ru_mu": True,
            "expected_type": "随鬼入墓",
            "expected_assessment": "闭塞与延迟"
        },
    ],
    "dufa": [
        {
            "name": "独发",
            "du_fa": True,
            "expected_type": "独发",
            "expected_assessment": "应期校准器"
        },
        {
            "name": "独静",
            "du_jing": True,
            "expected_type": "独静",
            "expected_assessment": "应期校准器"
        },
    ],
    "liangxian": [
        {
            "name": "用神两现",
            "yongshen_liang_xian": True,
            "expected_type": "用神两现",
            "expected_assessment": "取旺不取衰"
        },
    ],
}
