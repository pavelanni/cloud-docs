#!/bin/bash

# Configurable variables
SITE="cloud-docs-server-679412990936.us-central1.run.app/docs"

# Check if 1Password CLI is installed
if ! command -v op &> /dev/null; then
    echo "Error: 1Password CLI (op) is not installed."
    echo "To install it:"
    echo "  - macOS: brew install 1password-cli"
    echo "  - Linux: See https://developer.1password.com/docs/cli/get-started#install"
    echo "  - Windows: See https://developer.1password.com/docs/cli/get-started#install"
    exit 1
fi

# Get token from 1Password
TOKEN=$(op read op://Training/cloud-docs/token 2>/dev/null)
if [ -z "$TOKEN" ]; then
    echo "Error: Failed to retrieve token from 1Password."
    echo "Make sure you are signed in to 1Password CLI (run 'op signin' if needed)."
    exit 1
fi

# Usage function
usage() {
    echo "Usage: $0 <file-path>"
    exit 1
}

# Check for argument
if [ $# -ne 1 ]; then
    usage
fi

INPUT_PATH="$1"

# Get project root (directory containing 'courses')
PROJECT_ROOT=$(git rev-parse --show-toplevel 2>/dev/null)
if [ -z "$PROJECT_ROOT" ]; then
    echo "Error: Could not determine project root. Are you in a git repo?"
    exit 1
fi

# Absolute path to input file
if [[ "$INPUT_PATH" = /* ]]; then
    ABS_PATH="$INPUT_PATH"
else
    ABS_PATH="$PWD/$INPUT_PATH"
fi

# Check if file exists
if [ ! -e "$ABS_PATH" ]; then
    echo "Error: File not found: $INPUT_PATH"
    exit 1
fi

# Path relative to project root using Python for cross-platform compatibility
REL_PATH=$(python3 -c "import os; print(os.path.relpath(os.path.abspath('$ABS_PATH'), '$PROJECT_ROOT'))")

# Ensure the path starts with 'courses/'
if [[ "$REL_PATH" == courses/* ]]; then
    URL_PATH="$REL_PATH"
else
    # Try to find the path relative to 'courses/'
    # Find the 'courses' directory in the path
    COURSES_INDEX=$(echo "$REL_PATH" | grep -b -o 'courses/' | head -n1 | cut -d: -f1)
    if [ -n "$COURSES_INDEX" ]; then
        URL_PATH="${REL_PATH:$COURSES_INDEX}"
    else
        echo "Error: File must be inside the 'courses' directory."
        exit 1
    fi
fi

# Change .md extension to .html if needed
if [[ "$URL_PATH" == *.md ]]; then
    URL_PATH="${URL_PATH%.md}.html"
fi

# Compose the iframe URL
IFRAME_URL="https://$SITE/$URL_PATH?token=$TOKEN"

# Print the iframe HTML
cat <<EOF
<iframe src="$IFRAME_URL" name="myiFrame" scrolling="yes" frameborder="0" marginheight="0px" marginwidth="0px" allowfullscreen="true" allow="clipboard-write"></iframe>
EOF
