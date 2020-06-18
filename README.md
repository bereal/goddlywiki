# goddlywiki
A standalone app that serves TiddlyWiki from a local WebDav allowing seamless save experience

## Installation:

  go get github.com/bereal/goddlywiki/cmd/goddlywiki

## Usage

    goddlywiki [OPTIONS]

    -h string
            home directory (default "/home/bereal/.tiddly")
    -n string
            wiki name (default "default")
    -f string
        wiki file (overrides both -n and -h)
    -o	open in the browser
    -p int
            port (default 8080)
