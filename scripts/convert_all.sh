#!/bin/bash

# Help message function
usage() {
    echo "Usage: $0 -s <source_dir> -d <dest_dir> -t <token> [options]"
    echo "  -s, --src      Source directory containing Markdown files"
    echo "  -d, --dest     Destination directory for generated HTML"
    echo "  -t, --token    Authentication token for protected assets"
    echo "  -p, --path     Path to be used as root for all docs (default: /docs)"
    echo "  -c, --css-dir  Directory containing static assets (CSS, JS, etc.)"
    echo "  -h, --help     Show this help message"
    echo ""
    echo "Example:"
    echo "  $0 -s courses/ -d output/ -t 'abc123...' -c /path/to/static"
    exit 1
}

# Function to get the appropriate mktemp command
get_mktemp_cmd() {
    if command -v gmktemp >/dev/null 2>&1; then
        echo "gmktemp"
    else
        # Check if we're on macOS and warn about gmktemp
        if [[ "$OSTYPE" == "darwin"* ]]; then
            echo "Warning: gmktemp not found on macOS. The script may not work properly." >&2
            echo "Please install gmktemp with: brew install coreutils" >&2
        fi
        echo "mktemp"
    fi
}

# Function to create a temporary file with suffix
create_temp_with_suffix() {
    local suffix="$1"
    local mktemp_cmd=$(get_mktemp_cmd)

    if [[ "$mktemp_cmd" == "gmktemp" ]]; then
        # GNU mktemp supports --suffix
        TMPDIR="$file_dir" gmktemp --suffix="$suffix"
    else
        # Standard mktemp (macOS) - create temp file and rename
        local temp_file=$(TMPDIR="$file_dir" mktemp)
        local new_name="${temp_file}${suffix}"
        mv "$temp_file" "$new_name"
        echo "$new_name"
    fi
}

# Parse command-line options
SRC_DIR=""
DEST_DIR=""
TOKEN=""
DOCS_PATH="/docs"
CSS_DIR=""
while [[ $# -gt 0 ]]; do
    case "$1" in
    -s | --src)
        SRC_DIR="$2"
        shift 2
        ;;
    -d | --dest)
        DEST_DIR="$2"
        shift 2
        ;;
    -t | --token)
        TOKEN="$2"
        shift 2
        ;;
    -p | --path)
        DOCS_PATH="$2"
        shift 2
        ;;
    -c | --css-dir)
        CSS_DIR="$2"
        shift 2
        ;;
    -h | --help)
        usage
        ;;
    *)
        echo "Unknown option: $1"
        usage
        ;;
    esac
done

# Check required parameters
if [[ -z "$SRC_DIR" || -z "$DEST_DIR" || -z "$TOKEN" ]]; then
    echo "Error: Source directory, destination directory, and token are required."
    usage
fi

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Path to your puppeteer-config.json file for CI/CD environments.
PUPPETEER_CONFIG_PATH="${SCRIPT_DIR}/puppeteer-config.json"
if [ ! -f "$PUPPETEER_CONFIG_PATH" ]; then
    echo "Warning: puppeteer-config.json not found. This may cause errors in CI."
fi

# Create a temporary Lua filter with the actual token
temp_lua=$(mktemp)
trap 'rm -f "$temp_lua"' EXIT
sed "s|TOKEN_PLACEHOLDER|$TOKEN|g" "${SCRIPT_DIR}/add_token.lua" > "$temp_lua"

# Pandoc options (no mermaid filter needed)
PANDOC_OPTS="--standalone --css ${DOCS_PATH}/static/css/minio_docs.css --include-after-body=${SCRIPT_DIR}/copy_btn.html --lua-filter=$temp_lua"

# Create courses output directory
COURSES_DIR="$DEST_DIR/courses"
mkdir -p "$COURSES_DIR"

# Find all .md files and process them
find "$SRC_DIR" -type f -name "*.md" | while read -r file; do
    echo "Processing $file..."

    # Define a temporary markdown file for the processed output
    file_dir=$(dirname "$file")
    processed_md=$(create_temp_with_suffix ".md")

    # 1. Use mmdc to convert mermaid diagrams and create a new markdown file
    mmdc -i "$file" -o "$processed_md" --puppeteerConfigFile "$PUPPETEER_CONFIG_PATH"

    # --- Compute paths for Pandoc ---
    # Compute the relative path
    rel_path="${file#"$SRC_DIR"/}"
    dest_dir="$COURSES_DIR/$(dirname "$rel_path")"
    mkdir -p "$dest_dir"

    base_name="$(basename "$file" .md)"
    output_html="$dest_dir/$base_name.html"

    # 2. Run Pandoc on the processed markdown file from mmdc
    pandoc $PANDOC_OPTS "$processed_md" -o "$output_html"
    echo "Converted $file -> $output_html"

    # Clean up the temporary processed file
    rm "$processed_md"
done

# Copy all non-MD files to courses directory (preserving directory structure)
find "$SRC_DIR" -type f ! -name "*.md" | while read -r file; do
    rel_path="${file#"$SRC_DIR"/}"
    dest_file="$COURSES_DIR/$rel_path"
    dest_dir="$(dirname "$dest_file")"
    mkdir -p "$dest_dir"
    cp "$file" "$dest_file"
    echo "Copied $file -> $dest_file"
done

# Copy static assets if CSS_DIR is provided
if [[ -n "$CSS_DIR" && -d "$CSS_DIR" ]]; then
    STATIC_DIR="$DEST_DIR/static"
    mkdir -p "$STATIC_DIR"
    
    echo "Copying static assets from $CSS_DIR to $STATIC_DIR..."
    cp -r "${CSS_DIR:?}"/* "$STATIC_DIR/"
    echo "Static assets copied successfully."
else
    if [[ -n "$CSS_DIR" ]]; then
        echo "Warning: CSS directory '$CSS_DIR' does not exist, skipping static assets."
    else
        echo "No CSS directory specified, skipping static assets."
    fi
fi

echo "All files converted successfully."
