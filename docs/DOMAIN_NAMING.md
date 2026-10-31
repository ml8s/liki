# Domain naming matrix

This is the canonical vocabulary for repository, image, service, and artifact names.

## Runtime and deployment names

| Canonical term | Name | Meaning |
| --- | --- | --- |
| Engine compute | `liki-engine` | Deterministic computation image. |
| Engine MCP service | `engine-mcp` | Streamable MCP compute service. |
| Counsel judgment | `liki-counsel` | Judgment image. |
| Counsel MCP service | `counsel-mcp` | Streamable MCP judgment service. |
| Agent runtime | `liki-agent-runtime` | Generic ADK runtime base image. |
| Expert assembly | `liki-expert` | Production single-expert aggregate image. |
| Expert service | `expert` | Compose service running `liki-expert`. |
| Expert pack | `liki` | The one WorkBuddy expert pack; domains are internal routing cards. |
| Web BFF | `liki-web` / `web` | Commerce, identity, chat BFF, and public site. |

`liki-agent-runtime` is the runtime base. `liki-expert` is the production aggregate image.

## Artifact names

| Artifact | Name |
| --- | --- |
| Runtime base | local `liki-agent-runtime` image tagged by git SHA |
| AgentDeployment | `deployment.json` |

## Contract versions

All JSON contracts use:

```text
liki.<domain>.<artifact>.v<major>
```

Examples:

```text
liki.agent-deployment.v1
liki.feedback.v1
liki.aipay.v1
```

## Naming domain

The service/domain is `naming`; `qiming` is not the tool prefix. MCP tools use `naming_*`.
