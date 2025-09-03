#!/bin/bash

# ─── Configuration & Color Definitions ────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
GRAY='\033[0;90m'
BRIGHT_RED='\033[1;31m'
BRIGHT_GREEN='\033[1;32m'
BRIGHT_YELLOW='\033[1;33m'
BRIGHT_BLUE='\033[1;34m'
BRIGHT_PURPLE='\033[1;35m'
BRIGHT_CYAN='\033[1;36m'
BG_BLUE='\033[44m'
BG_GREEN='\033[42m'
BG_RED='\033[41m'
NC='\033[0m'

# Backend API endpoint
API_URL="http://13.232.198.0:8080/login"

# ─── Helper Functions ─────────────────────────────────────────────────────────

# Force interactive mode for Git hooks
force_interactive() {
    if [[ ! -t 0 ]] || [[ ! -t 1 ]] || [[ ! -t 2 ]]; then
        exec < /dev/tty > /dev/tty 2>&1
    fi

    if [[ ! -t 0 ]]; then
        echo "Error: Cannot access terminal for input"
        exit 1
    fi
}

# Safe input function that works in Git hooks
safe_read() {
    local prompt="$1"
    local var_name="$2"
    local is_password="$3"

    echo -ne "$prompt"

    if [[ "$is_password" == "true" ]]; then
        read -s -r "$var_name" < /dev/tty
        echo
    else
        read -r "$var_name" < /dev/tty
    fi
}

# Simple centered text
print_centered() {
    local color="$1"
    local text="$2"
    local width=60
    local plain_text=$(echo -e "$text" | sed 's/\x1b\[[0-9;]*m//g')
    local len=${#plain_text}
    local pad=$(( (width - len) / 2 ))
    local extra=$(( width - len - pad*2 ))

    printf "${CYAN}║${NC}%*s" $pad ""
    echo -ne "${color}${text}${NC}"
    printf "%*s" $((pad + extra)) ""
    echo -e "${CYAN}║${NC}"
}

# Simple spinner
spinner() {
    local pid=$1
    local delay=0.1
    local spinstr='⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏'
    local colors=("${BRIGHT_CYAN}" "${BRIGHT_YELLOW}" "${BRIGHT_GREEN}")
    local i=0

    while kill -0 "$pid" 2>/dev/null; do
        local char_idx=$((i % ${#spinstr}))
        local color_idx=$((i % ${#colors[@]}))
        printf "${colors[color_idx]} [%c] ${NC}" "${spinstr:$char_idx:1}"
        sleep $delay
        printf "\b\b\b\b\b"
        ((i++))
    done
    printf "    \b\b\b\b"
}

# Simple progress bar
progress_bar() {
    local width=50
    local duration=1.5

    for ((i=0; i<=width; i++)); do
        local percent=$(( (i * 100) / width ))
        local filled=$(printf "%-${i}s" | tr ' ' '█')
        local empty=$(printf "%-$((width-i))s" | tr ' ' '▒')

        printf "\r${CYAN}[${GREEN}%s${GRAY}%s${CYAN}] ${YELLOW}%d%%${NC}" "$filled" "$empty" "$percent"
        sleep $(echo "scale=3; $duration / $width" | bc -l 2>/dev/null || echo "0.03")
    done
    echo
}

# Simple typewriter effect
typewriter() {
    local text="$1"
    local color="$2"
    local delay="${3:-0.03}"

    for (( i=0; i<${#text}; i++ )); do
        echo -ne "${color}${text:$i:1}${NC}"
        sleep "$delay"
    done
    echo
}

# ─── Main Script Logic - ALWAYS PROMPTS FOR CREDENTIALS ─────────────────────

# Force interactive mode first
force_interactive

# Clear screen
clear

# Check if credentials already exist (for informational purposes only)
EXISTING_EMAIL=$(git config --local user.challengeEmail 2>/dev/null)
if [[ -n "$EXISTING_EMAIL" ]]; then
    echo -e "${YELLOW}⚠️  Existing credentials found for: ${CYAN}$EXISTING_EMAIL${NC}"
    echo -e "${GRAY}   These will be replaced with new credentials...${NC}"
    echo
fi

# Title
echo -e "${BRIGHT_CYAN}"
cat <<'EOF'
++------------------------------------------------------------------------------------------------++
++------------------------------------------------------------------------------------------------++
||██████╗ ██╗   ██╗██╗██╗     ██████╗     ██╗    ██╗██╗████████╗██╗  ██╗     ██████╗ ██╗████████╗ ||
||██╔══██╗██║   ██║██║██║     ██╔══██╗    ██║    ██║██║╚══██╔══╝██║  ██║    ██╔════╝ ██║╚══██╔══╝ ||
||██████╔╝██║   ██║██║██║     ██║  ██║    ██║ █╗ ██║██║   ██║   ███████║    ██║  ███╗██║   ██║    ||
||██╔══██╗██║   ██║██║██║     ██║  ██║    ██║███╗██║██║   ██║   ██╔══██║    ██║   ██║██║   ██║    ||
||██████╔╝╚██████╔╝██║███████╗██████╔╝    ╚███╔███╔╝██║   ██║   ██║  ██║    ╚██████╔╝██║   ██║    ||
||╚═════╝  ╚═════╝ ╚═╝╚══════╝╚═════╝      ╚══╝╚══╝ ╚═╝   ╚═╝   ╚═╝  ╚═╝     ╚═════╝ ╚═╝   ╚═╝    ||
++------------------------------------------------------------------------------------------------++
++------------------------------------------------------------------------------------------------++
EOF

echo -e "${NC}"

echo
typewriter "🚀 Welcome to BuildWithGit Setup!" "${BRIGHT_WHITE}" 0.04
typewriter "    Authentication required for all users..." "${GRAY}" 0.02

echo
echo -e "${YELLOW}⚡ Initializing system...${NC}"
progress_bar
echo -e "${GREEN}✅ System ready!${NC}"

# ALWAYS require authentication - no conditions
echo
echo -e "${BRIGHT_WHITE}${BG_BLUE} 🔐 AUTHENTICATION REQUIRED ${NC}"
echo

echo -e "${BLUE}╭─────────────────────────────────────────────╮${NC}"
echo -e "${BLUE}│${NC} 📧 ${WHITE}Email Address${NC}                        ${BLUE}│${NC}"
echo -e "${BLUE}╰─────────────────────────────────────────────╯${NC}"

# Always prompt for email
safe_read "${CYAN}➤ ${NC}" USER_EMAIL false

echo
echo -e "${BLUE}╭─────────────────────────────────────────────╮${NC}"
echo -e "${BLUE}│${NC} 🔑 ${WHITE}Password${NC}                             ${BLUE}│${NC}"
echo -e "${BLUE}╰─────────────────────────────────────────────╯${NC}"

# Always prompt for password
safe_read "${CYAN}➤ ${NC}" USER_PASSWORD true

echo
echo -ne "${YELLOW}🔄 Authenticating... ${NC}"

# Authentication with spinner
RESPONSE=""
(
    RESPONSE=$(curl -s -X POST "$API_URL" \
      -H "Content-Type: application/json" \
      -d "{\"email\":\"$USER_EMAIL\",\"password\":\"$USER_PASSWORD\"}")
    echo "$RESPONSE" > /tmp/auth_response.tmp
) & PID=$!

spinner $PID
echo

# Read response
RESPONSE=$(cat /tmp/auth_response.tmp 2>/dev/null)
rm -f /tmp/auth_response.tmp

# Parse response
SUCCESS=$(echo "$RESPONSE" | jq -r '.success' 2>/dev/null)
TOKEN=$(echo "$RESPONSE" | jq -r '.token' 2>/dev/null)
EMAIL=$(echo "$RESPONSE" | jq -r '.email' 2>/dev/null)

if [[ "$SUCCESS" == "true" && -n "$TOKEN" && "$TOKEN" != "null" ]]; then
    # Success - ALWAYS save/replace credentials
    echo -e "${BRIGHT_GREEN}${BG_GREEN} ✓ ${NC} ${BRIGHT_GREEN}Authentication Successful!${NC}"
    echo
    echo -e "${WHITE}🎉 Welcome, ${CYAN}$EMAIL${NC}!"

    # Save/Replace credentials (no condition checking)
    git config --local user.challengeEmail "$EMAIL"
    git config --local user.challengeToken "$TOKEN"

    if [[ -n "$EXISTING_EMAIL" ]]; then
        typewriter "🔄 Credentials updated and replaced..." "${GRAY}" 0.02
    else
        typewriter "📝 New credentials saved locally..." "${GRAY}" 0.02
    fi

    echo

    # Instructions
    echo -e "${CYAN}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${NC}                    ${WHITE}INSTRUCTIONS${NC}                            ${CYAN}║${NC}"
    echo -e "${CYAN}╠════════════════════════════════════════════════════════════╣${NC}"
    echo -e "${CYAN}║${NC}                                                            ${CYAN}║${NC}"
    print_centered "${WHITE}" "1. Write your code solution 🎯"
    print_centered "${WHITE}" "2. Stage changes: ${YELLOW}git add .${NC}"
    print_centered "${WHITE}" "3. Test solution: ${GREEN}git commit -m \"test\"${NC}"
    print_centered "${WHITE}" "4. Submit final: ${PURPLE}git commit -m \"submit\"${NC}"
    echo -e "${CYAN}║${NC}                                                            ${CYAN}║${NC}"
    echo -e "${CYAN}╚════════════════════════════════════════════════════════════╝${NC}"

    echo
    echo -e "${BRIGHT_GREEN}🚀 Ready to code! Good luck!${NC}"

else
    # Failure
    echo -e "${BRIGHT_RED}${BG_RED} ✗ ${NC} ${BRIGHT_RED}Authentication Failed${NC}"

    ERROR_MSG=$(echo "$RESPONSE" | jq -r '.error' 2>/dev/null)
    echo -e "${RED}❌ Access denied!${NC}"
    echo

    if [[ -n "$ERROR_MSG" && "$ERROR_MSG" != "null" ]]; then
        echo -e "${YELLOW}Reason: ${ERROR_MSG}${NC}"
    else
        echo -e "${YELLOW}Please check:${NC}"
        echo -e "${GRAY}• Your email and password${NC}"
        echo -e "${GRAY}• Server connection (localhost:8080)${NC}"
    fi

    echo
    echo -e "${RED}Please try again...${NC}"
    exit 1
fi

exit 0
