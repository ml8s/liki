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
