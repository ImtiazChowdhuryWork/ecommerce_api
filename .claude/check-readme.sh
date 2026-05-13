#!/usr/bin/env bash
# Called by the PostToolUse hook after Write/Edit operations.
# Reads the hook JSON from stdin, checks if a .go or .sql file was modified,
# and injects a README.md update reminder into Claude's context.
f=$(jq -r '.tool_input.file_path // ""')
if [[ "$f" =~ \.(go|sql)$ ]]; then
  printf '{"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":"REMINDER: %s was just modified. If this change affects the API surface, endpoints, request/response shapes, business rules, setup steps, or project structure, update README.md now before moving on."}}\n' "$f"
fi
