### Project: Medianne

**One-liner**  
Medianne is a Go 1.19 digital health record / healthcare API (patients, visits, institutions, products, SatuSehat) built as a layered/clean-style backend: Chi HTTP handlers → usecases (business logic) → repos (PostgreSQL via Xorm/pgx, in-memory cache, external HTTP), with entity interfaces for repos and usecases and module-based dependency injection in `internal/`.

---

## Workflow Orchestration

### Plan Mode & Execution
* **Default to Plan:** Enter **Plan Mode** for any non-trivial task (3+ steps or architectural decisions).
* **Stop & Re-plan:** If a task deviates or fails, **STOP** immediately. Do not push through a broken plan; re-evaluate.
* **Verification Planning:** Use Plan Mode to define verification steps, not just implementation steps.
* **Detailed Specs:** Write exhaustive specifications upfront to eliminate ambiguity before coding.

### Subagent Strategy
* **Liberal Use:** Offload research and parallel analysis to subagents to keep the main context clean.
* **Focused Execution:** Assign exactly **one task per subagent** for maximum focus.
* **Compute Scaling:** For complex problems, deploy specialized subagents for sub-tasks.

---

## Core Principles

* **Simplicity First:** Every change should be as simple as possible. Minimize the code footprint.
* **No Laziness:** Identify and fix root causes. No "band-aid" fixes. Senior developer standards.
* **Minimal Impact:** Touch only what is necessary. Avoid accidental refactors to prevent regressions.
* **Demand Elegance:** For non-trivial changes, ask: *"Is there a more elegant way?"* If a fix feels hacky, implement the elegant solution. (Skip for simple/obvious fixes).

---

## Autonomous Bug Fixing

* **Ownership:** When a bug is reported, fix it. Do not ask for "hand-holding."
* **Evidence-Based:** Reference logs, error traces, and failing tests to resolve issues.
* **CI/CD Responsibility:** Proactively fix failing CI tests without being told how.
* **Zero Context Switching:** Aim to complete the fix with zero additional user input.

---

## Verification & Quality Control

* **Proof of Work:** Never mark a task complete without proving it works via tests or logs.
* **The Staff Engineer Test:** Before presenting, ask: *"Would a Staff Engineer approve this?"*
* **Differential Review:** Diff behavior between the original state and changes when relevant.
