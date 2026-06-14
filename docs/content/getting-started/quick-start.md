---
title: "Quick start"
description: "Run your first kr36 command."
weight: 30
---

Once `kr36` is on your `PATH`:

```bash
kr36 --help       # see the command tree
kr36 version      # build info
```

This is a fresh scaffold, so the command tree is just `version` for now. Add
your first real command in `cli/`, build on the `kr36-cli` library package,
and document it here.

A good first command usually fetches one thing and prints it as JSON, so the
output pipes straight into `jq` and the rest of your tools.
