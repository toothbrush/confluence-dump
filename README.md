# confluence-dump

This little tool is inspired by the excellent [`ghorg`](https://github.com/gabrie30/ghorg), and aims
to keep a local "database" of Markdown files in sync with a remote Atlassian Confluence instance.

It has many rough edges still, but here's roughly how you can use it:

1. Get a "Personal Access Token" from your Atlassian account, see
   https://id.atlassian.com/manage-profile/security/api-tokens, and store it safely.  I use the
   `pass` password manager.
1. Clone this repo
1. Install the tool: `go install ./cmd/confluence-dump`
1. Create a rudimentary config file at `~/.config/confluence-dump.yaml` (or elsewhere with the
   `--config` flag) and populate it.  For now the only documentation is [the example config file in
   this repo](./confluence-dump.yaml).
1. Run it! `confluence-dump download` 🎉

## TODO

* Set a base href for Markdown conversion so links work

## DONE

* Do not download pages we already have
* Download all users' blog posts, optionally
* Config/flags/env with Cobra
* remove all prints from anything not in `package main`
* wrap API in a struct that has a logger
* Contexts on slow API calls
* Parallelise downloading among a few workers
* LocalMarkdownCache map is a field in a struct, in preparation for parallel stuff
* prune
* progress output
* Option to skip archived/non-current content.
* retry logic
* post-cmd to ... do things (e.g. clean empty directories or touch `.projectile`)
* You can put the config file anywhere and set `CONFLUENCE_DUMP_CONFIG` to point there
* hmmmmmmm does renaming a file's ancestor get reflected??
* deal with "personal spaces".. somehow
