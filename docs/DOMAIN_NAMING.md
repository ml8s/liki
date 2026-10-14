# Domain naming matrix

This is the canonical vocabulary for repository, image, service, and artifact
names.

## Runtime and deployment names

| Canonical term | Name | Meaning |
| --- | --- | --- |
| Engine compute | `liki-engine` | Deterministic computation image. |
| Engine MCP service | `engine-mcp` | Streamable MCP compute service. |
| Engine compatibility runtime | `engine-rpc` | Retained JSON-RPC compatibility server. |
| Counsel judgment | `liki-counsel` | Judgment image. |
| Counsel MCP service | `counsel-mcp` | Streamable MCP judgment service. |
| Agent runtime | `liki-agent-runtime` | Generic ADK runtime base image. |
| Expert assembly | `liki-experts` | Production expert assembly image. |
| Expert service | `experts` | Compose service running `liki-experts`. |
| Web BFF | `liki-web` / `web` | Commerce, identity, chat BFF, and public site. |

Do not call `liki-agent-runtime` the production expert image. Do not call
`liki-experts` the generic runtime.

## Artifact names

| Artifact | Name |
| --- | --- |
| Runtime base pin | `deps/liki-agent-runtime.base` |
| AgentDeployment | `deployment.json` |
| Candidate manifest | `build/candidate/candidate-digests.json` |
| Assembly evidence | `assembly-evidence.json` |
| Release manifest | `liki-release-manifest.json` |
| Deployment lock | `env/<region>.lock` |

## Contract versions

All JSON contracts use:

```text
liki.<domain>.<artifact>.v<major>
```

Examples:

```text
liki.agent-deployment.v1
liki.candidate-manifest.v1
liki.assembly-evidence.v1
liki.release-manifest.v1
liki.feedback.v1
liki.aipay.v1
```

## Naming domain

The service/domain is `naming`; `qiming` is not the tool prefix. MCP tools use
`naming_*`.
