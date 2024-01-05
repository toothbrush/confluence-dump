# confluence-dump

This little tool is inspired by the excellent [`ghorg`](https://github.com/gabrie30/ghorg), and aims
to keep a local "database" of Markdown files in sync with a remote Atlassian Confluence instance.

It has many rough edges still, but here's roughly how you can use it:

1. Get a "Personal Access Token" from your Atlassian account, see
   https://id.atlassian.com/manage-profile/security/api-tokens, and store it safely.  I use the
   `pass` password manager.
2. Install this tool: `go install -u github.com/toothbrush/confluence-dump`
3. Create a rudimentary config file at `~/.config/confluence-dump.yaml` (or elsewhere with the
   `--config` flag) and populate it.  For now the only documentation is the example config file in
   this repo.

## TODO

* Parallelise downloading among a few workers
* Set a base href for Markdown conversion so links work
* Contexts on slow API calls
* wrap API in a struct that has a logger and a context
* remove all prints from anything not in `package main`
* LocalMarkdownCache map should probably be a field in a in a struct, in preparation for parallel stuff

## DONE

* Do not download pages we already have
* Download all users' blog posts, optionally
* Config/flags/env with Cobra
