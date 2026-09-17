"""Contract tests for the autonomous skill feedback payload."""

import json
import re
import unittest

from helpers import ROOT, SKILL_ROOT, skill_version
from jsonschema import Draft202012Validator
from jsonschema.exceptions import ValidationError


SKILLS_DIR = SKILL_ROOT


class TestFeedbackContract(unittest.TestCase):
    def test_all_skill_schemas_are_identical(self):
        digests = set()
        raw = (SKILLS_DIR / "feedback.schema.json").read_bytes()
        self.assertEqual(len({raw}), 1)

    def test_schema_compiles_and_accepts_minimal_payload(self):
        schema = json.loads((SKILLS_DIR / "feedback.schema.json").read_text(encoding="utf-8"))
        validator = Draft202012Validator(schema)
        payload = {
            "schema_version": "feedback-v1",
            "meta": {
                "source": "skill-agent",
                "skill": "liki",
                "skill_version": skill_version(),
                "engine_version": skill_version(),
            },
            "agent": {"name": "codex-cli", "version": "0.21.6"},
            "llm": {"provider": "openai", "model": "gpt-5.1"},
            "problem": {
                "type": "clarity",
                "severity": "medium",
                "tool": "ziwei.liuyue",
                "summary": "resolved_period usage is not obvious from the result schema",
                "expected": "The output contract identifies the field to explain",
                "observed": "The agent had to infer it from documentation",
            },
        }
        validator.validate(payload)

    def test_schema_pins_meta_skill_to_unified_product(self):
        schema = json.loads((SKILLS_DIR / "feedback.schema.json").read_text(encoding="utf-8"))
        self.assertEqual(schema["properties"]["meta"]["properties"]["skill"], {"const": "liki"})

        validator = Draft202012Validator(schema)
        valid = {
            "schema_version": "feedback-v1",
            "meta": {
                "source": "skill-agent",
                "skill": "liki",
                "skill_version": skill_version(),
                "engine_version": skill_version(),
            },
            "agent": {"name": "unknown", "version": "unknown"},
            "llm": {"provider": "unknown", "model": "unknown"},
            "problem": {"type": "clarity", "severity": "low", "summary": "valid problem"},
        }
        validator.validate(valid)
        for old_skill in ("liki-bazi", "liki-divination", "liki-fengshui", "liki-naming"):
            invalid = dict(valid)
            invalid["meta"] = valid["meta"] | {"skill": old_skill}
            with self.assertRaises(ValidationError):
                validator.validate(invalid)

    def test_schema_keeps_v1_minimal(self):
        schema = json.loads((SKILLS_DIR / "feedback.schema.json").read_text(encoding="utf-8"))
        self.assertEqual(set(schema["required"]), {"schema_version", "meta", "agent", "llm", "problem"})
        self.assertEqual(set(schema["properties"]), {"schema_version", "meta", "agent", "llm", "problem"})
        for group, allowed in {
            "meta": {"source", "skill", "skill_version", "engine_version", "session_hash"},
            "agent": {"name", "version"},
            "llm": {"provider", "model", "model_version"},
            "problem": {"type", "severity", "tool", "summary", "expected", "observed"},
        }.items():
            self.assertEqual(set(schema["properties"][group]["properties"]), allowed)

    def test_schema_rejects_unneeded_diagnostic_groups(self):
        schema = json.loads((SKILLS_DIR / "feedback.schema.json").read_text(encoding="utf-8"))
        validator = Draft202012Validator(schema)
        payload = {
            "schema_version": "feedback-v1",
            "meta": {
                "source": "skill-agent",
                "skill": "liki",
                "skill_version": skill_version(),
                "engine_version": skill_version(),
            },
            "agent": {"name": "unknown", "version": "unknown"},
            "llm": {"provider": "unknown", "model": "unknown"},
            "rpc": {"method": "ziwei.liuyue"},
            "problem": {"type": "clarity", "severity": "low", "summary": "valid problem"},
        }
        with self.assertRaises(ValidationError):
            validator.validate(payload)

    def test_schema_enforces_text_boundaries(self):
        schema = json.loads((SKILLS_DIR / "feedback.schema.json").read_text(encoding="utf-8"))
        validator = Draft202012Validator(schema)
        payload = {
            "schema_version": "feedback-v1",
            "meta": {
                "source": "skill-agent",
                "skill": "liki",
                "skill_version": skill_version(),
                "engine_version": skill_version(),
            },
            "agent": {"name": "unknown", "version": "unknown"},
            "llm": {"provider": "unknown", "model": "unknown"},
            "problem": {"type": "clarity", "severity": "low", "summary": "valid problem"},
        }

        boundaries = {
            ("problem", "summary"): (3, 240),
            ("problem", "expected"): (0, 500),
            ("problem", "observed"): (0, 500),
        }
        for (group, field), (valid_length, max_length) in boundaries.items():
            valid = json.loads(json.dumps(payload))
            valid[group][field] = "x" * valid_length
            validator.validate(valid)

            invalid = json.loads(json.dumps(payload))
            invalid[group][field] = "x" * (max_length + 1)
            with self.assertRaises(ValidationError):
                validator.validate(invalid)

    def test_docs_example_matches_feedback_schema(self):
        schema = json.loads((SKILLS_DIR / "feedback.schema.json").read_text(encoding="utf-8"))
        text = (ROOT / "docs" / "FEEDBACK_MODEL.md").read_text(encoding="utf-8")
        blocks = re.findall(r"```json\n(.*?)\n```", text, flags=re.DOTALL)
        self.assertTrue(blocks)
        for block in blocks:
            payload = json.loads(block)
            Draft202012Validator(schema).validate(payload)

    def test_unified_skill_documents_feedback_runtime_governance(self):
        text = (SKILLS_DIR / "SKILL.md").read_text(encoding="utf-8")
        self.assertIn("LIKI_FEEDBACK_URL", text)
        self.assertIn("LIKI_FEEDBACK_DISABLED=1", text)
        self.assertIn("同一会话最多 3 条", text)
        self.assertIn("Feedback: submitted|disabled|failed", text)
        self.assertIn("不向用户请求确认", text)
        self.assertIn("禁止用户原文", text)
        self.assertIn("出生数据", text)
        self.assertNotIn("LIKI_FEEDBACK_CONTEXT", text)

    def test_schema_rejects_unknown_and_oversized_content(self):
        schema = json.loads((SKILLS_DIR / "feedback.schema.json").read_text(encoding="utf-8"))
        validator = Draft202012Validator(schema)
        with self.assertRaises(ValidationError):
            validator.validate({
                "schema_version": "feedback-v1",
                "meta": {},
                "problem": {"type": "clarity", "severity": "low", "summary": "valid problem"},
            })
        with self.assertRaises(ValidationError):
            validator.validate({
                "schema_version": "feedback-v1",
                "meta": {
                    "source": "skill-agent",
                    "skill": "liki",
                    "skill_version": skill_version(),
                    "engine_version": skill_version(),
                },
                "conversation": "private user text",
                "problem": {"type": "clarity", "severity": "low", "summary": "valid problem"},
            })

    def test_unified_skill_publishes_feedback_v1_policy(self):
        text = (SKILLS_DIR / "SKILL.md").read_text(encoding="utf-8")
        self.assertIn("https://liki.hk/api/feedback", text)
        self.assertIn("feedback-v1", text)
        for group in ("meta", "agent", "llm", "problem"):
            self.assertIn(group, text)
        schema = json.loads((SKILLS_DIR / "feedback.schema.json").read_text(encoding="utf-8"))
        for issue_type in schema["properties"]["problem"]["properties"]["type"]["enum"]:
            self.assertIn(issue_type, text)


if __name__ == "__main__":
    unittest.main()
