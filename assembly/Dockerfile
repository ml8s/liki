# syntax=docker/dockerfile:1@sha256:ecfaec9ed6d810b56388c508f4121597bfbba70d41a6dfeee4d8cad5f295fc32
#
# 由 liki CI 构建：FROM 显式 pin 的 liki-agent-runtime 纯运行时镜像（release base tag），
# COPY 组合工件（deployment.json + agents/），并通过 ENV 指向工件。
# 镜像不含任何源码、密钥或 MCP 凭据——端点/令牌仍由部署环境注入。
#
# 构建参数：
#   BASE_IMAGE       liki-agent-runtime 纯运行时镜像（必填，显式 immutable release pin）
#   PROFILE          工件 profile 名（experts / single）

ARG BASE_IMAGE
FROM ${BASE_IMAGE}

ARG DEPLOYMENT_DIR=/deployment
ARG PROFILE=experts
ARG VERSION=local
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

# 组合工件由 liki CI 在构建前生成（make build-deployment），随镜像烘焙。
COPY dist/agents/${PROFILE}/ ${DEPLOYMENT_DIR}/

# Skill 文件（阶段2 skilltoolset source rootFS）：
#   /expert-packs/<pack>/   4 专家包（= 四个顶层领域；各专家绑 pack/skills 为 rootFS）
COPY dist/skills/ /skills/
COPY dist/expert-packs/ /expert-packs/

ENV LIKI_AGENT_RUNTIME_DEPLOYMENT_FILE=${DEPLOYMENT_DIR}/deployment.json
LABEL org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.title="liki-agent-deployment-${PROFILE}" \
      org.opencontainers.image.description="Liki AgentDeployment assembly image (profile: ${PROFILE})"
