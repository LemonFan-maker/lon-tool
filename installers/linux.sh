#!/usr/bin/env bash
URL="https://git.timoxa0.su/timoxa0/lon-tool/releases/download/latest/lon-tool_lin_amd64"

[[ "$(uname -m)" != "x86_64*" ]] || {
    printf "Unsupported CPU arch\n"
    exit 1
}

[ -d "$HOME/.local/bin" ] && {
    mkdir --parent "$HOME/.local/bin"
}

printf "Downloading lon-tool\n"
curl "$URL" -o "$HOME/.local/bin/lon-tool" -#
chmod +x "$HOME/.local/bin/lon-tool"

command -v adb >/dev/null 2>&1 || { 
    printf "\e[33mWARNING: adb binary not found.\e[0m\nPlease install google platform tools using your package manager\n" >&2
}

printf "\e[32mDone!\e[0m Installed at %s\n" "$HOME/.local/bin/lon-tool"