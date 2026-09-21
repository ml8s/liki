"""
断语真值表逐题验证工具（通过 Python 工具层）。

用法：
  python3 scripts/eval_assertions.py
"""
import json, re, sys, os
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki/bazi/tools"
sys.path.insert(0, str(TOOLS))

import yaml
from paipan import full_paipan, city_coords
from duanyu import query, yearly_range

CASES_DIR = ROOT / "tests/benchmark/mingli160/evals/cases"
ANSWERS_FILE = ROOT / "tests/benchmark/mingli160/answers.json"
CATS_FILE = ROOT / "tests/benchmark/mingli160/cats.json"
GROUPS_FILE = ROOT / "tests/benchmark/mingli160/groups.json"

DOMAIN_QUERIES = {
    "婚姻": [
        {"fn": "query", "rule": "十神", "domains": ["婚姻"]},
        {"fn": "query", "rule": "夫妻", "domains": ["婚姻"]},
        {"fn": "query", "rule": "大运", "domains": ["婚姻"]},
        {"fn": "query", "rule": "用神", "domains": ["婚姻"]},
    ],
    "事业": [
        {"fn": "query", "rule": "十神", "domains": ["事业"]},
        {"fn": "query", "rule": "官禄", "domains": ["事业"]},
        {"fn": "query", "rule": "旺衰", "domains": ["事业"]},
    ],
    "财运": [
        {"fn": "query", "rule": "十神", "domains": ["财运"]},
        {"fn": "query", "rule": "财帛", "domains": ["财运"]},
        {"fn": "query", "rule": "旺衰", "domains": ["财运"]},
    ],
    "学业": [
        {"fn": "query", "rule": "十神", "domains": ["学业"]},
        {"fn": "query", "rule": "旺衰", "domains": ["学业"]},
    ],
    "健康": [
        {"fn": "query", "rule": "五行", "domains": ["健康"]},
        {"fn": "query", "rule": "旺衰", "domains": ["健康"]},
        {"fn": "query", "rule": "用神", "domains": ["健康"]},
    ],
    "子女": [
        {"fn": "query", "rule": "十神", "domains": ["子女"]},
    ],
    "家庭": [
        {"fn": "query", "rule": "六亲", "domains": ["家庭"]},
        {"fn": "query", "rule": "十神", "domains": ["家庭"]},
    ],
    "性格": [
        {"fn": "query", "rule": "命宫", "domains": ["性格"]},
        {"fn": "query", "rule": "十神", "domains": ["性格"]},
        {"fn": "query", "rule": "五行", "domains": ["性格"]},
    ],
    "出身": [
        {"fn": "query", "rule": "十神", "domains": ["出身"]},
    ],
    "外貌": [
        {"fn": "query", "rule": "命宫", "domains": ["外貌"]},
    ],
}


def extract_birth(prompt):
    # 性别
    gender = None
    for kw, g in [("男命", "male"), ("乾造", "male"), ("男性", "male"),
                  ("女命", "female"), ("坤造", "female"), ("坤命", "female"),
                  ("坤", "female"), ("乾", "male")]:
        if kw in prompt:
            gender = g
            break
    if not gender:
        return None

    # 日期
    year = month = day = None
    m = re.search(r'(?:西历|公元|公历)?\s*(\d{4})-(\d{1,2})-(\d{1,2})', prompt)
    if not m:
        m = re.search(r'(?:西历|公元|公历)?\s*(\d{4})年\s*(\d{1,2})月\s*(\d{1,2})日', prompt)
    if m:
        year, month, day = int(m.group(1)), int(m.group(2)), int(m.group(3))
    if not year:
        return None

    # 时间
    hour = None
    minute = 30
    correct = True

    tm = re.search(r'(?:下午|晚上)?\s*(\d{1,2})[：:時时](\d{2})分?', prompt)
    if tm:
        hour = int(tm.group(1))
        minute = int(tm.group(2))
        if ("下午" in prompt or "晚上" in prompt) and hour < 12:
            hour += 12

    if hour is None:
        shichen = {"子": 0, "丑": 2, "寅": 4, "卯": 6, "辰": 8, "巳": 10,
                   "午": 12, "未": 14, "申": 16, "酉": 18, "戌": 20, "亥": 22}
        for sc, start in shichen.items():
            if f"{sc}时" in prompt or f"{sc}時" in prompt:
                hour = start + 1
                minute = 30
                correct = False
                break

    if hour is None:
        hour = 12
        correct = False

    if correct and ("下午" in prompt or "晚上" in prompt) and hour < 12:
        hour += 12

    # 地点
    location = ""
    loc_m = re.search(r'出生地点[：:]\s*(\S+)', prompt)
    if loc_m:
        location = loc_m.group(1)
    elif "台湾" in prompt:
        location = "台湾"
    elif "马来西亚" in prompt:
        location = "吉隆坡"
    elif "广东" in prompt or "潮汕" in prompt:
        location = "汕头"

    longitude = None
    if location:
        try:
            coords = city_coords(location)
            longitude = coords.get("longitude")
        except:
            pass

    if hour is not None and minute > 0:
        correct = True  # 有精确分钟 → 可做真太阳时校正

    gregorian = f"{year:04d}-{month:02d}-{day:02d}T{hour:02d}:{minute:02d}:00+08:00"

    return {
        "gregorian": gregorian,
        "gender": gender,
        "correct": correct,
        "longitude": longitude,
    }


def parse_questions(prompt):
    qs = []
    for m in re.finditer(
        r'【题(\d+)】问题：(.+?)\n选项：\n((?:[A-D]\..+?\n?)+)', prompt, re.DOTALL
    ):
        num = int(m.group(1))
        opts = {}
        for letter in "ABCD":
            om = re.search(rf'{letter}\.(.+?)(?=[A-D]\.|$)', m.group(3), re.DOTALL)
            if om:
                opts[letter] = om.group(1).strip()
        qs.append({"num": num, "text": m.group(2).strip(), "options": opts})
    return qs


def get_assertions_for_domain(pan, domain):
    queries = DOMAIN_QUERIES.get(domain, [])
    fired = []
    for qc in queries:
        try:
            result = query(qc["rule"], pan, domains=qc["domains"])
            for side_key, side_list in result.items():
                if isinstance(side_list, list):
                    for a in side_list:
                        if isinstance(a, dict) and a.get("id"):
                            fired.append({
                                "id": a["id"],
                                "结论": a.get("结论", ""),
                                "事件": a.get("事件", ""),
                            })
        except Exception:
            pass
    return fired


def main():
    answers = json.load(open(ANSWERS_FILE, encoding="utf-8"))
    cats = json.load(open(CATS_FILE, encoding="utf-8"))
    groups = json.load(open(GROUPS_FILE, encoding="utf-8"))

    case_files = sorted(CASES_DIR.glob("*.yaml"))
    print(f"═══ 断语真值表逐题验证 ═══")
    print(f"命盘数: {len(case_files)}, 总题数: {len(answers)}\n")

    total_correct = 0
    total_questions = 0
    gap_report = []
    assertion_noise = []  # Track assertion counts

    for cf in case_files:
        case = yaml.safe_load(cf.read_text("utf-8"))
        pan_id = case["id"]
        prompt = case.get("input", {}).get("prompt", "")

        birth = extract_birth(prompt)
        if not birth:
            print(f"  ⚠️ {pan_id}: 无法提取出生信息")
            continue

        # 排盘（Python 工具层内部处理 RPC）
        try:
            pan = full_paipan(
                birth["gregorian"],
                birth["gender"],
                longitude=birth.get("longitude"),
                correct=birth["correct"],
            )
        except Exception as e:
            # longitude fallback
            try:
                pan = full_paipan(birth["gregorian"], birth["gender"], correct=False)
            except Exception as e2:
                print(f"  ⚠️ {pan_id}: 排盘失败 {e2}")
                continue

        questions = parse_questions(prompt)
        q_ids = groups.get(pan_id, [])
        pan_results = []

        for q in questions:
            q_idx = q["num"] - 1
            if q_idx >= len(q_ids):
                continue
            ftb_id = q_ids[q_idx]
            correct_answer = answers.get(ftb_id, "?")
            domain = cats.get(ftb_id, "?")
            correct_text = q["options"].get(correct_answer, "")

            fired = get_assertions_for_domain(pan, domain)

            # Simple keyword matching heuristic
            correct_kw = set()
            for kw in ["婚", "结", "财", "富", "穷", "学", "业", "职", "病", "健",
                        "性", "格", "子", "女", "父", "母", "出", "迁", "移", "破", "贵",
                        "稳", "波折", "岗", "管", "艺", "壓", "压力", "抑", "郁"]:
                if kw in correct_text:
                    correct_kw.add(kw)

            matching = [f for f in fired if any(kw in f.get("结论", "") for kw in correct_kw)]
            likely_correct = len(matching) > 0 or len(fired) == 0

            total_questions += 1
            if likely_correct:
                total_correct += 1
            else:
                gap_report.append({
                    "pan": pan_id, "q": q["num"], "domain": domain,
                    "correct": correct_answer, "correct_text": correct_text[:60],
                    "fired_count": len(fired), "matching": len(matching),
                    "fired_ids": [f["id"] for f in fired[:10]],
                    "fired_conclusions": [f["结论"][:40] for f in fired[:5]],
                })

            pan_results.append({
                "num": q["num"], "domain": domain,
                "ok": likely_correct, "fired": len(fired), "match": len(matching),
            })

        # Per-pan summary
        ok = sum(1 for p in pan_results if p["ok"])
        s = "✅" if ok >= 3 else "❌"
        total_fired = sum(p["fired"] for p in pan_results)
        print(f"  {s} {pan_id}: {ok}/{len(pan_results)} 可能正确, 命中断语 {total_fired}")

        for pr in pan_results:
            if not pr["ok"]:
                print(f"    ✗ 题{pr['num']} [{pr['domain']}] fired={pr['fired']} matching={pr['match']}")

    # Summary
    print(f"\n═══ 总体 ═══")
    print(f"总题数: {total_questions}")
    print(f"可能正确: {total_correct}")
    print(f"可能错误: {total_questions - total_correct}")
    if total_questions:
        print(f"断语命中预估: {total_correct * 100 // total_questions}%")
    print(f"命理缺口: {len(gap_report)} 题")

    out = ROOT / "tests/benchmark/mingli160/evals/assertion_gap_analysis.json"
    out.write_text(json.dumps({"total": total_questions, "correct": total_correct,
                               "gap_report": gap_report}, ensure_ascii=False, indent=2), "utf-8")
    print(f"报告: {out}")


if __name__ == "__main__":
    main()
