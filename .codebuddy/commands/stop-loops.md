---
description: Stop all loop engineering maintenance loops
allowed-tools: Read, Bash
---

# Stop Loop Engineering

Stop all running maintenance loops.

Use CronList to list all active cron jobs, then use CronDelete to delete each one.
After deleting all, confirm with CronList that none remain.
