# Release pipeline

Liki uses a small, standards-aligned implementation of the common
**build once, test once, sign once, deploy the same digest** model.

## Boundaries

| Stage | Owns | Never does |
|---|---|---|
| Commit / PR CI | source lint, checks, tests, audits | build or publish images |
| Release workflow | build, publish, smoke, sign images | build on ordinary commits |
| `liki-deploy` | verify signed digest lock, assembly e2e, deploy | build product images |

Local `make image` / `make smoke` targets remain available for manual release
preparation, but they are not part of ordinary commit CI.

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
4. build a candidate only for release;
5. run image and assembly tests against the candidate digest.

Never defer a check to the registry or deployment stage when the same
deterministic check can run earlier from source.

## Candidate / promotion gate

A formal release version is allocated only after a candidate has passed all
gates. A candidate is an immutable set of image digests built from one source
commit. It is not a new source branch or a second build of the same commit.

```text
source commit
  → build candidate images with candidate tags
  → publish candidate digest manifest
  → smoke / assembly / e2e the exact candidate digests
  → all gates pass
  → promote the same digests to release tags
  → sign the same release digests
  → create tag / GitHub Release / deployment lock change
```

Promotion never rebuilds images. It only gives the already tested digest set a
release identity.

### Human interaction

| You say | Pipeline does | Image side effect |
| --- | --- | --- |
| “commit and push” | run source checks/tests only | none |
| “build candidate” | build/push candidate images and record digests | creates candidate-tagged images |
| “test candidate” | run smoke/assembly/e2e against the candidate digests | none |
| “promote candidate” | allocate version, retag/promote exact digests, sign, create release | creates release-tag aliases to the same digests |
| “update lock / deploy” | verify signatures and run digest-pinned assembly | none |

If a candidate fails, fix the source and build a new candidate. Do not reuse the
formal version number and do not rebuild the failed source into the same release
tag.

### Candidate metadata

Every candidate produces a digest manifest containing at least:

```json
{
  "source_commit": "...",
  "images": {
    "engine": "sha256:...",
    "counsel": "sha256:...",
    "experts": "sha256:...",
    "single_expert": "sha256:...",
    "web": "sha256:..."
  }
}
```

Candidate tags are temporary aliases and may be cleaned by retention policy.
The digest manifest is the candidate identity.

## Operational workflows

| Repository | Build candidate | Promote release |
| --- | --- | --- |
| `liki-agents` | `release-candidate.yml` | `promote-release.yml` |
| `liki` | `release-candidate.yml` | `promote-release.yml` |
| `liki-web` | `release-candidate.yml` | `promote-release.yml` |
| `liki-deploy` | `candidate-e2e.yml` | lock update + release gate |

Typical dispatch order:

```bash
# 1. Build all candidates from green source refs.
gh workflow run release-candidate.yml --repo ml8s/liki --ref main -f ref=main
gh workflow run release-candidate.yml --repo ml8s/liki-agents --ref main -f ref=main
gh workflow run release-candidate.yml --repo ml8s/liki-web --ref main -f ref=main

# 2. Run assembly e2e with the candidate digest manifest values.
gh workflow run candidate-e2e.yml --repo ml8s/liki-deploy \
  -f engine_image='ghcr.io/ml8s/liki-engine@sha256:...' \
  -f counsel_image='ghcr.io/ml8s/liki-counsel@sha256:...' \
  -f experts_image='ghcr.io/ml8s/liki-multi-expert@sha256:...' \
  -f web_image='ghcr.io/ml8s/liki-web@sha256:...'

# 3. After e2e passes, promote each repository from its selected tag.
gh workflow run promote-release.yml --repo ml8s/liki-agents \
  --ref v2026.10.14.0 -f version=2026.10.14.0 -f candidate_run_id=<id>
gh workflow run promote-release.yml --repo ml8s/liki \
  --ref v2026.10.14.0 -f version=2026.10.14.0 -f candidate_run_id=<id>
gh workflow run promote-release.yml --repo ml8s/liki-web \
  --ref v2026.10.14.0 -f version=2026.10.14.0 -f candidate_run_id=<id>
```

The `liki-deploy` lock change is merged only after all promotions and signature
verifications succeed.

## Release gate

A release workflow is the only place that builds publishable images.

For each product image:

```text
build once
  → push immutable release tags
  → run smoke against the published artifact
  → record digest
  → cosign sign the digest
```

Assembly images are then built from the same release inputs and pinned runtime
base. They are validated with the published deployment before signing.

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
4. run digest-pinned assembly e2e;
5. deploy and run post-deploy smoke.

A failed release gate must not update the lock and must not be deployed.
