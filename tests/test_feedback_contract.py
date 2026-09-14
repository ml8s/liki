"""Contract tests for the autonomous skill feedback payload."""

import json
import pathlib
import unittest

from jsonschema import Draft202012Validator
from jsonschema.exceptions import ValidationError


ROOT = pathlib.Path(__file__).resolve().parents[1]
SKILLS = ["liki-bazi", "liki-divination", "liki-fengshui", "liki-naming"]
SKILLS_DIR = ROOT / "skills"


class TestFeedbackContract(unittest.TestCase):
    def test_all_skill_schemas_are_identical(self):
        digests = set()
        for skill in SKILLS:
            raw = (SKILLS_DIR / skill / "feedback.schema.json").read_bytes()
            digests.add(raw)
        self.assertEqual(len(digests), 1)

    def test_schema_compiles_and_accepts_minimal_payload(self):
        schema = json.loads((SKILLS_DIR / "liki-bazi" / "feedback.schema.json").read_text(encoding="utf-8"))
        validator = Draft202012Validator(schema)
        payload = {
            "schema_version": "feedback-v1",
            "meta": {
                "source": "skill-agent",
                "skill": "liki-bazi",
                "skill_version": "2026.09.15.1",
                "engine_version": "2026.09.15.1",
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

    def test_schema_keeps_v1_minimal(self):
        schema = json.loads((SKILLS_DIR / "liki-bazi" / "feedback.schema.json").read_text(encoding="utf-8"))
        self.assertEqual(set(schema["required"]), {"schema_version", "meta", "agent", "llm", "problem"})
        self.assertEqual(set(schema["properties"]), {"schema_version", "meta", "agent", "llm", "problem"})
        for group, allowed in {
            "meta": {"source", "skill", "skill_version", "engine_version", "session_hash"},
            "agent": {"name", "version"},
            "llm": {"provider", "model", "model_version"},
            "problem": {"type", "severity", "tool", "summary", "expected", "observed", "dedup_hash"},
        }.items():
            self.assertEqual(set(schema["properties"][group]["properties"]), allowed)

    def test_schema_rejects_unneeded_diagnostic_groups(self):
        schema = json.loads((SKILLS_DIR / "liki-bazi" / "feedback.schema.json").read_text(encoding="utf-8"))
        validator = Draft202012Validator(schema)
        payload = {
            "schema_version": "feedback-v1",
            "meta": {
                "source": "skill-agent",
                "skill": "liki-bazi",
                "skill_version": "2026.09.15.1",
                "engine_version": "2026.09.15.1",
            },
            "agent": {"name": "unknown", "version": "unknown"},
            "llm": {"provider": "unknown", "model": "unknown"},
            "rpc": {"method": "ziwei.liuyue"},
            "problem": {"type": "clarity", "severity": "low", "summary": "ok"},
        }
        with self.assertRaises(ValidationError):
            validator.validate(payload)

    def test_schema_rejects_unknown_and_oversized_content(self):
        schema = json.loads((SKILLS_DIR / "liki-bazi" / "feedback.schema.json").read_text(encoding="utf-8"))
        validator = Draft202012Validator(schema)
        with self.assertRaises(ValidationError):
            validator.validate({
                "schema_version": "feedback-v1",
                "meta": {},
                "problem": {"type": "clarity", "severity": "low", "summary": "ok"},
            })
        with self.assertRaises(ValidationError):
            validator.validate({
                "schema_version": "feedback-v1",
                "meta": {
                    "source": "skill-agent",
                    "skill": "liki-bazi",
                    "skill_version": "2026.09.15.1",
                    "engine_version": "2026.09.15.1",
                },
                "conversation": "private user text",
                "problem": {"type": "clarity", "severity": "low", "summary": "ok"},
            })

    def test_all_skills_publish_feedback_v1_policy(self):
        for skill in SKILLS:
            with self.subTest(skill=skill):
                text = (SKILLS_DIR / skill / "SKILL.md").read_text(encoding="utf-8")
                self.assertIn("https://liki.hk/api/feedback", text)
                self.assertIn("feedback-v1", text)
                for group in ("meta", "agent", "llm", "problem"):
                    self.assertIn(group, text)
                for issue_type in ("error", "gap", "conflict", "friction", "clarity"):
                    self.assertIn(issue_type, text)


if __name__ == "__main__":
    unittest.main()
