#!/usr/bin/env python3
"""Analyze Codex canary prompt snapshots, sessions, labels, and gates."""

from __future__ import annotations

import argparse
import hashlib
import json
import math
import os
import re
import statistics
import sys
from datetime import datetime, timezone


def read_json(path: str):
    with open(path, encoding="utf-8") as handle:
        return json.load(handle)


def read_jsonl(path: str) -> list[dict]:
    if not path or not os.path.exists(path):
        return []
    rows = []
    with open(path, encoding="utf-8") as handle:
        for number, line in enumerate(handle, 1):
            if not line.strip():
                continue
            try:
                rows.append(json.loads(line))
            except json.JSONDecodeError as error:
                raise ValueError(f"{path}:{number}: {error}") from error
    return rows


def emit(value: dict) -> None:
    json.dump(value, sys.stdout, indent=2, sort_keys=True)
    sys.stdout.write("\n")


def prompt_summary(path: str, router: str, workspace: str) -> dict:
    prompt = read_json(path)
    blocks: list[str] = []
    for item in prompt:
        if not isinstance(item, dict):
            continue
        for content in item.get("content", []):
            if not isinstance(content, dict):
                continue
            text = content.get("text")
            if isinstance(text, str) and "<skills_instructions>" in text:
                blocks.append(text)

    router_md = os.path.realpath(os.path.join(router, "SKILL.md"))
    aliases: dict[str, str] = {}
    for block in blocks:
        for line in block.splitlines():
            match = re.match(r"^- `([^`]+)` = `([^`]+)`$", line)
            if match:
                aliases[match.group(1)] = match.group(2)
    skills = []
    for block in blocks:
        for line in block.splitlines():
            if not line.startswith("- ") or " (file: " not in line or not line.endswith(")"):
                continue
            identity_and_description, skill_path = line[2:-1].rsplit(" (file: ", 1)
            if ": " not in identity_and_description:
                continue
            identity, description = identity_and_description.split(": ", 1)
            expanded_path = skill_path
            alias, separator, remainder = skill_path.partition("/")
            if separator and alias in aliases:
                expanded_path = os.path.join(aliases[alias], remainder)
            real_path = os.path.realpath(expanded_path)
            system = "/skills/.system/" in expanded_path.replace("\\", "/")
            managed_router = real_path == router_md
            skills.append(
                {
                    "id": identity,
                    "description": description,
                    "path": skill_path,
                    "resolved_path": expanded_path,
                    "system": system,
                    "managed_router": managed_router,
                    "metadata_bytes": len(line.encode("utf-8")),
                    "description_bytes": len(description.encode("utf-8")),
                }
            )

    non_system = [skill for skill in skills if not skill["system"]]
    unexpected = [skill["id"] for skill in non_system if not skill["managed_router"]]
    encoded_prompt = json.dumps(prompt, separators=(",", ":"), ensure_ascii=False).encode("utf-8")
    return {
        "workspace": os.path.realpath(workspace),
        "prompt_input_bytes": len(encoded_prompt),
        "skills_block_bytes": sum(len(block.encode("utf-8")) for block in blocks),
        "skill_count": len(skills),
        "system_skill_count": sum(skill["system"] for skill in skills),
        "non_system_skill_count": len(non_system),
        "skill_ids": [skill["id"] for skill in skills],
        "non_system_skill_ids": [skill["id"] for skill in non_system],
        "non_system_metadata_bytes": sum(skill["metadata_bytes"] for skill in non_system),
        "non_system_description_bytes": sum(skill["description_bytes"] for skill in non_system),
        "router_present": any(skill["managed_router"] for skill in skills),
        "unexpected_non_system_skills": unexpected,
    }


def reduction(baseline: int, canary: int) -> float | None:
    if baseline <= 0:
        return None
    return 1.0 - canary / baseline


def compare_summary(baseline_path: str, canary_path: str, router: str, workspace: str) -> dict:
    baseline = prompt_summary(baseline_path, router, workspace)
    canary = prompt_summary(canary_path, router, workspace)
    return {
        "version": 1,
        "captured_at": datetime.now(timezone.utc).isoformat(),
        "workspace": os.path.realpath(workspace),
        "baseline": baseline,
        "canary": canary,
        "non_system_metadata_reduction": reduction(
            baseline["non_system_metadata_bytes"], canary["non_system_metadata_bytes"]
        ),
        "non_system_description_reduction": reduction(
            baseline["non_system_description_bytes"], canary["non_system_description_bytes"]
        ),
    }


def session_summary(args: argparse.Namespace) -> dict:
    events = read_jsonl(args.events)
    routes = read_jsonl(args.routes)
    usage = {}
    thread_id = ""
    terminal = "missing"
    for event in events:
        if event.get("type") == "thread.started":
            thread_id = event.get("thread_id", thread_id)
        if event.get("type") == "turn.completed":
            terminal = "completed"
            usage = event.get("usage", {})
        elif event.get("type") == "turn.failed":
            terminal = "failed"

    candidate_ids: list[str] = []
    warnings: list[str] = []
    route_ids: list[str] = []
    for route in routes:
        if route.get("route_id"):
            route_ids.append(route["route_id"])
        for candidate in route.get("candidates", []):
            identity = candidate.get("id")
            if identity and identity not in candidate_ids:
                candidate_ids.append(identity)
        for warning in route.get("warnings", []):
            if warning not in warnings:
                warnings.append(warning)

    expected = [value for value in args.expected.split(",") if value]
    expected_in_candidates = None
    if expected:
        expected_in_candidates = all(value in candidate_ids for value in expected)
    prompt_hash = hashlib.sha256(" ".join(args.prompt.split()).lower().encode("utf-8")).hexdigest()
    return {
        "version": 1,
        "session_id": args.session,
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "workspace": os.path.realpath(args.workspace),
        "prompt_hash": f"sha256:{prompt_hash}",
        "expected": expected,
        "expected_in_candidates": expected_in_candidates,
        "codex_exit_code": args.exit_code,
        "terminal_event": terminal,
        "thread_id": thread_id,
        "usage": usage,
        "route_count": len(routes),
        "route_ids": route_ids,
        "candidate_ids": candidate_ids,
        "warnings": warnings,
        "events_path": os.path.realpath(args.events),
        "routes_path": os.path.realpath(args.routes),
    }


def percentile(values: list[int], quantile: float) -> int | None:
    if not values:
        return None
    ordered = sorted(values)
    return ordered[max(0, math.ceil(quantile * len(ordered)) - 1)]


def report_summary(args: argparse.Namespace) -> tuple[dict, bool]:
    sessions = read_jsonl(args.sessions)
    labels = read_jsonl(args.labels)
    latest_labels = {row["session_id"]: row for row in labels}
    labeled = [latest_labels[row["session_id"]] for row in sessions if row["session_id"] in latest_labels]

    task_scores = {"pass": 1.0, "partial": 0.5, "fail": 0.0}
    routing_scores = {"correct": 1.0, "abstain": 1.0, "partial": 0.5, "wrong": 0.0}
    task_score = statistics.fmean(task_scores[row["task"]] for row in labeled) if labeled else None
    routing_score = statistics.fmean(routing_scores[row["routing"]] for row in labeled) if labeled else None
    route_coverage = (
        sum(row.get("route_count", 0) > 0 for row in sessions) / len(sessions) if sessions else None
    )
    input_tokens = [
        int(row.get("usage", {}).get("input_tokens", 0))
        for row in sessions
        if row.get("usage", {}).get("input_tokens") is not None
    ]
    cached_tokens = [
        int(row.get("usage", {}).get("cached_input_tokens", 0))
        for row in sessions
        if row.get("usage", {}).get("cached_input_tokens") is not None
    ]

    context = read_json(args.context) if os.path.exists(args.context) else {}
    metadata_reduction = context.get("non_system_metadata_reduction")
    unexpected = context.get("canary", {}).get("unexpected_non_system_skills")
    gates = {
        "sessions_between_20_and_30": 20 <= len(sessions) <= 30,
        "trustworthy_sessions_at_least_30": len(labeled) >= 30,
        "all_sessions_labeled": bool(sessions) and len(labeled) == len(sessions),
        "task_score_at_least_0_85": task_score is not None and task_score >= 0.85,
        "routing_score_at_least_0_90": routing_score is not None and routing_score >= 0.90,
        "route_coverage_at_least_0_90": route_coverage is not None and route_coverage >= 0.90,
        "metadata_reduction_at_least_0_80": metadata_reduction is not None and metadata_reduction >= 0.80,
        "no_unexpected_non_system_skills": unexpected == [],
    }
    gates["activation_ready"] = all(gates.values())
    report = {
        "version": 1,
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "verdict": "go" if gates["activation_ready"] else "hold",
        "sessions": len(sessions),
        "labeled_sessions": len(labeled),
        "task_score": task_score,
        "routing_score": routing_score,
        "route_coverage": route_coverage,
        "context": context,
        "usage": {
            "input_tokens_total": sum(input_tokens),
            "input_tokens_p50": percentile(input_tokens, 0.50),
            "input_tokens_p95": percentile(input_tokens, 0.95),
            "cached_input_tokens_total": sum(cached_tokens),
        },
        "gates": gates,
    }
    return report, gates["activation_ready"]


def main() -> int:
    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers(dest="command", required=True)

    prompt = subparsers.add_parser("prompt")
    prompt.add_argument("--input", required=True)
    prompt.add_argument("--router", required=True)
    prompt.add_argument("--workspace", required=True)

    compare = subparsers.add_parser("compare")
    compare.add_argument("--baseline", required=True)
    compare.add_argument("--canary", required=True)
    compare.add_argument("--router", required=True)
    compare.add_argument("--workspace", required=True)

    session = subparsers.add_parser("session")
    session.add_argument("--events", required=True)
    session.add_argument("--routes", required=True)
    session.add_argument("--session", required=True)
    session.add_argument("--workspace", required=True)
    session.add_argument("--expected", default="")
    session.add_argument("--prompt", required=True)
    session.add_argument("--exit-code", type=int, required=True)

    label = subparsers.add_parser("label")
    label.add_argument("--session", required=True)
    label.add_argument("--task", choices=("pass", "partial", "fail"), required=True)
    label.add_argument("--routing", choices=("correct", "partial", "wrong", "abstain"), required=True)
    label.add_argument("--note", default="")

    report = subparsers.add_parser("report")
    report.add_argument("--sessions", required=True)
    report.add_argument("--labels", required=True)
    report.add_argument("--context", required=True)
    report.add_argument("--check-gates", action="store_true")

    args = parser.parse_args()
    try:
        if args.command == "prompt":
            emit(prompt_summary(args.input, args.router, args.workspace))
        elif args.command == "compare":
            emit(compare_summary(args.baseline, args.canary, args.router, args.workspace))
        elif args.command == "session":
            emit(session_summary(args))
        elif args.command == "label":
            emit(
                {
                    "version": 1,
                    "session_id": args.session,
                    "timestamp": datetime.now(timezone.utc).isoformat(),
                    "task": args.task,
                    "routing": args.routing,
                    "note": args.note,
                }
            )
        elif args.command == "report":
            report_value, ready = report_summary(args)
            emit(report_value)
            if args.check_gates and not ready:
                return 1
    except (OSError, ValueError, KeyError) as error:
        print(error, file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
