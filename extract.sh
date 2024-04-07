#!/usr/bin/env bash

if [ -f $1 ] ; then
    case $1 in
        *.tar.bz2)   tar xf $1     ;;
        *.tar.gz)    tar xf $1     ;;
        *.tar.xz)    tar xf $1     ;;
        *.tar.zst)   tar xf $1     ;;
        *.tar)       tar xf $1     ;;
        *.tbz)       tar xf $1     ;;
        *.tbz2)      tar xf $1     ;;
        *.tgz)       tar xf $1     ;;
        *.gz)        gunzip $1     ;;
        *.zip)       unzip -q $1   ;;
        *.7z)        7z x $1       ;;
        *)           echo "'$1' cannot be extracted" ;;
    esac
else
    echo "'$1' is not a valid file"
fi

