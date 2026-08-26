"""答案双源一致性守卫。

评测答案存在两个必然形态（skill-up script judge 被复制到容器 workspace 执行，
不能读外部文件，只能内嵌）：tests/answers.json（判分数据源）与
tests/grade-case.py 内嵌 Q_ANSWERS（judge 自包含副本）。本测试防两处静默漂移。
"""
import importlib.util
import json
import os


def _load_judge_module():
    path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "grade-case.py")
    spec = importlib.util.spec_from_file_location("grade_case", path)
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)  # grade-case.py 顶层只有数据定义，无副作用
    return mod


def test_grade_case_answers_match_answers_json():
    base = os.path.dirname(os.path.abspath(__file__))
    answers = json.load(open(os.path.join(base, "answers.json"), encoding="utf-8"))
    embedded = _load_judge_module().Q_ANSWERS
    assert set(embedded) == set(answers), (
        f"题号集合不一致：仅在内嵌={set(embedded) - set(answers)} "
        f"仅在 answers.json={set(answers) - set(embedded)}"
    )
    diff = {q: (embedded[q], answers[q]) for q in answers if embedded[q] != answers[q]}
    assert not diff, f"答案漂移（内嵌, json）：{diff}"
