#!/bin/bash
# FilePath    : .gitalias/save.sh
# Author      : jiaopengzi
# Blog        : https://jiaopengzi.com
# Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
# Description : 保存当前修改并推送到远程仓库(支持双仓库推送和强制推送)

# 设置 Git 别名命令:
# 在 .git/config 文件中添加以下内容：
# [alias]
#     save = "!bash ./.gitalias/save.sh"
# 运行命令: git save "提交信息"
#          git save -a "提交信息"  (推送到两个远程仓库)
#          git save -f "提交信息"  (强制推送)
#          git save --no-push "提交信息"  (仅提交，不推送)

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# ========== 函数定义 ==========

# 显示帮助信息
show_help() {
    echo ""
    echo -e "${CYAN}=== save.sh 使用说明 ===${NC}"
    echo ""
    echo -e "${GREEN}用法:${NC}"
    echo "  git save \"提交信息\"              保存所有修改并推送到当前分支的默认远程"
    echo "  git save -a \"提交信息\"           推送到所有远程仓库 (origin + gitlab)"
    echo "  git save -f \"提交信息\"           强制推送 (--force-with-lease)"
    echo "  git save --no-push \"提交信息\"    仅提交，不推送"
    echo "  git save --amend                  追加到上一次提交"
    echo "  git save --help                   显示此帮助"
    echo ""
    echo -e "${GREEN}示例:${NC}"
    echo "  git save \"feat: 添加登录功能\""
    echo "  git save -a \"fix: 修复样式问题\""
    echo "  git save --no-push \"WIP: 开发中\""
    echo "  git save --amend"
    echo ""
}

# 检查是否有远程仓库
check_remote() {
    local remote=$1
    if git remote | grep -q "^$remote$"; then
        return 0
    else
        return 1
    fi
}

# 检查是否有未跟踪文件
has_untracked_files() {
    if [ -n "$(git ls-files --others --exclude-standard)" ]; then
        return 0
    else
        return 1
    fi
}

# 检查是否有未提交的更改
check_changes() {
    if git diff --quiet && git diff --cached --quiet && ! has_untracked_files; then
        return 1
    else
        return 0
    fi
}

# 显示即将提交的更改摘要
show_changes_summary() {
    echo -e "${CYAN}即将提交的更改:${NC}"
    echo ""
    
    # 已暂存的更改
    if ! git diff --cached --quiet; then
        echo -e "${GREEN}已暂存:${NC}"
        git diff --cached --stat
        echo ""
    fi
    
    # 未暂存的更改
    if ! git diff --quiet; then
        echo -e "${YELLOW}未暂存 (将被自动添加):${NC}"
        git diff --stat
        echo ""
    fi
    
    # 未跟踪的文件
    if [ -n "$(git ls-files --others --exclude-standard)" ]; then
        echo -e "${BLUE}未跟踪的文件 (将被自动添加):${NC}"
        git ls-files --others --exclude-standard | head -10
        local count=$(git ls-files --others --exclude-standard | wc -l)
        if [ "$count" -gt 10 ]; then
            echo "... 及其他 $((count - 10)) 个文件"
        fi
        echo ""
    fi
}

# ========== 参数解析 ==========

PUSH_ALL=false
FORCE_PUSH=false
NO_PUSH=false
AMEND=false
COMMIT_MSG=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --help|-h|help)
            show_help
            exit 0
            ;;
        -a|--all)
            PUSH_ALL=true
            shift
            ;;
        -f|--force)
            FORCE_PUSH=true
            shift
            ;;
        --no-push)
            NO_PUSH=true
            shift
            ;;
        --amend)
            AMEND=true
            shift
            ;;
        -*)
            echo -e "${RED}❌ 错误: 未知选项 $1${NC}"
            show_help
            exit 1
            ;;
        *)
            COMMIT_MSG="$*"
            break
            ;;
    esac
done

# ========== 主流程 ==========

echo ""
echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}                        Git Save                                ${NC}"
echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
echo ""

# 获取当前分支
CURRENT_BRANCH=$(git branch --show-current)
echo -e "${BLUE}📂 当前分支: $CURRENT_BRANCH${NC}"

# 检查是否有远程跟踪分支
TRACKING_REMOTE=$(git config "branch.$CURRENT_BRANCH.remote" 2>/dev/null || echo "")
if [ -n "$TRACKING_REMOTE" ]; then
    echo -e "${BLUE}🔗 跟踪远程: $TRACKING_REMOTE/$CURRENT_BRANCH${NC}"
fi

echo ""

# 追加模式
if [ "$AMEND" = true ]; then
    echo -e "${YELLOW}[amend] 追加到上一次提交...${NC}"
    
    git add -A
    git commit --amend --no-edit
    
    if [ "$NO_PUSH" = false ]; then
        if [ "$FORCE_PUSH" = true ]; then
            git push --force-with-lease
            echo -e "${GREEN}✅ 已强制推送 (amend)${NC}"
        else
            git push
            echo -e "${GREEN}✅ 已推送 (amend)${NC}"
        fi
    else
        echo -e "${GREEN}✅ 已追加提交 (未推送)${NC}"
    fi
    
    exit 0
fi

# 检查提交信息
if [ -z "$COMMIT_MSG" ]; then
    echo -e "${RED}❌ 错误: 请提供提交信息${NC}"
    echo -e "${YELLOW}   用法: git save \"提交信息\"${NC}"
    
    # 询问是否使用默认信息
    read -p "是否使用默认提交信息 'update'？(y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        COMMIT_MSG="update"
    else
        exit 1
    fi
fi

# 检查是否有更改
if ! check_changes; then
    echo -e "${YELLOW}⚠️  没有检测到任何更改${NC}"
    exit 0
fi

# 显示更改摘要
show_changes_summary

# 确认操作
echo -e "${YELLOW}════════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}提交信息: ${GREEN}$COMMIT_MSG${NC}"
echo -e "${CYAN}推送目标: ${GREEN}$([ "$NO_PUSH" = true ] && echo "不推送" || echo "$([ "$PUSH_ALL" = true ] && echo "origin + gitlab" || echo "$TRACKING_REMOTE")")${NC}"
echo -e "${YELLOW}════════════════════════════════════════════════════════════════${NC}"
echo ""

read -p "确认执行？(y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${RED}操作已取消${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}[1/3] 添加所有更改...${NC}"
git add -A
echo -e "${GREEN}✅ 已添加所有更改${NC}"

echo -e "${YELLOW}[2/3] 提交更改...${NC}"
if git commit -m "$COMMIT_MSG"; then
    echo -e "${GREEN}✅ 提交成功${NC}"
else
    echo -e "${RED}❌ 提交失败${NC}"
    echo -e "${YELLOW}正在回滚 add 操作...${NC}"
    git reset -q HEAD -- .
    exit 1
fi

if [ "$NO_PUSH" = false ]; then
    echo -e "${YELLOW}[3/3] 推送到远程仓库...${NC}"
    
    PUSH_OPTIONS=""
    if [ "$FORCE_PUSH" = true ]; then
        PUSH_OPTIONS="--force-with-lease"
    fi
    
    if [ "$PUSH_ALL" = true ]; then
        # 推送到 origin
        if check_remote "origin"; then
            echo -e "${YELLOW}  推送 origin/$CURRENT_BRANCH...${NC}"
            git push origin "$CURRENT_BRANCH" $PUSH_OPTIONS
            echo -e "${GREEN}  ✅ origin 推送成功${NC}"
        fi
        
        # 推送到 gitlab
        if check_remote "gitlab"; then
            echo -e "${YELLOW}  推送 gitlab/$CURRENT_BRANCH...${NC}"
            git push gitlab "$CURRENT_BRANCH" $PUSH_OPTIONS
            echo -e "${GREEN}  ✅ gitlab 推送成功${NC}"
        fi
    else
        # 推送到默认跟踪远程
        git push $PUSH_OPTIONS
        echo -e "${GREEN}✅ 推送成功${NC}"
    fi
else
    echo -e "${YELLOW}[3/3] 跳过推送 (--no-push)${NC}"
fi

echo ""
echo -e "${GREEN}════════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}                    ✅ 完成！                                   ${NC}"
echo -e "${GREEN}════════════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${CYAN}提交: ${GREEN}$COMMIT_MSG${NC}"
echo -e "${CYAN}分支: ${GREEN}$CURRENT_BRANCH${NC}"