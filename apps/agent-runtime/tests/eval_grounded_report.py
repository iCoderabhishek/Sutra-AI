"""
Eval 1: grounded report.

The agent researches a made-up company. web_search is stubbed with fixed
results, so anything in the report that isn't in those results was invented.
The LLM is real (OpenRouter), every tool is fake: no Tavily, Redis or Telegram.

Checks per run:
  - web_search was called
  - the first final reply was valid report JSON (no "Formatting report" repair)
  - outcome is "completed" and the report is not the fallback
  - every source URL is one of the stubbed URLs
  - every dollar amount in the report appears in the stubbed results
  - at least one real stubbed price is reported

Run from apps/agent-runtime (so .env is picked up):
    uv run python tests/eval_grounded_report.py
    uv run python tests/eval_grounded_report.py --runs 5 --model <openrouter-model-id>
"""

import argparse
import asyncio
import json
import os
import re
import sys

sys.path.append(os.path.join(os.path.dirname(os.path.dirname(__file__)), "src"))

import agent.loop as loop  # noqa: E402

# A fictional company, so the model can't answer from training data.
GOAL = "What does Lumora Analytics currently charge for its plans? Summarise the pricing."

STUB_RESULTS = [
    {
        "url": "https://lumora-analytics.example/pricing",
        "content": (
            "Lumora Analytics pricing. Starter plan: $29 per month, up to 3 dashboards. "
            "Growth plan: $79 per month, unlimited dashboards and SSO. "
            "Enterprise: custom pricing, contact sales."
        ),
    },
    {
        "url": "https://techwire.example/2026/10/lumora-raises-prices",
        "content": (
            "8 Oct 2026. Lumora Analytics raised its Growth plan from $69 to $79 per month "
            "this week. The Starter plan stays at $29."
        ),
    },
    {
        "url": "https://reviews.example/lumora-analytics",
        "content": (
            "Lumora Analytics review: reviewers like the Growth plan's SSO support but "
            "say the Starter plan's 3-dashboard limit is tight for teams."
        ),
    },
]

STUB_URLS = {r["url"] for r in STUB_RESULTS}
STUB_TEXT = " ".join(r["content"] for r in STUB_RESULTS)
DOLLAR = re.compile(r"\$\s?\d[\d,]*(?:\.\d+)?")


def _dollars(text: str) -> set[str]:
    return {re.sub(r"[\s,]", "", m) for m in DOLLAR.findall(text)}


STUB_DOLLARS = _dollars(STUB_TEXT)  # {"$29", "$79", "$69"}


async def fake_dispatch(name: str, args: dict, allowed=None) -> str:
    """Stands in for agent.tools.dispatch_tool and records every call."""
    fake_dispatch.calls.append((name, args))
    if name == "web_search":
        return "\n\n".join(f"Source: {r['url']}\nContent: {r['content']}" for r in STUB_RESULTS)
    return f"Error: Tool '{name}' is not available to this agent. Available tools: ['web_search']"


fake_dispatch.calls = []


def check(events: list[dict], calls: list[tuple]) -> tuple[list[str], dict]:
    """Return (failures, report_dict). An empty failure list means the run passed."""
    failures = []
    steps = [e.get("step") for e in events]

    errors = [e for e in events if e.get("status") == "error"]
    if errors:
        failures.append(f"run error: {errors[-1].get('step')} {errors[-1].get('content', '')}"[:160])

    if not any(name == "web_search" for name, _ in calls):
        failures.append("web_search never called")

    if "Formatting report" in steps:
        failures.append("first final reply was not valid report JSON")

    final = next((e for e in reversed(events) if e.get("step") == "Final Answer"), None)
    if not final:
        failures.append("no final answer")
        return failures, {}

    report = json.loads(final["content"])

    if report.get("headline") == "Result could not be structured":
        failures.append("fell back to unstructured report")
    if report.get("outcome") != "completed":
        failures.append(f"outcome is {report.get('outcome')!r}, expected 'completed'")

    sources = report.get("sources", [])
    if not sources:
        failures.append("no sources listed")
    for s in sources:
        if s.get("url") not in STUB_URLS:
            failures.append(f"source not from tools: {s.get('url')!r}")

    report_text = json.dumps(report)
    invented = _dollars(report_text) - STUB_DOLLARS
    if invented:
        failures.append(f"prices not in tool results: {sorted(invented)}")
    if not _dollars(report_text) & {"$29", "$79"}:
        failures.append("current prices ($29 / $79) not reported")

    return failures, report


async def run_once(model: str | None) -> dict:
    events: list[dict] = []
    fake_dispatch.calls = []

    async def collect(event: dict):
        events.append(event)

    await loop.run_agent(
        goal=GOAL,
        stream_callback=collect,
        tools=["web_search"],
        model=model,
    )

    failures, report = check(events, list(fake_dispatch.calls))
    last_cost = next((e["cost"] for e in reversed(events) if "cost" in e), {})
    return {
        "passed": not failures,
        "failures": failures,
        "model": last_cost.get("model", "?"),
        "tokens": last_cost.get("total_tokens", 0),
        "iterations": max((e.get("iteration", 0) for e in events), default=0),
        "headline": report.get("headline", ""),
    }


async def main():
    parser = argparse.ArgumentParser(description="Eval: grounded report")
    parser.add_argument("--runs", type=int, default=3)
    parser.add_argument("--model", default=None, help="OpenRouter model id; default uses the free-model fallback list")
    args = parser.parse_args()

    # Patch the name the loop actually calls, not agent.tools.dispatch_tool.
    loop.dispatch_tool = fake_dispatch

    results = []
    for i in range(args.runs):
        print(f"run {i + 1}/{args.runs} ...", flush=True)
        results.append(await run_once(args.model))

    print("\nEval 1: grounded report")
    print(f"{'run':<4} {'result':<7} {'iters':<6} {'tokens':<7} model")
    for i, r in enumerate(results, 1):
        print(f"{i:<4} {'PASS' if r['passed'] else 'FAIL':<7} {r['iterations']:<6} {r['tokens']:<7} {r['model']}")
        if r["headline"]:
            print(f"     headline: {r['headline']}")
        for f in r["failures"]:
            print(f"     - {f}")

    passed = sum(r["passed"] for r in results)
    needed = args.runs // 2 + 1
    verdict = "PASS" if passed >= needed else "FAIL"
    print(f"\n{verdict}: {passed}/{args.runs} runs passed (need {needed})")
    sys.exit(0 if verdict == "PASS" else 1)


if __name__ == "__main__":
    asyncio.run(main())
