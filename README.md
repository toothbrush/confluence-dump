## TODO

* Download users' blog posts
* Config/flags/env with Cobra
* Parallelise downloading among a few workers

## DONE

* Do not download pages we already have

### Questions

* Do we mind extracting org and/or space key from the file path?  E.g.,
  `~/confluence/redbubble/CORE/some/page.md`?  I think we can live with it.  Might also mean we have
  slightly less passing around of info to do.
