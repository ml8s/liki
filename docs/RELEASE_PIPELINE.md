# Release pipeline

Liki uses a small, standards-aligned implementation of the common
**build once, test once, sign once, deploy the same digest** model, driven by
CalVer tags.

The ordered operator runbook lives in
[`liki-deploy/docs/RELEASE_RUNBOOK.md`](https://github.com/ml8s/liki-deploy/blob/main/docs/RELEASE_RUNBOOK.md).

## Boundaries

| Stage | Owns | Never does |
|---|---|---|
| Commit / PR CI | source lint, checks, tests, audits | build or publish images |
| Release workflow (`release.yml` on a `vYYYY.MM.DD.N` tag) | gate, build, smoke, sign, GitHub Release | deploy anything |
| `liki-deploy` | verify a signed digest lock, run the assembly harness, deploy | build product images or orchestrate product releases |

`make image` and `make smoke` remain available for development; production
images only come from `release.yml`.

## Commit / PR gate

```text
make check
make test
```

This validates source, configuration, schemas, and tests. It must not produce a
GHCR image.

Testing is shifted as far left as possible:

1. run the fastest local static checks before committing;
2. run local `make check` and `make test` before push;
3. run CI source gates on push/PR;
4. build release images only for a release tag;
5. run image and assembly tests against the released digests.

Never defer a check to the registry or deployment stage when the same
deterministic check can run earlier from source.

## Test layers

Product repositories own their images, while `liki-deploy` owns the assembled
system. Do not duplicate every full-system journey in every product repository.

| Layer | Owner | Runs before | Verifies |
| --- | --- | --- | --- |
| Source gate | product repo | commit / PR | lint, unit/integration, configuration, schema, supply-chain checks |
| Image gate | product repo `release.yml` | GHCR push | CI build, startup, health/readiness, product entry point, deployment artifact validation |
| Signature gate | product repo `release.yml` | GitHub Release | cosign sign + verify with the workflow identity |
| Assembly gate | `liki-deploy` | lock recording / deploy | exact release digests, Compose topology, network boundaries, cross-service journeys |
| Post-deploy gate | `liki-deploy` | regional deploy | real edge/public URL, readiness, public and authenticated paths |

## Release flow

A release starts from a green `main` commit whose version surfaces
(`make version`), changelog, and dependency pins already describe the release:

```text
green main commit
  → push the tag:            git push origin vYYYY.MM.DD.N
  → release.yml guard:       CalVer shape, tag on main, images/release immutable
  → release gate:            make release-gate (check + test + runtime validate + artifacts)
  → buildx build + push      ghcr.io/ml8s/<image>:vYYYY.MM.DD.N (+ latest)
  → image smoke              against the pushed digest
  → cosign sign + verify     workflow identity via OIDC
  → GitHub Release           candidate-digests.json + signed artifacts
  → liki-deploy              fetch release manifests → compose lock → assembly e2e → deploy
```

Release tags and their image tags are immutable. If a release fails, fix the
source and cut the next `vYYYY.MM.DD.N`; never rebuild into an existing tag.

### Human interaction

| You say | Pipeline does | Image side effect |
| --- | --- | --- |
| "commit and push" | run source checks/tests only | none |
| "push a release tag" | `release.yml`: gate → build → smoke → sign → release | creates release-tagged GHCR images |
| "update lock / deploy" | verify signatures and deploy the digest-pinned matrix | none |

### Release manifest

Every release publishes a `candidate-digests.json` asset containing at least:

```json
{
  "source_commit": "...",
  "images": {
    "engine": "sha256:...",
    "counsel": "sha256:...",
    "experts": "sha256:..."
  }
}
```

This asset is the cross-repository contract: `liki-deploy` resolves the exact
digest matrix from the latest product releases through it. Tags are
human-readable aliases; the digest manifest is the release identity.

## Operational workflows

| Repository | Release path |
| --- | --- |
| `liki` | tag → `release.yml` (engine / counsel / experts) |
| `liki-web` | tag → `release.yml` (web) |
| `liki-agent-runtime` | tag → `release.yml` (runtime base) |
| `liki-deploy` | `deploy.yml` / `deploy-staging.yml` / `assembly.yml` (no tags) |

Typical order:

```bash
# 1. Create and push the release-preparation commit: selected CalVer, version
#    surfaces, changelog, and any released dependency pins.

# 2. Push release tags only for changed repositories (runtime first when the
#    base changed — liki's pin must be updated afterwards).
git -C liki-agent-runtime push origin vYYYY.MM.DD.N   # if runtime changed
#    → update liki deps/liki-agent-runtime.base, gate, include in liki's tag
git -C liki-web push origin vYYYY.MM.DD.N             # if web changed
git -C liki push origin vYYYY.MM.DD.N                 # if liki changed

# 3. liki-deploy: the regional deploy resolves the latest release manifests
#    itself; for a pre-approved staging lock, compose and commit it first:
cd liki-deploy
CANDIDATE_MANIFESTS=<liki.json>:<liki-web.json> make update-release-lock \
  DEPLOY_LOCK=env/staging.lock
gh workflow run deploy-staging.yml -f lock_ref=main -f apply=true
```

### Dependency changes

```text
liki-agent-runtime changed
  → tag its release.yml; update deps/liki-agent-runtime.base in liki
    (version + digest) and pass make runtime-validate
  → tag liki; experts rebuild FROM the new pinned base

liki skill/domain content changed
  → tag liki
  → republish liki-web only if it consumes a changed skill bundle

liki-web changed alone
  → tag liki-web; the next deploy resolves only its new digest
```

Never point a production assembly build at an unreleased or dirty
`liki-agent-runtime`.

## Immutable identity

Tags are human-readable aliases. The deployment identity is always the OCI
digest:

```text
ghcr.io/ml8s/liki-engine@sha256:<digest>
```

The digest tested, signed, released, and deployed must be identical.

## Deployment acceptance

`liki-deploy` consumes only digest-pinned release images:

1. validate lock fields and digest policy;
2. ensure rendered Compose uses the locked digest;
3. verify registry manifest and cosign signature;
4. deploy;
5. verify live containers use the locked digests and run post-deploy smoke.

Assembly runs before the lock is recorded. A failed gate must not update the
lock and must not be deployed.
