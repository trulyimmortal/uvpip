# uvpip reliability and evaluation roadmap

Updated: 2026-10-16 (Asia/Calcutta). Initial checkout: `4bcbdb2` (v1.0.2).
This is the continuation/handoff document for Codex or Antigravity.

## Objective and honest baseline

Improve the standalone CLI's real reliability and verifiable test evidence,
prioritizing Architecture and Test Coverage. The supplied evaluation reports
80 overall (B), not 88+. Category scores are judgments, not statement coverage:

| Category | Report score | Priority |
| --- | ---: | --- |
| Architecture & Robustness | 75 | First |
| Test Coverage | 82 | First |
| History & Maintenance | 55 | Organic, ongoing |
| Security Hygiene | 85 | Preserve; address concrete boundary flaws |
| Docs & Onboarding | 85 | Correct actual contradictions |
| Dependency Health | 85 | Verify existing update automation |
| Code Cleanliness | 88 | Preserve minimal standard-library design |
| CI/CD Maturity | 90 | Preserve native three-OS evidence |

Previous verified implementation commit `9cfa17d` had 97.2% Windows statement
coverage and green native CI run 38766401800. Reverify current source; do not
reuse that run as evidence for later changes. No guaranteed score-point gains.

The report calls three accounts three humans. `yv3000` and `trulyimmortal`
belong to the same maintainer; account count is not human diversity.

## Non-negotiable working rules

- Pure Go CLI, zero third-party runtime dependencies. No Docker, devcontainers,
  servers, metrics endpoints, daemons, or infrastructure.
- Read `AGENT.md`, `CONTRIBUTING.md`, and relevant source/tests before changing
  behavior. AGENT.md's baseline is stale; current code and user instructions win.
- Tests, builds, reports and temporary fixtures belong under
  `X:\ALL FOLDER\TESTING FOR EVERYTHING`.
- Go/tool caches belong under
  `X:\ALL FOLDER\MANAGER of all packages and etc`.
- Dot-source `scripts/env.ps1`, then redirect TEMP/TMP/TMPDIR/GOTMPDIR to the
  testing root. Resolve managed Go after this setup. Never use C: caches.
  Reading installed skill instructions or existing credentials is different
  from storing build/cache output there. Never print or persist tokens.
- Every production Go edit includes its sibling table-driven regression tests
  in the same commit. Keep each commit one coherent concern, verified first.
  Do not split code from tests merely to inflate commit count.
- Main fixes use yv3000. A small secondary-account PR may use trulyimmortal,
  with explicit same-maintainer and AI-assistance disclosure. Merge only after
  review and green CI. Do not manufacture independent reviews or contributors.
- Commit after each completed concern using Conventional Commits; push verified
  changes. No autonomous push-on-every-file agent: it could publish partial fixes.
- No new tags/releases unless explicitly requested. Preserve existing releases.

## Evidence from initial audit

What this repo does: one short-lived wrapper process forwards pip-compatible
arguments and streams to uv; per-user installers manage shims and profiles.
Expected load is ordinary interactive and scripted CLI use, including redirected
streams and activated environments, not a long-running backend.

1. **Running-child cancellation is not tested.**
   `runner_test.go:TestRunUVContextCancellation` cancels before starting uv.
   It cannot detect a regression that leaves a running child alive. Add a
   readiness handshake and cancel only after the fixture is actually running.
   Verify prompt return, non-success status and child cleanup. Keep pre-cancel
   coverage as a separate table case.
2. **Runner pipe draining has no bound.**
   `runUVContext` uses CommandContext with no WaitDelay; with non-file writers,
   descendants retaining output pipes can prevent Wait from returning.
   Doctor already sets WaitDelay, but its regression is POSIX-only. First prove
   the runner case with bounded fixtures, then add a minimal bound and test
   the error/exit classification. A wait bound does not kill an entire process
   tree; documentation must not promise that it does.
3. **Architecture descriptions contradict runtime behavior.**
   ARCHITECTURE.md says logging discards by default, but logging.go emits Info
   failures. The flow only mentions blank-command validation although size,
   NUL and verbosity checks exist. Correct these and explain why one package
   and zero runtime dependencies fit the CLI. Do not create duplicate ADR files.
4. **Fuzz checks are weak.**
   FuzzValidateUVOverride currently ignores the result; translation only checks
   nonempty output; environment fuzz only checks output length. Strengthen
   invariants for immutability, argument preservation and typed rejection.
   Do not duplicate implementation logic as a test oracle.

The evaluator missed existing ARCHITECTURE.md and .github/dependabot.yml.
Both exist; Dependabot covers gomod and github-actions weekly. CI already
uploads separate OS coverage artifacts and enforces 90% per OS. Do not add
duplicate gates/files or a schema library solely for scanner keywords.

## Ordered implementation batches

| ID | Work and acceptance evidence | Status |
| --- | --- | --- |
| R0 | Save baseline, prioritized roadmap, constraints and handoff | Done: this document |
| R1 | Running-child cancellation regression, readiness handshake, cleanup; normal and pre-cancel paths preserved | Implemented; Windows focused x3/full/vet/lint pass; native CI pending |
| R2 | Reproduce inherited-pipe runner wait; fix minimally with sibling regression; state process-tree limits accurately | Implemented; Windows repeated/full/vet/lint pass; native CI pending |
| R3 | Native Windows pip.cmd/pip3.cmd execution with spaced paths, literal args, stderr and child exit; compare existing installer checks to avoid duplication | Pending audit/test |
| R4 | Stronger fuzz invariants for validation/translation/environment; seeds plus bounded fuzz smoke run; no external dependencies | Pending |
| R5 | Correct existing architecture document; add rationale, CLI failure/cancellation contract, test mapping and direct README link | Pending; suitable small disclosed PR |
| R6 | Confirm Dependabot config is tracked and remote updates/CI visible; only fix a demonstrated configuration gap | Config verified locally/remotely; historical update PRs exist; no new configuration needed |
| R7 | Full verification and fresh committed clone; native CI on final SHA; update this ledger with evidence | Pending |
| R8 | Re-run external DataFactor evaluation on final SHA and compare categories; explicitly distinguish unmeasured gains | External reassessment needed |

For R3, a filename alone is not progress. Add Windows-specific tests only for
behavior not already proven by scripts/test-installers.ps1 and existing Go tests.
Do not turn on Windows -race without verifying a compatible C toolchain and
the incremental benefit; Linux/macOS already run the race detector.

Security follow-up during review: check whether translation flag scanning
mistakes an option value for a flag, and whether bundled verbosity flags are
handled consistently. Reproduce actual CLI behavior before expanding the parser.
Do not invent a complete pip parser or reject otherwise valid pass-through data.

Release workflow follow-up: current publishing does not depend on CI and accepts
manual dispatch. Document or propose a scoped gate separately after inspecting
the exact workflow; never publish a release as a validation side effect.

## Verification and completion evidence

Run from repository root in a fresh PowerShell process:

```powershell
$tests = 'X:\ALL FOLDER\TESTING FOR EVERYTHING'
$cache = 'X:\ALL FOLDER\MANAGER of all packages and etc'
. .\scripts\env.ps1 -CacheRoot $cache
$env:TEMP = $env:TMP = $env:TMPDIR = $env:GOTMPDIR = $tests
$goBin = Join-Path $env:LOCALAPPDATA 'go1.26.8\go\bin'
$env:PATH = "$goBin;$env:PATH"
powershell.exe -NoLogo -NoProfile -NonInteractive -ExecutionPolicy Bypass -File scripts/verify.ps1 -TestRoot $tests -CacheRoot $cache
if ($LASTEXITCODE -ne 0) { throw 'Verification failed' }
```

Require gofmt, go vet, full tests, >=90% coverage per OS (preserve ~97% baseline),
lint, module tidy/checksums, govulncheck, gitleaks, actionlint when workflows
change, shell checks and offline installers. Inspect actual test assertions,
not merely success badges. Mutation/revert proof for a bug fix must happen in
an isolated X: scratch checkout, never leave deliberately broken source committed.

Use CONTRIBUTING.md's fresh clone process after committing. Record SHA, artifact
path, exact command/result, and native CI URL below. Cross-builds are not native
execution evidence. Test sandbox fixtures must not mutate real profiles or PATH.

## Handoff ledger

- 2026-10-16: read all five supplied images; recorded score 80 and category table.
- Checked clean checkout at 4bcbdb2; v1.0.2 release change predates this work.
- Read runner/doctor/fuzz paths, architecture and dependency automation.
- R1: runner_lifecycle_test.go uses child stdout readiness to cancel only after
  startup; bounded sleeping fixture exits if cancellation regresses. CommandContext
  Run waits/reaps the direct child. Pre-start and post-readiness cases are checked.
  This does not prove process-tree termination or inherited-pipe handling (R2).
- Windows focused lifecycle tests passed three repetitions; full tests passed
  at 97.2% statement coverage; go vet and golangci-lint passed (0 issues).
- No production/runtime code changed in R1. runner_test.go remains 397 lines.
- R1 native CI run 38897701084: all four jobs passed.
- R2 red proof: parent returned success only after 9.41 seconds waiting for
  descendant pipes; expected exit 1 failed before production edits. New WaitDelay
  bounds drain after exit/cancellation; ErrWaitDelay prints a specific diagnostic
  and returns 1, not misleading startup failure 126. Repeated focused tests,
  full Windows tests (97.3%), vet and lint passed. Descendant fixture is bounded
  and explicitly killed during test cleanup. Runtime makes no tree-kill promise.
- R3 audit: existing Windows suite covers spaced/quoted home, both cmd shims,
  argument forwarding, exit code 23 and profile lifecycle. Add missing stdin/
  stderr assertions to that suite rather than duplicate it in windows_test.go.
- R6 remote .github/dependabot.yml exists. Historical update PRs #1/#2 were
  closed; no claim they merged or prove current schedule execution. Old Docker
  PRs #3/#4 are historical only; current config has no Docker ecosystem.
- Next: commit/push R2, then R3 stream assertions and R4 fuzz invariants.
- Completion is NOT claimed: R1-R8 need evidence or an explicit documented
  disposition. Update this ledger after every completed concern.

## History and score interpretation

Healthy history comes from independently understandable fixes with regression
tests and real review over time. One-line artificial commits, two accounts of
one person, and fabricated review independence do not reduce maintainer risk.
Keep a few honest good-first-issue tasks for real contributors, credit actual
contributions, and avoid promised response SLAs the maintainer cannot meet.

The evaluation's dependency and architecture omissions may be evidence-bundle
limitations. Supply the existing tracked files and CI artifacts on reassessment
instead of bloating the CLI. The 88+ goal remains unverified until that report.
