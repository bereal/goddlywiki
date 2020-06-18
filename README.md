# goddlywiki
A self-contained app that serves TiddlyWiki from a local WebDAV server with nicely working saving

## Installation:

    go get github.com/bereal/goddlywiki

## Usage

    goddlywiki start [OPTIONS]

    -h string  home directory (default "${HOME}/.tiddly")
    -n string  wiki name (default "default")
    -f string  wiki file (overrides both -n and -h)
    -p int     port (default 8080)
    -o         open in a browser
    -d         run as a daemon


    goddlywiki stop


    goddlywiki create [OPTIONS]

    -f string  wiki file (overrides both -n and -h)
    -h string  home directory (default "${HOME}/.tiddly")
    -n string  wiki name (default "default")
