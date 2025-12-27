#!/bin/bash

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TOOLS_JSON="$SCRIPT_DIR/tools.json"

# Get the project root directory (parent of scripts directory)
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Log file for debugging
LOG_FILE="/tmp/mcp-server-bash.log"

echo "Starting Backend JSON-Driven MCP Server $(date)" >> "$LOG_FILE"
echo "Script directory: $SCRIPT_DIR" >> "$LOG_FILE"
echo "Project root: $PROJECT_ROOT" >> "$LOG_FILE"

# Function to substitute parameters in command template
substitute_parameters() {
    local command_template="$1"
    local params_json="$2"
    local result="$command_template"
    
    echo "Original command: $command_template" >> "$LOG_FILE"
    echo "Parameters: $params_json" >> "$LOG_FILE"
    
    # Handle parameters with default values (e.g., {{package:=./...}})
    while [[ "$result" =~ \{\{([^}:]+):=([^}]+)\}\} ]]; do
        local param_name="${BASH_REMATCH[1]}"
        local default_value="${BASH_REMATCH[2]}"
        local param_value
        
        param_value=$(echo "$params_json" | jq -r --arg name "$param_name" '.[$name] // empty' 2>/dev/null)
        
        if [[ -z "$param_value" || "$param_value" == "null" ]]; then
            param_value="$default_value"
        fi
        
        result="${result//\{\{${param_name}:=${default_value}\}\}/$param_value}"
        echo "Substituted $param_name with default: $param_value" >> "$LOG_FILE"
    done
    
    # Handle regular parameters (e.g., {{num1}})
    while [[ "$result" =~ \{\{([^}]+)\}\} ]]; do
        local param_name="${BASH_REMATCH[1]}"
        local param_value
        
        param_value=$(echo "$params_json" | jq -r --arg name "$param_name" '.[$name] // empty' 2>/dev/null)
        
        if [[ -z "$param_value" || "$param_value" == "null" ]]; then
            echo "Error: Missing required parameter: $param_name" >> "$LOG_FILE"
            echo "ERROR_MISSING_PARAM:$param_name"
            return 1
        fi
        
        result="${result//\{\{${param_name}\}\}/$param_value}"
        echo "Substituted $param_name: $param_value" >> "$LOG_FILE"
    done
    
    echo "Final command: $result" >> "$LOG_FILE"
    echo "$result"
}

# Function to execute command and return formatted result
execute_tool_command() {
    local command="$1"
    local tool_name="$2"
    
    echo "Executing tool '$tool_name': $command" >> "$LOG_FILE"
    echo "Working directory: $PROJECT_ROOT" >> "$LOG_FILE"
    
    # Execute the command and capture both stdout and stderr
    # Change to project root before executing to ensure make commands work correctly
    local output
    local exit_code
    
    output=$(cd "$PROJECT_ROOT" && eval "$command" 2>&1)
    exit_code=$?
    
    echo "Exit code: $exit_code" >> "$LOG_FILE"
    echo "Output length: ${#output} chars" >> "$LOG_FILE"
    
    # Escape the output for JSON
    local escaped_output
    escaped_output=$(echo "$output" | jq -R -s .)
    
    if [[ $exit_code -eq 0 ]]; then
        echo '{"content":[{"type":"text","text":'"$escaped_output"'}],"isError":false}'
    else
        echo '{"content":[{"type":"text","text":'"$escaped_output"'}],"isError":true}'
    fi
}

# Function to generate tools list from JSON
generate_tools_list() {
    if [[ ! -f "$TOOLS_JSON" ]]; then
        echo "Error: tools.json not found at $TOOLS_JSON" >> "$LOG_FILE"
        echo '{"tools":[]}'
        return
    fi
    
    # Remove the handler field and return clean tools list
    jq -c '[.tools[] | del(.handler)]' "$TOOLS_JSON"
}

while read -r line; do
    echo "$line" >> "$LOG_FILE"
    # Parse JSON input using jq
    method=$(echo "$line" | jq -r '.method' 2>/dev/null)
    id=$(echo "$line" | jq -r '.id' 2>/dev/null)
    
    if [[ "$method" == "initialize" ]]; then
        echo '{"jsonrpc":"2.0","id":'"$id"',"result":{"protocolVersion":"2024-11-05","capabilities":{"experimental":{},"prompts":{"listChanged":false},"resources":{"subscribe":false,"listChanged":false},"tools":{"listChanged":false}},"serverInfo":{"name":"json-driven-mcp-server","version":"2.0.0"}}}'
    
    elif [[ "$method" == "notifications/initialized" ]]; then
        : #do nothing
    
    elif [[ "$method" == "tools/list" ]]; then
        tools_list=$(generate_tools_list)
        echo '{"jsonrpc":"2.0","id":'"$id"',"result":{"tools":'"$tools_list"'}}'
    
    elif [[ "$method" == "resources/list" ]]; then
        echo '{"jsonrpc":"2.0","id":'"$id"',"result":{"resources":[]}}'

    elif [[ "$method" == "prompts/list" ]]; then
        echo '{"jsonrpc":"2.0","id":'"$id"',"result":{"prompts":[]}}'

    elif [[ "$method" == "tools/call" ]]; then
        # Parse tool call parameters
        tool_name=$(echo "$line" | jq -r '.params.name' 2>/dev/null)
        arguments=$(echo "$line" | jq -c '.params.arguments // {}' 2>/dev/null)
        
        echo "Tool call: $tool_name" >> "$LOG_FILE"
        
        # Get tool configuration from JSON
        tool_config=$(jq -r --arg name "$tool_name" '.tools[] | select(.name == $name)' "$TOOLS_JSON" 2>/dev/null)
        
        if [[ -z "$tool_config" || "$tool_config" == "null" ]]; then
            echo '{"jsonrpc":"2.0","id":'"$id"',"error":{"code":-32602,"message":"Tool not found: '"$tool_name"'"}}'
            continue
        fi
        
        # Get the command template from the tool configuration
        command_template=$(echo "$tool_config" | jq -r '.command')
        
        # Substitute parameters in the command
        final_command=$(substitute_parameters "$command_template" "$arguments")
        
        # Check if substitution failed
        if [[ "$final_command" =~ ^ERROR_MISSING_PARAM: ]]; then
            missing_param="${final_command#ERROR_MISSING_PARAM:}"
            echo '{"jsonrpc":"2.0","id":'"$id"',"error":{"code":-32602,"message":"Missing required parameter: '"$missing_param"'"}}'
            continue
        fi
        
        # Execute the command and get the result
        result=$(execute_tool_command "$final_command" "$tool_name")
        
        echo '{"jsonrpc":"2.0","id":'"$id"',"result":'"$result"'}'
    
    else
        echo '{"jsonrpc":"2.0","id":'"$id"',"error":{"code":-32601,"message":"Method not found"}}'
    fi
done || break