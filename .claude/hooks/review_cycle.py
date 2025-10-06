#!/usr/bin/env python3
"""Review cycle hook with guideline hints."""
import json
import os
import sys
from datetime import datetime, timezone
from pathlib import Path

# Add hook's parent directory to path to allow importing shared modules
sys.path.append(str(Path(__file__).parent))
from review_guidelines import get_guideline_findings

def main():
    """Parses PostToolUse input and logs it with guideline findings."""
    if os.isatty(sys.stdin.fileno()):
        return

    try:
        input_data = json.load(sys.stdin)
        session_id = input_data.get("session_id", "unknown-session")
        hook_event = input_data.get("event", "PostToolUse")
        tool_name = input_data.get("tool_name", "UnknownTool")
        cwd = input_data.get("cwd", ".")
        tool_input = input_data.get("tool_input", {})
        tool_response = input_data.get("tool_response", {})

        # --- Data Extraction ---
        file_path = tool_input.get("file_path") or tool_response.get("filePath")
        success = tool_response.get("success", False)

        summary = {
            "tool_input_keys": list(tool_input.keys()),
            "tool_response_keys": list(tool_response.keys()),
            "file_path": file_path,
            "success": success,
        }

        # --- Guideline Analysis ---
        findings = get_guideline_findings(tool_name, tool_input)

        # --- Logging ---
        log_dir = Path(cwd) / ".claude/context/review-audit"
        log_dir.mkdir(exist_ok=True)
        log_file = log_dir / "cycle.jsonl"

        log_entry = {
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "session_id": session_id,
            "hook_event": hook_event,
            "tool_name": tool_name,
            "cwd": cwd,
            "summary": summary,
            "findings": findings,
        }

        with open(log_file, "a") as f:
            f.write(json.dumps(log_entry) + "\n")

    except json.JSONDecodeError:
        # Non-json input, ignore
        pass
    except Exception:
        # For safety, hooks should not crash the main process
        pass

if __name__ == "__main__":
    main()