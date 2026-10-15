from __future__ import annotations

import re
import subprocess
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]
GITHUB_EXPRESSION = re.compile(r"\$\{\{[^}]+\}\}")


def test_workflow_run_scripts_are_valid_bash():
    for path in sorted((ROOT / ".github/workflows").glob("*.yml")):
        workflow = yaml.safe_load(path.read_text(encoding="utf-8"))
        jobs = workflow.get("jobs") or {}
        assert isinstance(jobs, dict)
        for job_name, job in jobs.items():
            for step_number, step in enumerate(job.get("steps") or []):
                script = step.get("run")
                if not script:
                    continue
                placeholder_script = GITHUB_EXPRESSION.sub("__GITHUB_EXPRESSION__", script)
                result = subprocess.run(
                    ["bash", "-n"],
                    input=placeholder_script,
                    text=True,
                    capture_output=True,
                    check=False,
                )
                assert result.returncode == 0, (
                    f"invalid bash in {path.name}:job={job_name}:step={step_number}\n"
                    f"{result.stderr}"
                )
