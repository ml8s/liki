"""断语条件组可达性审计：防止互斥条件造成永久不可达断语。"""
import csv
import re
from functools import lru_cache
from pathlib import Path

TOOLS = Path(__file__).resolve().parents[1] / "skills/liki/natal/tools"


def _rows(path: str) -> list[dict[str, str]]:
    with (TOOLS / path).open(encoding="utf-8-sig", newline="") as fh:
        return list(csv.DictReader(fh))


def _predicate(expression: str, expected: str):
    match = re.match(r"^([^\[]+)\[(.*)\]$", expression)
    if not match:
        return ("POS", "DIRECT", expression, str(expected))
    return (
        "POS",
        match.group(1),
        tuple(item.strip() for item in match.group(2).split(",")),
        str(expected),
    )


def _bind_scalar_expected(dnf: frozenset, expected: str) -> frozenset:
    """把 direct 通路的“任意”实例化为断语约束值，使标量分支可做互斥判定。"""
    if expected in ("0", "1"):
        return dnf

    output: set[frozenset] = set()
    for branch in dnf:
        next_branch: set[tuple] = set()
        for sign, kind, args, branch_expected in branch:
            if args and args[-1] == "任意":
                branch_expected = expected
            next_branch.add((sign, kind, args, branch_expected))
        output.add(frozenset(next_branch))
    return frozenset(output)


def _negation(predicate):
    sign, kind, args, expected = predicate
    return ("NEG" if sign == "POS" else "POS", kind, args, expected)


def _has_conflict(conjunction: frozenset) -> bool:
    items = list(conjunction)
    return any(
        _conflict(left, right)
        for index, left in enumerate(items)
        for right in items[index + 1:]
    )


FACTOR_ROWS = _rows("factors/factors.csv") + _rows("factors/factors_liunian.csv")
ASSERTIONS = _rows("assertions/assertions.csv")
CONDITIONS = _rows("assertions/assertion_conditions.csv")

FACTOR_DNF_RAW: dict[str, dict[int, list[tuple]]] = {}
# 性别不是因子长表条目，而是排盘上下文；审计必须显式纳入同一个封闭标量域。
FACTOR_DNF_RAW["性别"] = {
    1: [(("POS", "直读", ("gender", "任意"), "male"), "context")],
    2: [(("POS", "直读", ("gender", "任意"), "female"), "context")],
}
for row in FACTOR_ROWS:
    factor_id = row["factor_id"]
    group_id = int(row["group_id"])
    expression = row["expression"]
    expected = row["expected"] or ("任意" if expression.endswith("任意]") else "")
    predicate = _predicate(expression, expected)
    FACTOR_DNF_RAW.setdefault(factor_id, {}).setdefault(group_id, []).append(
        (predicate, row["kind"])
    )

CONDITION_GROUPS: dict[str, dict[int, list[tuple[str, str]]]] = {}
for row in CONDITIONS:
    CONDITION_GROUPS.setdefault(row["assertion_id"], {}).setdefault(
        int(row["condition_group_id"]), []
    ).append((row["factor"], row["expected"]))


@lru_cache(None)
def _factor_dnf(factor_id: str, seen: frozenset = frozenset()):
    """Factor DNF: set[frozenset[predicate]]；空集合表示不可满足。"""
    if factor_id not in FACTOR_DNF_RAW:
        return frozenset({frozenset({("POS", "REF_MISSING", factor_id, "1")})})
    if factor_id in seen:
        return frozenset()

    output: set[frozenset] = set()
    for _, terms in FACTOR_DNF_RAW[factor_id].items():
        branch = frozenset()
        for predicate, kind in terms:
            if kind == "factor_ref":
                # factor_ref predicates are signed as
                # ("POS", "DIRECT", referenced_factor_id, expected).
                ref_id = predicate[2]
                sub = _factor_dnf(ref_id, seen | {factor_id})
                if predicate[3] == "0":
                    sub = _complement(sub)
            else:
                sub = frozenset({frozenset({predicate})})

            next_branch: set[frozenset] = set()
            for left in branch or {frozenset()}:
                for right in sub or {frozenset()}:
                    conjunction = left | right
                    if not _has_conflict(conjunction):
                        next_branch.add(conjunction)
            branch = frozenset(next_branch)
            if not branch:
                break
        output.update(branch)
    return frozenset(output)


def _intersect(left: frozenset, right: frozenset) -> frozenset:
    output: set[frozenset] = set()
    for lft in left:
        for rgt in right:
            conjunction = lft | rgt
            if not _has_conflict(conjunction):
                output.add(conjunction)
    return frozenset(output)


def _complement(dnf: frozenset) -> frozenset:
    clauses = [frozenset(_negation(p) for p in branch) for branch in dnf]
    output = frozenset({frozenset()})
    for clause in clauses:
        next_output: set[frozenset] = set()
        for left in output:
            for predicate in clause:
                conjunction = left | {predicate}
                if not _has_conflict(conjunction):
                    next_output.add(conjunction)
        output = frozenset(next_output)
        if not output:
            break
    return output


def _conflict(left, right):
    if left == right:
        return False
    lsign, lkind, largs, lexpected = left
    rsign, rkind, rargs, rexpected = right
    if lsign == rsign:
        # A direct wildcard path is one collapsed scalar value; two distinct
        # instantiated values therefore cannot both be positive.
        if (
            lsign == "POS"
            and lkind == rkind
            and largs == rargs
            and largs
            and largs[-1] == "任意"
            and lexpected not in ("", "任意")
            and rexpected not in ("", "任意")
        ):
            return lexpected != rexpected
        return False

    # Same predicate cannot be true and false.
    if lsign != rsign and lkind == rkind and largs == rargs and lexpected != rexpected:
        return True

    # Direct gender values are mutually exclusive.
    if lkind == rkind == "DIRECT" and largs == rargs == ("gender",):
        return lexpected != rexpected

    # Engine collapses spouse palace to one state.
    if lkind == rkind == "夫妻宫状态" and largs and rargs and largs[0] != rargs[0]:
        return True

    # Day-branch special type is one collapsed value.
    if lkind == rkind == "日支类型" and largs and rargs and largs[0] != rargs[0]:
        return True

    # Same target cannot be simultaneously strong/weak or useful/avoided.
    if lkind in {"旺", "弱"} and rkind in {"旺", "弱"} and largs == rargs:
        return lexpected != rexpected
    if lkind in {"为用", "为忌"} and rkind in {"为用", "为忌"} and largs == rargs:
        return lexpected != rexpected

    # Same palace star cannot have two brightness values. Wildcard means presence.
    if (
        lkind == rkind == "宫含"
        and len(largs) >= 3
        and len(rargs) >= 3
        and largs[:2] == rargs[:2]
        and largs[2] != rargs[2]
        and "任意" not in {largs[2], rargs[2]}
    ):
        return True

    return False


def _group_dnf(factors: list[tuple[str, str]]) -> frozenset:
    """展开一个断语条件组的 AND 约束；空集表示该分支永久不可达。"""
    dnf = frozenset({frozenset()})
    for factor_id, expected in factors:
        sub = _factor_dnf(factor_id)
        if expected == "0":
            sub = _complement(sub)
        else:
            sub = _bind_scalar_expected(sub, expected)
        dnf = _intersect(dnf, sub)
        if not dnf:
            break
    return dnf


def test_all_assertion_condition_groups_are_reachable():
    unreachable = []
    assertion_ids = {row["assertion_id"] for row in ASSERTIONS}
    missing_groups = sorted(assertion_ids - set(CONDITION_GROUPS))
    assert not missing_groups, f"断语缺少静态 DNF 条件组: {missing_groups}"

    for assertion_id, groups in CONDITION_GROUPS.items():
        if not groups:
            continue
        for group_id, factors in sorted(groups.items()):
            if not _group_dnf(factors):
                unreachable.append(f"{assertion_id}#{group_id}")

    assert not unreachable, f"永久不可达断语条件组: {unreachable}"


def test_known_spouse_control_state_remains_reachable():
    """有根配偶星也可以受制；不得把得地与受克定义为互斥。"""
    group = CONDITION_GROUPS["hun_207"][1]
    factors = dict(group)
    assert factors["配偶星受克"] == "1"
    assert factors["配偶星得地"] == "1"
    assert factors["夫妻宫破"] == "1"

    dnf = frozenset({frozenset()})
    for factor_id, expected in group:
        sub = _factor_dnf(factor_id)
        if expected == "0":
            sub = _complement(sub)
        dnf = _intersect(dnf, sub)
    assert dnf, "hun_207 became permanently unreachable"


def test_static_dnf_audit_remains_sensitive_to_contradictions():
    """审计本身不得恒真：同一标量因子的两个状态必须判为不可满足。"""
    assert not _group_dnf([("身强弱", "身强"), ("身强弱", "身弱")])
    assert not _group_dnf([("性别", "male"), ("性别", "female")])
