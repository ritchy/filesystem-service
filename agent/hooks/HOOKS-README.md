
# AI Hooks

Agent hooks are like git hooks, but for your AI agent. There are defined lifecycles like pre-build, build, post build.

You define handlers, shell commands, LLM prompts or sub-agents

See https://code.claude.com/docs/en/hooks-guide

## Types
 - Command Hooks
 - Prompt Hooks
 - Agent Hooks

# Command hooks 
run shell commands as child processes. They receive JSON on stdin with the session ID, transcript path, working directory, tool name, input parameters, and tool response for post-execution hooks.

PostToolUse -> example: manually running gofmt or prettier
PreToolUse -> example: inspect every bash command before it runs

## Example of PreTool dangerous command blocker as a bash script:
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit",
        "hooks": [
          { "type": "command", "command": "danger.sh" || true" }
        ]
      }
    ]
  }
}

## danger.sh
#!/bin/bash
INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command')

if echo "$COMMAND" | grep -qE "rm -rf[[:space:]]+(/|~|\$HOME)|:(){ :|:& };:"; then
  echo "Blocked: dangerous command pattern detected" >&2
  exit 2
fi
exit 0

# Prompt hooks 
send a text prompt to a fast agent model for single-turn semantic evaluation. The $ARGUMENTS placeholder injects the hook's input JSON. This is how you get intelligent, context-aware decisions without writing custom scripts. 

Here is an example:

{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          {
            "type": "prompt",
            "prompt": "Analyze this context: $ARGUMENTS. Are all tasks complete and were tests run and passing? Respond with {\"decision\": \"approve\"} or {\"decision\": \"block\", \"reason\": \"explanation\"}.",
            "timeout": 30
          }
        ]
      }
    ]
  }
}


# Agent hooks (Claud specific)

spawn a sub-agent with access to tools like Read, Grep, and Glob for multi-turn codebase verification. This is the heaviest handler type, suitable for deep validation like confirming that all modified files have corresponding test coverage.


# Stages
 - PreToolUse
 - PostToolUse handle the before and after of tool execution. 
 - PostToolUseFailure fires on tool errors. 
 - PermissionRequest fires when the user would normally see a permission dialog.
 - SessionStart and SessionEnd handle session lifecycle, with SessionStart being particularly useful because its stdout becomes Claude's context. 
 - Stop and SubagentStop fire when the agent or sub-agent finishes, and can force continuation. 
 - SubagentStart fires when a sub-agent spawns.

<!-- 
-->