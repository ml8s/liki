# Security policy

## Supported runtime

Security fixes are supported on the current `master` release and the newest
runtime CalVer on active release branches.

## Reporting a vulnerability

Please report vulnerabilities privately to [api@liki.hk](mailto:api@liki.hk).
Do not open a public GitHub issue for a suspected vulnerability.

Include:

- affected endpoint, binary, image digest, or file;
- runtime version from `/version` or `/healthz`;
- reproduction steps or request shape;
- impact assessment;
- whether the report may be shared with upstream dependencies.

We aim to acknowledge reports within two business days. Please allow up to 90
days for coordinated disclosure unless a shorter timeline is required by a
downstream consumer.

## Security boundaries

- Skill feedback must not contain birth data, names, dialogue text, addresses,
  API keys, or full tool payloads.
- Engine and counsel are stateless; do not add request logging that captures
  birth data.
- Self-hosted MCP deployments should set their inbound and outbound MCP tokens,
  HTTPS at the edge, and explicit `LIKI_ALLOWED_ORIGINS` when browser clients
  are used. counsel uses `LIKI_ENGINE_MCP_TOKEN` for engine calls and
  `LIKI_MCP_TOKEN` for its own inbound gate.
- Deployments behind a proxy must set the exact `LIKI_TRUSTED_PROXY_HOPS`.

## Supply chain

The repository uses:

- exact-commit GitHub Actions pins;
- digest-pinned base images and tools;
- hashed Python requirements;
- npm lockfile auditing;
- Go vulnerability scanning;
- image smoke tests before release publication.
