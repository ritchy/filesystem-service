#!/bin/bash
INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command')

if echo "$COMMAND" | grep -qE "rm -rf[[:space:]]+(/|~|\$HOME)|:(){ :|:& };:"; then
  echo "Blocked: dangerous command pattern detected" >&2
  exit 2
fi
exit 0
