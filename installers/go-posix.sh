#!/usr/bin/env bash
command -v go >/dev/null 2>&1 || { 
    printf "\e[33mGo is required but not installed\e[0m\n" >&2
    rnm=1
}
command -v git >/dev/null 2>&1 || { 
    printf "\e[33mGit is required but not installed\e[0m\n" >&2
    rnm=1
}
[[ "$rnm" == "1" ]] && exit 1

[ -d "$HOME/.local/bin" ] && {
    mkdir --parent "$HOME/.local/bin"
}

[ -d ~/.lon-tool-src ] && rm ~/.lon-tool-src -rf
git clone https://git.timoxa0.su/timoxa0/lon-tool.git ~/.lon-tool-src || {
    rm ~/.lon-tool-src -rf
    exit 2
}

pushd ~/.lon-tool-src &> /dev/null
rev=$(git describe --abbrev=4 --dirty --always --tags)
go get git.timoxa0.su/timoxa0/lon-tool/cmd
go build -ldflags "-X git.timoxa0.su/timoxa0/lon-tool/cmd.version=$rev" -o "$HOME/.local/bin/lon-tool" main.go && {
    printf "\e[32mDone!\e[0m Installed at %s\n" "$HOME/go/bin/lon-tool"
}
popd &> /dev/null
