---
name: baron-engine
description: Native Codex bridge to the canonical Baron Core lifecycle and selected resources.
---

# Baron Engine Codex Bridge

Baron Core is the project orchestration authority and semantic source of truth. The root `AGENTS.md` contract owns automatic lifecycle behavior.

For a meaningful user task, construct `PrepareRequestV1` as structured data and invoke `baron control-plane prepare --adapter codex --json`; pass task text through the supported structured transport, never through shell interpolation. Consume the existing `PreparePacketV1` and follow its work shape, risk, intent, Task State, bounded route explanation, selected skills, selected agents, verification, blockers, and next action.

A selected skill identifier maps to `.baron/core/skills/<name>/SKILL.md`. Load only the selected skill entrypoint and resolve its `references/`, `scripts/`, `assets/`, and nested files relative to `.baron/core/skills/<name>/`. A selected agent identifier maps to `.baron/core/agents/<name>` and remains subject to the parent session's lifecycle and proof gates.

Do not preload or recursively read all of `.baron/core/**`; this bridge contains no copied Baron skill body. Preserve explicit user intent and surface safety, integrity, identity, or migration blockers clearly. If Core or the Baron runtime is unavailable, stop and report the blocker instead of guessing.
