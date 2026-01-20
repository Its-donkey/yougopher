# Git Hooks

This repository uses a local hooks directory wired via `core.hooksPath` to share useful checks.

## Pre‑push: branch scope guard

Goal: keep each feature branch and PR tightly scoped by flagging unrelated files before they are pushed.

Behavior:
- Only runs on branches named with the convention: `<type>/<section>[/<subsection>]/<feature>` where `type` is one of
  `feature|bugfix|hotfix|design|refactor|test|doc`.
- Extracts the section (and optional subsection) from the branch name.
- Examines the diff being pushed and warns or blocks when files do not include the section token in their path.
- Safe to bypass with `git push --no-verify` or by setting env `SKIP_SCOPE_GUARD=1`.

Notes:
- The guard is heuristic; some changes naturally cross boundaries (e.g., shared utils). Use `--no-verify` when appropriate.
- You can extend tokens per-branch by creating an optional JSON file `.scope-allow.json` (root) like:

```
{
  "feature/items/icon-stroke-menu": ["icons", "components"],
  "feature/items/items-dnd": ["utils"]
}
```

Those tokens are matched as path segments (e.g., `components/Header.tsx`).

## Install locally

The repo config sets `core.hooksPath` to `.githooks`. If you cloned before this change, run:

```
git config core.hooksPath .githooks
```

## Bypass

```
SKIP_SCOPE_GUARD=1 git push
# or
git push --no-verify
```

