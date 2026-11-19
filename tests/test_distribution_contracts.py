#!/usr/bin/env python3
"""Keep the skill installer pointed at its supported free and pro surfaces."""

import json
import pathlib
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[1]


class DistributionContracts(unittest.TestCase):
    def test_free_skill_uses_free_surfaces(self):
        declaration = json.loads((ROOT / "skills/liki/.mcp.json").read_text(encoding="utf-8"))
        servers = declaration["mcpServers"]
        self.assertEqual(list(servers), ["engine", "counsel"])
        self.assertEqual(servers["engine"]["url"], "https://mcp.tokflux.com/engine/mcp")
        self.assertEqual(servers["counsel"]["url"], "https://mcp.tokflux.com/counsel/mcp")

    def test_readmes_direct_users_to_current_endpoint_and_upgrade(self):
        for name in ("README.md", "README.zh.md"):
            text = (ROOT / name).read_text(encoding="utf-8")
            self.assertIn("https://mcp.tokflux.com/engine/mcp", text)
            self.assertIn("https://mcp.tokflux.com/counsel/mcp", text)
            self.assertIn("https://mcp.tokflux.com/mcp", text)
            self.assertIn("npx skills add ml8s/liki -y", text)

    def test_feedback_includes_upgrade_prompt(self):
        skill = (ROOT / "skills/liki/SKILL.md").read_text(encoding="utf-8")
        operations = (ROOT / "skills/liki/references/ops-FEEDBACK.md").read_text(encoding="utf-8")
        self.assertIn("npx skills add ml8s/liki -y", skill)
        self.assertIn("npx skills add ml8s/liki -y", operations)


if __name__ == "__main__":
    unittest.main()
