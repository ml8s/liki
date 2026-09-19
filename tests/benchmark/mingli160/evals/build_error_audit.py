#!/usr/bin/env python3
"""从 MingLi 160 运行结果生成逐题命理逻辑审查台账。

用途：把 80 道错题每题的盘面任务、命理对象、条件链、失败假设和表格优化动作固化，
避免只按类别泛泛修断语。输出 CSV 供评审和回归追踪。
"""
from __future__ import annotations

import csv
import json
import re
from collections import defaultdict
from pathlib import Path

HERE = Path(__file__).resolve().parent
ROOT = HERE / "../../../.."
ROOT = HERE / "../../../.."
RESULT = ROOT / "tests/benchmark/mingli160/workspace/iteration-1/result.json"
GROUPS = HERE / "../groups.json"
CATS = HERE / "../cats.json"
OUT_CSV = HERE / "error_audit_iteration1.csv"
OUT_JSON = HERE / "error_audit_iteration1.json"


def task_and_focus(domain: str, question: str) -> tuple[str, str]:
    q = question.lower()
    if re.search(r"何年|哪一年|哪年|何时|什么时候|哪段时间|几岁", q):
        task = "应期"
    elif re.search(r"发生(了)?何|发生(了)?什么|发生何|发生过|重要事件", q):
        task = "流年事件"
    elif re.search(r"原因|为什么|为何", q):
        task = "因果归因"
    elif re.search(r"状况|情况|如何|是否|有没有|什么职业|什么行业|为何|身材|性格|学历|婚姻|子女|父|母|家", q):
        task = "状态取象"
    else:
        task = "事件细则"

    if re.search(r"结婚|离婚|婚姻|感情|恋爱|拍拖|配偶|夫妻", question):
        focus = "婚姻感情"
    elif re.search(r"职业|行业|工作|创业|收入|老板|高管", question):
        focus = "事业职业"
    elif re.search(r"财|富|房产|买房|身家|破财|债务", question):
        focus = "财运房产"
    elif re.search(r"病|健康|手术|抑郁|骨折|疾", question):
        focus = "健康"
    elif re.search(r"子女|孩子|生育|生子", question):
        focus = "子女"
    elif re.search(r"父亲|母親|母亲|父母|祖上|家庭|家里|原生|叔叔", question):
        focus = "家庭六亲"
    elif re.search(r"学历|大学|科系|读书", question):
        focus = "学业"
    elif re.search(r"性格|个性|评价|样貌|身材|外貌", question):
        focus = "性格外貌"
    elif domain == "官非":
        focus = "官非"
    elif domain == "灾劫":
        focus = "灾劫"
    else:
        focus = domain
    return task, focus


def logic_profile(domain: str, task: str, focus: str, question: str) -> tuple[str, str, str, str]:
    """返回：命理对象、盘层、条件链、失败假设。"""

    if focus == "婚姻感情" or domain == "婚姻":
        obj = "配偶星（男财/女官杀）、夫妻宫、配偶宫星曜"
        layers = "本命八字 + 本命紫微；事件题加大运/大限/流年"
        if task in {"应期", "流年事件"}:
            chain = ("婚姻阶段未婚/已婚/离异 + 配偶星喜忌与透藏 + 大运配偶星窗口 "
                     "+ 流年配偶星透/值/合/冲 + 夫妻宫冲合刑害 + 红鸾天喜桃花佐证 → 成婚/婚变候选")
            hypo = "候选年可能命中泛婚动，但缺婚姻阶段、方向与证据闭环，导致成婚/婚变误选"
            opt = "婚姻断语条件组必须加入婚姻阶段、配偶星喜忌、夫妻宫状态和流年方向门槛"
        elif task == "因果归因":
            chain = ("夫妻宫星曜与四化 + 配偶星喜忌混杂受克 + 流年引动 + 选项语义（外遇/暴力/异地/长辈）闭环")
            hypo = "同域婚动/婚变候选同现，缺少对方行为、婚姻阶段和事件因果的选项语义条件"
            opt = "增加婚姻因果复合因子；禁止用泛婚动直接解释分手/暴力/再婚"
        else:
            chain = ("配偶星现/不现、喜忌、透藏、得地受克、混杂取清 + 夫妻宫安静/破 + "
                     "紫微夫妻宫主星四化 → 当前婚姻状态候选")
            hypo = "本命婚姻候选过宽或紫微宫位证据被泛化，未区分未婚/已婚/离异/再婚"
            opt = "建立婚姻稳定/不稳/迟滞/再婚倾向复合因子，并锁定状态题只用本命证据"
        return obj, layers, chain, hypo, opt

    if focus in {"事业职业", "事业"} or domain == "事业":
        obj = "官杀（名分压力/权柄）、印星（单位文书）、食伤（技艺表达）、财星（收入）、官禄宫"
        layers = "本命八字 + 本命紫微官禄宫；现职/应期加当前大运、大限、流年"
        if task in {"应期", "流年事件"}:
            chain = ("官杀为用/为忌 + 官杀有力受制 + 换运/岁运并临 + 流年事业星 "
                     "+ 紫微官禄宫四化 → 事业突破/变动/受阻候选")
            hypo = "事业升/阻/变同现，缺职业形态和事件方向门槛"
            opt = "事业流年断语加入受雇/自营/技艺/管理形态与用喜忌方向门槛"
        elif "行业" in question or "职业" in question or "工作" in question:
            chain = ("十神职业类象（官印/食伤/比劫伤官/财星/七杀） + 官禄宫主星吉煞 "
                     "+ 大限主题 + 身强弱任官任财 → 职业形态候选")
            hypo = "职业类象过宽，受雇/自营/技艺/管理/特殊行业缺少复合因子"
            opt = "新增职业形态复合因子，并要求选项词与类象证据闭环"
        else:
            chain = "财星喜忌 + 身强任财 + 比劫夺财 + 食伤生财 + 官杀压力 + 官禄宫吉凶 → 收入与事业状态候选"
            hypo = "收入层级、稳定性、自营负债混用同一财运信号"
            opt = "拆分收入稳定性、财富层级、自营/受雇三组条件"
        return obj, layers, chain, hypo, opt

    if focus == "财运房产" or domain == "财运":
        obj = "财星、身强弱、比劫、食伤、财库、田宅宫"
        layers = "本命八字 + 紫微田宅/财帛；应期加流年"
        if task in {"应期", "流年事件"}:
            chain = ("流年财星透/克 + 财喜忌 + 身强任财 + 比劫夺财 + 食伤生财 "
                     "+ 财库冲动 + 田宅宫四化 → 得财/破财/置业候选")
            hypo = "得财与破财候选同现，缺财富对象（现金/房产/父亲财）和方向门槛"
            opt = "财运断语按命主现金、房产、出身、父亲财分对象建条件组"
        else:
            chain = "身强任财 + 财为用透根 + 比劫不夺 + 食伤生财 + 财库田宅吉 → 财富层级候选"
            hypo = "财富层级断语过宽，未区分当前收入、存量资产、出身家境和父亲身家"
            opt = "新增财富层级/房产/父财复合因子，不得用泛财运作答"
        return obj, layers, chain, hypo, opt

    if focus == "子女" or domain == "子女":
        obj = "子女星（男官杀/女食伤）、时柱、子女宫、流年子女曜"
        layers = "本命八字 + 紫微子女宫；生育应期加大限流年"
        chain = "子女星透藏喜忌 + 时柱/子女宫吉凶 + 流年子女星宫引动 + 孕产刑冲 → 生育/孕产候选"
        hypo = "生育应期、孕产风险、子女性别与数量混用；传统规则本身低置信"
        opt = "子女断语拆分应期、状态、数量、性别；孕产凶险不得否认定育，数量性别只作低置信"
        return obj, layers, chain, hypo, opt

    if focus == "家庭六亲" or domain == "家庭":
        if "父" in question:
            obj = "偏财父星、年柱父母宫、父母宫星曜"
            chain = "父星喜忌透藏 + 年柱/父母宫冲刑 + 流年克父星/父星入墓/吊客丧门 → 父亲状态/应期候选"
        elif "母" in question:
            obj = "正印母星、父母宫、印星旺衰"
            chain = "母星喜忌透藏 + 父母宫冲刑 + 流年克母星/印星受损 → 母亲状态/应期候选"
        else:
            obj = "年柱父母宫、父母宫星曜、兄弟宫、相应六亲星"
            chain = "六亲星喜忌 + 对应宫位冲合刑害 + 大运流年引动 → 家庭状态候选"
        layers = "本命八字 + 紫微父母/兄弟宫；事件题加大运流年"
        hypo = "六亲灾信号过泛，未分层到关系、健康、分离、经济与严重度"
        opt = "六亲断语拆分关系/健康/分离/经济/重大灾；重大结论须星、宫、岁运三重闭环"
        return obj, layers, chain, hypo, opt

    if focus == "健康" or domain == "健康":
        obj = "日主旺衰、五行亢衰、疾厄宫、福德宫、流年病符"
        layers = "本命八字 + 紫微疾厄/福德；事件题加流年"
        chain = "日主强弱 + 五行亢衰受制 + 疾厄宫四化煞曜 + 流年病符/刑冲 → 健康压力候选"
        hypo = "五行取象能给候选，但器官、病名、手术细节缺确定性条件"
        opt = "健康只作低置信候选；不硬凑器官和病名，重大医疗结论必须提示现实核验"
        return obj, layers, chain, hypo, opt

    if focus == "学业" or domain == "学业":
        obj = "印星、官杀、食伤、官禄宫、文昌文曲"
        layers = "本命八字 + 紫微官禄/学业相关宫；升学事件加流年"
        chain = "印星喜忌旺衰 + 官杀化印/财坏印 + 食伤泄秀 + 文昌文曲/流年科甲 → 学历与考试候选"
        hypo = "学历层次与院校路径缺分档条件，流年考试噪声影响本命状态"
        opt = "学历断语按高中/专科/本科/研究生命名分档；状态题禁止流年覆盖本命"
        return obj, layers, chain, hypo, opt

    if focus == "性格外貌" or domain in {"性格", "外貌"}:
        obj = "月令本气十神、十神旺衰、日主五行、命宫主星"
        layers = "本命八字 + 本命紫微；不使用流年"
        chain = "月令本气十神定主面 + 十神旺衰辅面 + 日主五形体态 + 命宫星曜表达 → 性格外貌候选"
        hypo = "多断语同现时辅面覆盖月令主面，或传统外貌类象不足以区分现代描述"
        opt = "锁定主面优先，并对外貌/性格类象做候选置信标注"
        return obj, layers, chain, hypo, opt

    if focus == "官非" or domain == "官非":
        obj = "官杀、刑冲、灾煞、疾厄宫"
        layers = "本命八字 + 大运流年"
        chain = "伤官见官 + 官杀为忌 + 三刑/六冲 + 灾煞羊刃 + 流年引动 → 官非候选"
        hypo = "官非候选存在，但缺少应期强度与事件类型门槛"
        opt = "官非断语加入刑冲官杀灾煞复合条件，并标低置信"
        return obj, layers, chain, hypo, opt

    if focus == "灾劫" or domain == "灾劫":
        obj = "日主、羊刃、灾煞、天克地冲、疾厄宫"
        layers = "本命八字 + 大运流年"
        chain = "羊刃灾煞 + 天克地冲/三刑 + 五行攻身 + 疾厄宫凶 → 灾劫候选"
        hypo = "凶象候选同现，但具体灾种和年份排序缺闭环"
        opt = "灾劫断语区分意外、手术、血光、官非，并要求流年引动"
        return obj, layers, chain, hypo, opt

    if focus == "运势" or domain == "运势":
        obj = "大运干支十神、用喜忌、大限主题"
        layers = "本命 + 大运 + 大限 + 流年"
        chain = "大运十神喜忌 + 大限主题 + 流年引动 + 三方四正 → 阶段吉凶候选"
        hypo = "大运好坏缺统一评分条件，且流年杂音覆盖阶段主象"
        opt = "大运定性按受雇/自营/健康/迁移等事类分别建条件组"
        return obj, layers, chain, hypo, opt

    obj = "对应六亲星或十神"
    layers = "本命八字 + 本命紫微；事件题加岁运"
    chain = "领域目标星 + 对应宫位 + 喜忌旺衰 + 岁运引动 → 领域候选"
    hypo = "领域候选与选项语义未闭环"
    opt = "补齐领域专属复合因子，禁止跨域信号作主证"
    return obj, layers, chain, hypo, opt


def optimization_from_hypo(hypo: str, opt: str) -> str:
    if "候选年可能命中" in hypo:
        return "table: 给应期断语增加阶段/方向/喜忌条件组"
    if "候选过宽" in hypo:
        return "table: 增加状态/形态复合因子，并收敛断语条件组"
    if "同现" in hypo:
        return "table: 增加事件方向复合因子与排除条件"
    if "低置信" in hypo:
        return "table: 标注低置信候选，不新增确定性断语"
    return f"table: {opt}"


def main() -> int:
    result = json.loads(RESULT.read_text(encoding="utf-8"))
    groups = json.loads(GROUPS.read_text(encoding="utf-8"))
    cats = json.loads(CATS.read_text(encoding="utf-8"))
    prompts = {c["case_id"]: c.get("prompt", "") for c in result["case_results"]}

    audit: list[dict] = []
    for case in result["case_results"]:
        evidence = next(
            (a["evidence"] for a in case.get("grading", {}).get("assertion_results", [])
             if a.get("text", "").startswith("script:")),
            "",
        )
        prompt = prompts[case["case_id"]]
        response = case.get("response", "")
        for line in evidence.splitlines():
            m = re.match(r"\s*题(\d+): 答案=([A-D]) 正确=([A-D]) ([✓✗])", line)
            if not m:
                continue
            q_no = int(m.group(1))
            std = m.group(3)
            model = m.group(2)
            passed = m.group(4) == "✓"
            if passed:
                continue

            qid = groups[case["case_id"]][q_no - 1]
            qm = re.search(rf"【题{q_no}】(.*?)(?=【题{q_no + 1}】|要求：|$)", prompt, re.S)
            block = qm.group(1) if qm else ""
            question = re.sub(r"选项：.*", "", block, flags=re.S).strip()
            options = dict(re.findall(r"([A-D])[.、]\s*([^\n]+)", block))
            task, focus = task_and_focus(cats[qid], question)
            obj, layers, chain, hypo, opt = logic_profile(
                cats[qid], task, focus, question
            )
            cited = sorted(set(re.findall(
                r"\b(?:y[A-Za-z]{1,8}|x[A-Za-z]{1,8}|hun|jk|zy|sh|zw)_\d+[A-Za-z0-9\-]*\b",
                response,
            )))
            audit.append({
                "error_id": qid,
                "case_id": case["case_id"],
                "question_no": q_no,
                "domain": cats[qid],
                "task_type": task,
                "focus": focus,
                "standard_answer": std,
                "model_answer": model,
                "question": question,
                "standard_option": options.get(std, ""),
                "model_option": options.get(model, ""),
                "mingli_object": obj,
                "chart_layers": layers,
                "mingli_condition_chain": chain,
                "failure_hypothesis": hypo,
                "table_optimization": optimization_from_hypo(hypo, opt),
                "case_cited_assertions": ";".join(cited),
            })

    fields = list(audit[0].keys())
    with OUT_CSV.open("w", encoding="utf-8", newline="") as fh:
        writer = csv.DictWriter(fh, fieldnames=fields, lineterminator="\n")
        writer.writeheader()
        writer.writerows(audit)
    OUT_JSON.write_text(json.dumps(audit, ensure_ascii=False, indent=2), encoding="utf-8")
    print(f"written {len(audit)} errors -> {OUT_CSV}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
