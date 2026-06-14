---
title: "Introduction"
description: "What kr36 is and how it is put together."
weight: 10
---

Browse 36kr Chinese tech news via RSS (36氪)

kr36 is a single binary. It speaks to kr36-cli over plain HTTPS,
shapes the responses into clean records, and gets out of your way. There is
nothing to sign up for and nothing to run alongside it.

## How it is built

- A **library package** (`kr36-cli`) holds the HTTP client and the typed
  data models. It paces requests, sets an honest User-Agent, and retries the
  transient failures any public site throws under load.
- A **command tree** (`cli`) wraps the library in subcommands with shared
  output formats and flags.
- One **`cmd/kr36`** entry point ties them together.

## Scope

kr36 is a read-only client over data kr36-cli already serves
publicly. It reads that data and shapes it for you. That narrow scope keeps it a
single small binary with no database, no daemon, and no setup.

Next: [install it](/getting-started/installation/), then take the
[quick start](/getting-started/quick-start/).
