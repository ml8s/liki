# syntax=docker/dockerfile:1@sha256:ecfaec9ed6d810b56388c508f4121597bfbba70d41a6dfeee4d8cad5f295fc32
#
# 由 liki CI 构建：FROM 显式 pin 的 agent-runtime 纯运行时镜像（release base tag），
# COPY 单聚合组合工件（deployment.json + agents/）。
# 镜像不含任何源码、密钥或 MCP 凭据——端点/令牌仍由部署环境注入。
#
# 构建参数：
#   BASE_IMAGE       agent-runtime 纯运行时镜像（必填，显式 immutable release pin）

ARG BASE_IMAGE
FROM ${BASE_IMAGE}

ARG DEPLOYMENT_DIR=/deployment
ARG VERSION=local
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

# 组合工件由 liki CI 在构建前生成（make build-deployment）；运行时进程加载
# 单聚合 profile（一 app 一服务，ADK 生产实践）。
COPY dist/agents/ ${DEPLOYMENT_DIR}/

# Skill 文件（阶段2 skilltoolset source rootFS）：
#   /expert-packs/<pack>/   聚合专家包（根 skill 全量方法论卡）
COPY dist/skills/ /skills/
COPY dist/expert-packs/ /expert-packs/

LABEL org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.title="liki-agent-deployment" \
      org.opencontainers.image.description="Single aggregate AgentDeployment assembly image"
