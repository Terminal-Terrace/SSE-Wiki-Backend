#!/bin/bash

# Auth Service API 测试脚本
# 使用方法: ./test_api.sh

BASE_URL="http://localhost:8081/api/v1/auth"
COOKIE_FILE="cookies.txt"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "Auth Service API 测试"
echo "========================================="
echo ""

# 1. 测试预登录接口
echo -e "${YELLOW}[1/5] 测试预登录接口...${NC}"
PRELOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/prelogin" \
  -H "Content-Type: application/json" \
  -d '{"redirect_url":"http://localhost:3000/callback"}' \
  --noproxy "*")

STATE=$(echo $PRELOGIN_RESPONSE | grep -o '"state":"[^"]*"' | cut -d'"' -f4)

if [ -n "$STATE" ]; then
  echo -e "${GREEN}✓ 预登录成功${NC}"
  echo "  State: $STATE"
else
  echo -e "${RED}✗ 预登录失败${NC}"
  echo "  响应: $PRELOGIN_RESPONSE"
  exit 1
fi
echo ""

# 2. 测试发送验证码接口
echo -e "${YELLOW}[2/5] 测试发送验证码接口...${NC}"
read -p "请输入邮箱地址: " EMAIL
CODE_RESPONSE=$(curl -s -X POST "$BASE_URL/code" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"type\":1}" \
  --noproxy "*")

if echo "$CODE_RESPONSE" | grep -q '"code":100'; then
  echo -e "${GREEN}✓ 验证码发送成功${NC}"
  echo "  请查收邮件"
else
  echo -e "${RED}✗ 验证码发送失败${NC}"
  echo "  响应: $CODE_RESPONSE"
  exit 1
fi
echo ""

# 3. 测试注册接口
echo -e "${YELLOW}[3/5] 测试注册接口...${NC}"
read -p "请输入用户名: " USERNAME
read -sp "请输入密码: " PASSWORD
echo ""
read -p "请输入收到的验证码: " CODE

REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/register" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\",\"confirm_password\":\"$PASSWORD\",\"email\":\"$EMAIL\",\"code\":\"$CODE\",\"state\":\"$STATE\"}" \
  -c "$COOKIE_FILE" \
  --noproxy "*")

if echo "$REGISTER_RESPONSE" | grep -q '"code":100'; then
  echo -e "${GREEN}✓ 注册成功${NC}"
  echo "  响应: $REGISTER_RESPONSE"
else
  echo -e "${RED}✗ 注册失败${NC}"
  echo "  响应: $REGISTER_RESPONSE"
  exit 1
fi
echo ""

# 4. 测试登录接口
echo -e "${YELLOW}[4/5] 测试登录接口...${NC}"
# 获取新的 state
PRELOGIN_RESPONSE2=$(curl -s -X POST "$BASE_URL/prelogin" \
  -H "Content-Type: application/json" \
  -d '{"redirect_url":"http://localhost:3000/home"}' \
  --noproxy "*")
STATE2=$(echo $PRELOGIN_RESPONSE2 | grep -o '"state":"[^"]*"' | cut -d'"' -f4)

LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/login" \
  -H "Content-Type: application/json" \
  -d "{\"type\":\"sse-wiki\",\"state\":\"$STATE2\",\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}" \
  -c "$COOKIE_FILE" \
  --noproxy "*")

if echo "$LOGIN_RESPONSE" | grep -q '"code":100'; then
  echo -e "${GREEN}✓ 登录成功${NC}"
  REFRESH_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)
  echo "  Refresh Token: ${REFRESH_TOKEN:0:20}..."
else
  echo -e "${RED}✗ 登录失败${NC}"
  echo "  响应: $LOGIN_RESPONSE"
  exit 1
fi
echo ""

# 5. 测试刷新令牌接口
echo -e "${YELLOW}[5/5] 测试刷新令牌接口...${NC}"
REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/refresh" \
  -b "$COOKIE_FILE" \
  --noproxy "*")

if echo "$REFRESH_RESPONSE" | grep -q '"access_token"'; then
  echo -e "${GREEN}✓ 刷新令牌成功${NC}"
  ACCESS_TOKEN=$(echo $REFRESH_RESPONSE | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
  echo "  Access Token: ${ACCESS_TOKEN:0:50}..."
else
  echo -e "${RED}✗ 刷新令牌失败${NC}"
  echo "  响应: $REFRESH_RESPONSE"
  exit 1
fi
echo ""

# 清理
rm -f "$COOKIE_FILE"

echo "========================================="
echo -e "${GREEN}所有测试通过！${NC}"
echo "========================================="
