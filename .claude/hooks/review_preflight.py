#!/usr/bin/env python3
"""Review preflight hook with guideline hints and test execution."""
import json
import os
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

# Add hook's parent directory to path to allow importing shared modules
sys.path.append(str(Path(__file__).parent))
from review_guidelines import get_guideline_findings

def run_tests(cwd: Path) -> dict:
    """Runs the test suite via a shell script and returns the outcome."""
    script_path = cwd / "scripts/run_make_test.sh"
    if not script_path.is_file():
        return {"status": "skipped", "reason": "run_make_test.sh not found"}

    try:
        result = subprocess.run(
            [str(script_path)],
            capture_output=True,
            text=True,
            cwd=cwd,
            check=False, # Don't raise exception on non-zero exit
        )
        return {
            "status": "completed",
            "exit_code": result.returncode,
            "stdout": result.stdout[-500:], # Capture last 500 chars
            "stderr": result.stderr[-500:],
        }
    except Exception as e:
        return {"status": "error", "reason": str(e)}

def main():
    """Parses PreToolUse input, logs it, and runs tests."""
    if os.isatty(sys.stdin.fileno()):
        return

    try:
        input_data = json.load(sys.stdin)
        session_id = input_data.get("session_id", "unknown-session")
        hook_event = input_data.get("event", "PreToolUse")
        tool_name = input_data.get("tool_name", "UnknownTool")
        cwd = Path(input_data.get("cwd", "."))
        tool_input = input_data.get("tool_input", {})

        # --- Data Extraction ---
        file_path = tool_input.get("file_path")
        description = tool_input.get("description", "")

        extra = {
            "tool_input_keys": list(tool_input.keys()),
            "description": description,
        }

        # --- Guideline Analysis ---
        findings = get_guideline_findings(tool_name, tool_input)

        # --- Test Execution ---
        test_results = None
        if tool_name in ["Edit", "MultiEdit", "Write", "FilePatch"]:
            test_results = run_tests(cwd)

        if test_results and test_results.get("exit_code", 0) != 0:
            reason = (
                "Action blocked: `make test` failed. Please fix tests before proceeding.\n\n"
                f"Stderr:\n{test_results.get('stderr', '(no stderr)')}"
            )
            print(json.dumps({
                "permissionDecision": "denied",
                "permissionDecisionReason": reason,
            }))
            return # Block the action and exit

        # --- Logging ---
        log_dir = cwd / ".claude/context/review-audit"
        log_dir.mkdir(exist_ok=True)
        log_file = log_dir / "preflight.jsonl"

        log_entry = {
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "session_id": session_id,
            "hook_event": hook_event,
            "tool_name": tool_name,
            "cwd": str(cwd),
            "file_path": file_path,
            "extra": extra,
            "findings": findings,
            "test_results": test_results,
        }

        with open(log_file, "a") as f:
            f.write(json.dumps(log_entry) + "\n")

    except json.JSONDecodeError:
        pass
    except Exception:
        pass

if __name__ == "__main__":
    main()