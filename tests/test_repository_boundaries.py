#!/usr/bin/env python3
"""Keep this repository a skill distribution, not service source."""

import pathlib
import subprocess
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[1]


class RepositoryBoundaries(unittest.TestCase):
    def test_tracks_only_skill_distribution(self):
        tracked = subprocess.check_output(
            ["git", "-C", str(ROOT), "ls-files"], text=True
        ).splitlines()
        allowed_prefixes = (
            ".github/", ".gitignore", "skills/", "tests/", "Makefile", "LICENSE",
            "assets/", "distribution/", "scripts/", "tools/", "README.md", "README.zh.md",
        )
        allowed_files = {
            ".markdownlint-cli2.jsonc",
            ".nvmrc",
            "package.json",
            "package-lock.json",
            "requirements-dev.txt",
        }
        for path in tracked:
            if path in allowed_files:
                continue
            self.assertTrue(
                path.startswith(allowed_prefixes),
                f"implementation or deployment source crossed skill boundary: {path}",
            )


if __name__ == "__main__":
    unittest.main()
