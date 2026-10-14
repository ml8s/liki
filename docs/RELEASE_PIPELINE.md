# Release pipeline

Liki uses a small, standards-aligned implementation of the common
**build once, test once, sign once, deploy the same digest** model.

The ordered operator runbook lives in
[`liki-deploy/docs/RELEASE_RUNBOOK.md`](https://github.com/ml8s/liki-deploy/blob/main/docs/RELEASE_RUNBOOK.md).

## Boundaries

| Stage | Owns | Never does |
|---|---|---|
| Commit / PR CI | source lint, checks, tests, audits | build or publish images |
| Local release candidate | build, smoke, and record image digests in the local registry | publish to GHCR before system test |
| Promotion workflow | retag the same digest, sign, and verify its release identity | build, smoke, or create a GitHub Release |
| `liki-deploy` | verify a signed digest lock, run the assembly harness, deploy | build product images or orchestrate product releases |

`make candidate` is the release preparation target. It is explicit and is never
part of ordinary commit CI. `make image` and `make smoke` remain available for
development.

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

## Candidate test layers

Candidate quality is layered. Product repositories own their image, while
`liki-deploy` owns the assembled system. Do not duplicate every full-system
journey in every product repository.

| Layer | Owner | Runs before | Verifies |
| --- | --- | --- | --- |
| Source gate | product repo | commit / PR / candidate | lint, unit/integration, configuration, schema, supply-chain checks |
| Image gate | product repo | `docker push` | local build, startup, health/readiness, product entry point, deployment artifact validation |
| Assembly gate | `liki-deploy` | candidate publication | exact candidate digests, Compose topology, network boundaries, cross-service journeys |
| Promotion gate | product repo | release tag | same digest, source commit, signature, and digest verification |
| Post-deploy gate | `liki-deploy` | regional deploy | real edge/public URL, readiness, public and authenticated paths |

`make candidate` may publish only after the product image gate passes. The
system-level `candidate-test` runs after candidate manifests exist.

## Candidate / promotion gate

Select the release version and update all version surfaces before creating the
release-preparation commit. A release tag is created only after a candidate has
passed all gates. A candidate is an immutable set of image digests built from
one clean source commit. It is not a new source branch or a second build of the
same commit.

```text
source commit
  → local source gate
  → local buildx build candidate images
  → local image smoke
  → record the exact digest manifest
  → assembly / e2e from the local registry; write assembly evidence
  → all system gates pass
  → publish the same image digests to GHCR
  → create the release tag on the same source commit
  → promotion retags, signs, and verifies the same digest
  → update the deployment lock only from the signed digest matrix
```

Promotion never rebuilds images. It only gives the already tested digest set a
release identity.

### Human interaction

| You say | Pipeline does | Image side effect |
| --- | --- | --- |
| “commit and push” | run source checks/tests only | none |
| “build candidate” | build, smoke, and record images in the local registry | creates local candidate-tagged images, no GHCR push |
| “test candidate” | run smoke/assembly/e2e against the candidate digests | none |
| “publish candidate” | copy system-tested local digests to GHCR | creates GHCR candidate aliases to the same digests |
| “promote candidate” | tag, retag/promote exact digests, sign, verify signature | creates release-tag aliases to the same digests |
| “update lock / deploy” | verify signatures and deploy the digest-pinned matrix | none |

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
    "single_expert": "sha256:..."
  }
}
```

Candidate tags are temporary aliases and may be cleaned by retention policy.
The digest manifest is the candidate identity.

Assembly writes an `assembly-evidence.json` next to each manifest. The evidence
records the manifest SHA-256, source commit, candidate tag, and pass status. A
product repository must verify that evidence before `candidate-publish`.

## Operational workflows

| Repository | Build candidate | Promote release |
| --- | --- | --- |
| `liki` | local `make candidate` | `promote-release.yml` |
| `liki-agents` | local `make candidate` | `promote-release.yml` |
| `liki-web` | local `make candidate` | `promote-release.yml` |
| `liki-deploy` | local `make candidate-test` | lock update + release gate |

Typical dispatch order:

```bash
# 1. Create and push the release-preparation commit: selected CalVer, version
#    surfaces, changelog, and any released dependency pins.

# 2. Build candidates only for changed repositories and their dependency
#    updates. Use RELEASE_VERSION matching the release-preparation surfaces.
cd liki-web && RELEASE_VERSION=... make candidate          # web-only change
cd liki && RELEASE_VERSION=... make candidate              # liki-only change
cd liki-agents && RELEASE_VERSION=... make candidate       # agents-only change

# 3. Run local-registry assembly e2e with manifests for changed repositories;
#    unchanged application images fall back to the approved base lock.
cd liki-deploy && CANDIDATE_MANIFESTS=... make candidate-test

# 4. After e2e passes, copy only the tested local digests to GHCR.
cd liki-web && make candidate-publish       # if web changed
cd liki && make candidate-publish           # if liki changed
cd liki-agents && make candidate-publish    # if agents changed

# 5. Tag and promote only the tested candidates. The
#    promotion command follows the same repository/manifest pair as above.
git -C liki-web tag v2026.10.14.0 <full-sha>       # if web changed
git -C liki tag v2026.10.14.0 <full-sha>           # if liki changed
git -C liki-agents tag v2026.10.14.0 <full-sha>    # if agents changed
gh workflow run promote-release.yml --repo ml8s/liki \
  --ref v2026.10.14.0 \
  -f version=2026.10.14.0 -f source_commit=<full-sha> \
  -f digests_json="$(cat ../liki/build/candidate/candidate-digests.json)"
gh workflow run promote-release.yml --repo ml8s/liki-agents \
  --ref v2026.10.14.0 \
  -f version=2026.10.14.0 -f source_commit=<full-sha> \
  -f digests_json="$(cat ../liki-agents/build/candidate/candidate-digests.json)"
gh workflow run promote-release.yml --repo ml8s/liki-web \
  --ref v2026.10.14.0 \
  -f version=2026.10.14.0 -f source_commit=<full-sha> \
  -f digests_json="$(cat ../liki-web/build/candidate/candidate-digests.json)"
```

The `liki-deploy` lock change is merged only after all promotions and signature
verifications succeed.

### Dependency changes

```text
liki-agents changed
  → run its candidate smoke, assembly-test, publish, promote, and sign it
  → update deps/liki-agents.base in liki to the new released digest
  → build, test, and promote the liki expert images

liki skill/domain content changed
  → candidate-test/promote liki
  → republish liki-web only if it consumes a changed skill bundle

liki-web changed alone
  → candidate-test web against the approved base lock
  → promote web and update only its deployment digest
```

Never point a production assembly build at an unreleased or dirty
`liki-agents` candidate.

## Release gate

For `liki`, local `make candidate` is the only image build path; Actions
promotion never builds images.

For each product image:

```text
local candidate build
  → local smoke and validation
  → record candidate digest
  → local-registry assembly/e2e on that digest
  → publish the same digest to GHCR
  → promotion retags the same digest as a release version
  → promotion cosign signs and verifies that digest
```

Assembly images are built locally from the pinned runtime base before promotion.
They are validated with the deployment manifest and local assembly E2E before
signing.

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

Assembly runs before publication. A failed release gate must not update the
lock and must not be deployed.
