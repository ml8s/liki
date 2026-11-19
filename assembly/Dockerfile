# syntax=docker/dockerfile:1@sha256:ecfaec9ed6d810b56388c508f4121597bfbba70d41a6dfeee4d8cad5f295fc32
#
# 由 Liki 构建流程组装：FROM 显式 pin 的 AgentGo 纯运行时镜像，
# COPY 专家 deployment 工件（deployment.json + agents/）。
# 镜像不含任何源码、密钥或 MCP 凭据——端点/令牌仍由部署环境注入。
#
# 构建参数：
#   BASE_IMAGE       agentgo 纯运行时镜像（必填，显式 immutable release pin）

ARG BASE_IMAGE
FROM ${BASE_IMAGE}

ARG DEPLOYMENT_DIR=/deployment
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

# 专家工件由构建流程在前面生成（make build-expert-deployment）；运行时进程加载
# 一个 Liki expert deployment（一专家一服务）。
COPY dist/deployment/ ${DEPLOYMENT_DIR}/

# Skill 文件（阶段2 skilltoolset source rootFS）：
COPY dist/skills/ /skills/

LABEL org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.title="liki-expert" \
      org.opencontainers.image.description="Liki expert runtime assembled on AgentGo"
