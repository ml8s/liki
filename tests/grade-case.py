#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""skill-up script judge：MingLi-Bench 命盘答案判分（自包含实现）

skill-up 标准机制输入（由 skill-up 注入环境变量）：
  $EVAL_FINAL_MESSAGE    agent 最终输出（qwen_code 的完整分析文本，含「题N 答案：X」）
  $EVAL_TRANSCRIPT_PATH  transcript.json（完整对话；兜底提取 assistant 消息里的答案行）

说明：
  - 判分数据内嵌（CASE_BIRTH / CASE_QTYPES / Q_ANSWERS）——skill-up 将本脚本上传复制到
    临时 workspace 执行，不能依赖外部文件（__file__ 指向副本）。
  - 盘号识别：transcript 的 user prompt 含「出生信息」→ 匹配 CASE_BIRTH（唯一标识）。
  - 全对 exit 0（PASS），否则 exit 1（FAIL）；stdout 为判分依据（进报告 rationale）。
"""
import json
import os
import re
import sys

# ===== MingLi-Bench 评测数据（内嵌自包含——script judge 在复制后的 workspace 执行，不能依赖外部文件）=====
# 答案与 tests/answers.json 为同一份事实的两个形态（内嵌是 skill-up 复制约束所迫）——
# 改答案时两处同步改，tests/test_grade_sync.py 会校验一致性
CASE_BIRTH = {
    'pan01': '男命：1974年4月28日下午4:40分 出生地点：usa',
    'pan02': '女命：阳历1981年5月26日凌晨02:17 出生地点：中国香港',
    'pan03': '女命（佩小姐）：1990年05月23日17：30 出生地点：中国大陆',
    'pan04': '女命：公历1977年10月26日11时10分 出生地点：malaysia',
    'pan05': '男命：公历1985年6月29日18时32分（酉时）出生地点：malaysia',
    'pan06': '男命：公历1983年11月1日（亥时） 出生地点：中国台湾',
    'pan07': '女命公历1984年12月9日（酉时） 出生地点：中国香港',
    'pan08': '男命公历1981年6月17日（戌时）出生地点：中国香港',
    'pan09': '女命：西历1966年10月18日夜晚23:15香港出生',
    'pan10': '男命：西历1962年8月6日已时-农历1962七月初七/巳时',
    'pan11': '男命：西历1986年4月24日晚上9.30亥时',
    'pan12': '男命，西历1956年3月11日巳时，malaysia出生',
    'pan13': '女命：西历1978年04月05日17：00-19：00酉时  台湾生',
    'pan14': '女命：西历1958-10-27辰时生  台湾人',
    'pan15': '女命：西历2002年6月1日早上08:30辰时  出生于香港',
    'pan16': '男命：西历1980年7月11日已时  香港出生',
    'pan17': '女命：1980年8月24日下午16:30 广东',
    'pan18': '男性：生于公元1972年1月8日午时，农历辛亥11月22日午时。于潮汕地区出生。',
    'pan19': '男命：1961年12月30日 02:00丑',
    'pan20': '女命：1983年10月28日丑时农历：癸亥年九月二十三丑时。',
    'pan21': '坤命：西元1971年4月12日申時生，出生地區：中國',
    'pan22': '女命：公元1983年3月26日未时，农历1983癸亥年，二月十二日，未时生，北京出世。',
    'pan23': '女命：1980年9月21日早上11：25午时，出生malaysia',
    'pan24': '男命：2007年1月31日0900巳时，malaysia出生。',
    'pan25': '坤造：广东出生 西历1951年11月14日巳时',
    'pan26': '女命：西历1987年7月5日午时，香港出生',
    'pan27': '男命：西历1983年4月21日上午6-7时，日本宫崎县出生',
    'pan28': '男命：西元1993年4月8日23:34子时，新加坡出生',
    'pan29': '乾造：西历1988年1月10日早上8时12分，马来西亚出生',
    'pan30': '男命：西历1973年8月24日00:35子时，马来西亚出生',
    'pan31': '女命：西历1988年2月15日16:50申时，台湾出生',
    'pan32': '男命：西元1970年7月22日申时，北京出生',
}
CASE_QTYPES = {
    'pan01': ['ftb_0001', 'ftb_0002', 'ftb_0003', 'ftb_0004', 'ftb_0005'],
    'pan02': ['ftb_0006', 'ftb_0007', 'ftb_0008', 'ftb_0009', 'ftb_0010'],
    'pan03': ['ftb_0011', 'ftb_0012', 'ftb_0013', 'ftb_0014', 'ftb_0015'],
    'pan04': ['ftb_0016', 'ftb_0017', 'ftb_0018', 'ftb_0019', 'ftb_0020'],
    'pan05': ['ftb_0021', 'ftb_0022', 'ftb_0023', 'ftb_0024', 'ftb_0025'],
    'pan06': ['ftb_0026', 'ftb_0027', 'ftb_0028', 'ftb_0029', 'ftb_0030'],
    'pan07': ['ftb_0031', 'ftb_0032', 'ftb_0033', 'ftb_0034', 'ftb_0035'],
    'pan08': ['ftb_0036', 'ftb_0037', 'ftb_0038', 'ftb_0039', 'ftb_0040'],
    'pan09': ['ftb_0041', 'ftb_0042', 'ftb_0043', 'ftb_0044', 'ftb_0045'],
    'pan10': ['ftb_0046', 'ftb_0047', 'ftb_0048', 'ftb_0049', 'ftb_0050'],
    'pan11': ['ftb_0051', 'ftb_0052', 'ftb_0053', 'ftb_0054', 'ftb_0055'],
    'pan12': ['ftb_0056', 'ftb_0057', 'ftb_0058', 'ftb_0059', 'ftb_0060'],
    'pan13': ['ftb_0061', 'ftb_0062', 'ftb_0063', 'ftb_0064', 'ftb_0065'],
    'pan14': ['ftb_0066', 'ftb_0067', 'ftb_0068', 'ftb_0069', 'ftb_0070'],
    'pan15': ['ftb_0071', 'ftb_0072', 'ftb_0073', 'ftb_0074', 'ftb_0075'],
    'pan16': ['ftb_0076', 'ftb_0077', 'ftb_0078', 'ftb_0079', 'ftb_0080'],
    'pan17': ['ftb_0081', 'ftb_0082', 'ftb_0083', 'ftb_0084', 'ftb_0085'],
    'pan18': ['ftb_0086', 'ftb_0087', 'ftb_0088', 'ftb_0089', 'ftb_0090'],
    'pan19': ['ftb_0091', 'ftb_0092', 'ftb_0093', 'ftb_0094', 'ftb_0095', 'ftb_0096'],
    'pan20': ['ftb_0097', 'ftb_0098', 'ftb_0099', 'ftb_0100'],
    'pan21': ['ftb_0101', 'ftb_0102', 'ftb_0103', 'ftb_0104', 'ftb_0105'],
    'pan22': ['ftb_0106', 'ftb_0107', 'ftb_0108', 'ftb_0109', 'ftb_0110'],
    'pan23': ['ftb_0111', 'ftb_0112', 'ftb_0113', 'ftb_0114', 'ftb_0115'],
    'pan24': ['ftb_0116', 'ftb_0117', 'ftb_0118', 'ftb_0119', 'ftb_0120'],
    'pan25': ['ftb_0121', 'ftb_0122', 'ftb_0123', 'ftb_0124', 'ftb_0125'],
    'pan26': ['ftb_0126', 'ftb_0127', 'ftb_0128', 'ftb_0129', 'ftb_0130'],
    'pan27': ['ftb_0131', 'ftb_0132', 'ftb_0133', 'ftb_0134', 'ftb_0135'],
    'pan28': ['ftb_0136', 'ftb_0137', 'ftb_0138', 'ftb_0139', 'ftb_0140'],
    'pan29': ['ftb_0141', 'ftb_0142', 'ftb_0143', 'ftb_0144', 'ftb_0145'],
    'pan30': ['ftb_0146', 'ftb_0147', 'ftb_0148', 'ftb_0149', 'ftb_0150'],
    'pan31': ['ftb_0151', 'ftb_0152', 'ftb_0153', 'ftb_0154', 'ftb_0155'],
    'pan32': ['ftb_0156', 'ftb_0157', 'ftb_0158', 'ftb_0159', 'ftb_0160'],
}
Q_ANSWERS = {
    'ftb_0001': 'A',
    'ftb_0002': 'C',
    'ftb_0003': 'C',
    'ftb_0004': 'C',
    'ftb_0005': 'B',
    'ftb_0006': 'B',
    'ftb_0007': 'B',
    'ftb_0008': 'A',
    'ftb_0009': 'D',
    'ftb_0010': 'B',
    'ftb_0011': 'D',
    'ftb_0012': 'C',
    'ftb_0013': 'D',
    'ftb_0014': 'C',
    'ftb_0015': 'B',
    'ftb_0016': 'A',
    'ftb_0017': 'A',
    'ftb_0018': 'B',
    'ftb_0019': 'A',
    'ftb_0020': 'D',
    'ftb_0021': 'A',
    'ftb_0022': 'B',
    'ftb_0023': 'B',
    'ftb_0024': 'C',
    'ftb_0025': 'B',
    'ftb_0026': 'A',
    'ftb_0027': 'D',
    'ftb_0028': 'D',
    'ftb_0029': 'C',
    'ftb_0030': 'B',
    'ftb_0031': 'D',
    'ftb_0032': 'C',
    'ftb_0033': 'A',
    'ftb_0034': 'C',
    'ftb_0035': 'B',
    'ftb_0036': 'C',
    'ftb_0037': 'C',
    'ftb_0038': 'B',
    'ftb_0039': 'A',
    'ftb_0040': 'B',
    'ftb_0041': 'B',
    'ftb_0042': 'C',
    'ftb_0043': 'B',
    'ftb_0044': 'A',
    'ftb_0045': 'D',
    'ftb_0046': 'A',
    'ftb_0047': 'B',
    'ftb_0048': 'B',
    'ftb_0049': 'D',
    'ftb_0050': 'D',
    'ftb_0051': 'D',
    'ftb_0052': 'D',
    'ftb_0053': 'C',
    'ftb_0054': 'C',
    'ftb_0055': 'D',
    'ftb_0056': 'C',
    'ftb_0057': 'A',
    'ftb_0058': 'A',
    'ftb_0059': 'C',
    'ftb_0060': 'C',
    'ftb_0061': 'B',
    'ftb_0062': 'C',
    'ftb_0063': 'C',
    'ftb_0064': 'B',
    'ftb_0065': 'B',
    'ftb_0066': 'C',
    'ftb_0067': 'A',
    'ftb_0068': 'C',
    'ftb_0069': 'C',
    'ftb_0070': 'D',
    'ftb_0071': 'B',
    'ftb_0072': 'D',
    'ftb_0073': 'C',
    'ftb_0074': 'D',
    'ftb_0075': 'A',
    'ftb_0076': 'C',
    'ftb_0077': 'C',
    'ftb_0078': 'B',
    'ftb_0079': 'B',
    'ftb_0080': 'D',
    'ftb_0081': 'B',
    'ftb_0082': 'B',
    'ftb_0083': 'D',
    'ftb_0084': 'C',
    'ftb_0085': 'C',
    'ftb_0086': 'D',
    'ftb_0087': 'C',
    'ftb_0088': 'D',
    'ftb_0089': 'C',
    'ftb_0090': 'A',
    'ftb_0091': 'B',
    'ftb_0092': 'A',
    'ftb_0093': 'B',
    'ftb_0094': 'A',
    'ftb_0095': 'C',
    'ftb_0096': 'B',
    'ftb_0097': 'A',
    'ftb_0098': 'A',
    'ftb_0099': 'D',
    'ftb_0100': 'A',
    'ftb_0101': 'B',
    'ftb_0102': 'A',
    'ftb_0103': 'C',
    'ftb_0104': 'B',
    'ftb_0105': 'B',
    'ftb_0106': 'B',
    'ftb_0107': 'A',
    'ftb_0108': 'B',
    'ftb_0109': 'D',
    'ftb_0110': 'B',
    'ftb_0111': 'C',
    'ftb_0112': 'D',
    'ftb_0113': 'D',
    'ftb_0114': 'D',
    'ftb_0115': 'D',
    'ftb_0116': 'B',
    'ftb_0117': 'C',
    'ftb_0118': 'A',
    'ftb_0119': 'D',
    'ftb_0120': 'C',
    'ftb_0121': 'B',
    'ftb_0122': 'D',
    'ftb_0123': 'B',
    'ftb_0124': 'A',
    'ftb_0125': 'B',
    'ftb_0126': 'D',
    'ftb_0127': 'B',
    'ftb_0128': 'D',
    'ftb_0129': 'D',
    'ftb_0130': 'B',
    'ftb_0131': 'B',
    'ftb_0132': 'B',
    'ftb_0133': 'D',
    'ftb_0134': 'C',
    'ftb_0135': 'A',
    'ftb_0136': 'C',
    'ftb_0137': 'D',
    'ftb_0138': 'B',
    'ftb_0139': 'A',
    'ftb_0140': 'B',
    'ftb_0141': 'D',
    'ftb_0142': 'A',
    'ftb_0143': 'D',
    'ftb_0144': 'A',
    'ftb_0145': 'B',
    'ftb_0146': 'A',
    'ftb_0147': 'B',
    'ftb_0148': 'B',
    'ftb_0149': 'D',
    'ftb_0150': 'C',
    'ftb_0151': 'A',
    'ftb_0152': 'C',
    'ftb_0153': 'B',
    'ftb_0154': 'D',
    'ftb_0155': 'B',
    'ftb_0156': 'C',
    'ftb_0157': 'A',
    'ftb_0158': 'D',
    'ftb_0159': 'C',
    'ftb_0160': 'A',
}

FINAL = os.environ.get("EVAL_FINAL_MESSAGE", "")
TRANSCRIPT_PATH = os.environ.get("EVAL_TRANSCRIPT_PATH", "")


def find_case_id(prompt):
    """出生信息 → 盘号：prompt 的出生信息片段与 CASE_BIRTH 互相包含匹配。"""
    m = re.search(r"出生信息[:：]\s*([^\n]{5,80})", prompt)
    if not m:
        return None
    birth = m.group(1)
    for cid, cbirth in CASE_BIRTH.items():
        if birth in cbirth or cbirth in birth:
            return cid
    return None


def extract_answers(text):
    """提取「题N 答案：X」→ {题号: 答案}；支持「- 题1 答案：A」等 markdown 前缀。"""
    by_no = {}
    for mm in re.finditer(r"题\s*([1-6])[\s：:]*答案[：:]\s*([A-D])", text):
        by_no[int(mm.group(1))] = mm.group(2)
    if by_no:
        return by_no
    tail = text[-600:]
    m1 = re.search(r"答案[：:]\s*([A-D](?:\s+[A-D]){3,5})\s*$", tail)
    m2 = re.findall(r"答案[：:]\s*([A-D])", tail)
    if m1:
        seq = m1.group(1).split()
    elif m2:
        seq = m2
    else:
        seq = re.findall(r"答案[：:]\s*([A-D])", text)
    return {i + 1: v for i, v in enumerate(seq)}


def main():
    rsp = FINAL or ""
    transcript = []
    if TRANSCRIPT_PATH and os.path.exists(TRANSCRIPT_PATH):
        try:
            with open(TRANSCRIPT_PATH, encoding="utf-8") as fh:
                transcript = json.load(fh)
        except Exception:
            transcript = []
    prompt = ""
    for msg in transcript:
        if msg.get("role") == "user":
            prompt = msg.get("content") or ""
            break
    for msg in transcript:
        if msg.get("role") == "assistant":
            rsp += "\n" + (msg.get("content") or "")

    case_id = find_case_id(prompt)
    if not case_id:
        print(f"ERROR: 无法匹配盘号。prompt 片段: {prompt[:100]!r}")
        sys.exit(2)
    qids = CASE_QTYPES.get(case_id) or []
    if not qids:
        print(f"ERROR: 无 {case_id} 的题序")
        sys.exit(2)
    gold = [Q_ANSWERS.get(q) for q in qids]

    got_map = extract_answers(rsp)
    correct = 0
    details = []
    for i, g in enumerate(gold, 1):
        got = got_map.get(i, "?")
        ok = got == g
        if ok:
            correct += 1
        details.append(f"题{i}: 答案={got} 正确={g} {'✓' if ok else '✗'}")

    wrong = [i for i, d in enumerate(details, 1) if "✗" in d]
    print(f"[{case_id}] 正确: {correct}/{len(qids)}")
    for d in details:
        print("  " + d)
    print(f"错题: {wrong if wrong else '无'}")
    sys.exit(0 if correct == len(qids) else 1)


if __name__ == "__main__":
    main()
