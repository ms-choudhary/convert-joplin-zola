# Convert Joplin Notes to Zola Pages

Joplin is a notes taking application I like to use for all my notes. I run a personal static website (on github pages) using Zola. This tool is the missing piece so I can start writing my Posts in Joplin, and then serve them statically via Zola. 

## Usage

```
go build

# reads notes from joplin website directory and writes them to zola (accounting for all changes)
convert-joplin-zola
```
