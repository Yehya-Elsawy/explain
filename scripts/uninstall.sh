#!/usr/bin/env bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${CYAN}"
echo "  ┌────────────────────────────────────────┐"
echo "  │       Uninstalling explain CLI         │"
echo "  └────────────────────────────────────────┘"
echo -e "${NC}"

TARGETS=("/usr/local/bin/explain" "$HOME/.local/bin/explain")
if command -v explain >/dev/null 2>&1; then
    TARGETS+=("$(command -v explain)")
fi

removed=false
declare -A SEEN

for target in "${TARGETS[@]}"; do
    if [ -z "${SEEN[$target]}" ] && [ -f "$target" ]; then
        SEEN[$target]=1
        if [ -w "$target" ]; then
            rm -f "$target"
        else
            if command -v sudo >/dev/null 2>&1; then
                sudo rm -f "$target"
            else
                rm -f "$target"
            fi
        fi
        echo -e "${GREEN}✓ Removed: $target${NC}"
        removed=true
    fi
done

if [ "$removed" = true ]; then
    echo
    echo -e "${GREEN}Successfully uninstalled explain CLI from your system.${NC}"
    echo -e "Thank you for trying explain!"
    echo
else
    echo -e "${YELLOW}explain binary not found in standard paths.${NC}"
fi
