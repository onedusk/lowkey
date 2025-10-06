#!/usr/bin/env python3
"""Shared guideline utilities for review hooks."""
from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Iterable, List

BASE_PATH = Path("docs/eng-practices")

@dataclass(frozen=True)
class Guideline:
    category: str
    title: str
    reference: str
    path: str

# Predefined guideline references (subset for illustration)
GUIDELINES: List[Guideline] = [
    Guideline(
        category="CLDescription",
        title="CL descriptions capture what and why",
        reference="docs/eng-practices/review/developer/cl-descriptions.md",
        path="docs/eng-practices/review/developer/cl-descriptions.md",
    ),
    Guideline(
        category="SmallCLs",
        title="Prefer small, self-contained changes",
        reference="docs/eng-practices/review/developer/small-cls.md",
        path="docs/eng-practices/review/developer/small-cls.md",
    ),
    Guideline(
        category="ReviewStandard",
        title="Approve when code health improves",
        reference="docs/eng-practices/review/reviewer/standard.md",
        path="docs/eng-practices/review/reviewer/standard.md",
    ),
]


def guideline_links(categories: Iterable[str]) -> List[Guideline]:
    requested = set(categories)
    return [g for g in GUIDELINES if g.category in requested]

def get_guideline_findings(tool_name: str, tool_input: dict) -> list:
    """Analyzes tool use against guidelines. Placeholder implementation."""
    # This is a placeholder. A real implementation would analyze the tool
    # input (e.g., commit message style, size of change) and return findings.
    findings = []
    if tool_name == "Edit" and len(tool_input.get("description", "")) < 10:
        findings.append({
            "guideline": "CLDescription",
            "finding": "Edit description is very short. Consider providing more detail.",
        })
    return findings
