#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DESIGN_DIR="$REPO_ROOT/design"
ENV_FILE="$DESIGN_DIR/.env.publish"

usage() {
    cat <<EOF
Usage: $0 [--publish <major|minor|patch>]

Options:
  --publish <version>  Publish after building (requires $ENV_FILE)
                       Version: major, minor, or patch (default: patch)
  -h, --help           Show this help
EOF
    exit 1
}

PUBLISH=""
while [[ $# -gt 0 ]]; do
    case $1 in
        --publish)
            PUBLISH="${2:-patch}"
            if [[ ! "$PUBLISH" =~ ^(major|minor|patch)$ ]]; then
                echo "Invalid version: $PUBLISH"; usage
            fi
            shift 2 || shift
            ;;
        -h|--help) usage ;;
        *) echo "Unknown option: $1"; usage ;;
    esac
done

# Build
echo "Building Go binaries..."
cd "$DESIGN_DIR/lsp"
go build ./...
go install ./cmd/twf

echo "Building visualizer..."
cd "$DESIGN_DIR/visualizer"
npm run build:webview

echo "Compiling extension..."
cd "$DESIGN_DIR/editors/vscode"
npm run compile

# Publish
if [ -n "$PUBLISH" ]; then
    if [ ! -f "$ENV_FILE" ]; then
        echo "Error: $ENV_FILE not found"; exit 1
    fi
    source "$ENV_FILE"
    if [ -z "${VSCE_TOKEN:-}" ] || [ -z "${OVSX_TOKEN:-}" ]; then
        echo "Error: VSCE_TOKEN and OVSX_TOKEN required"; exit 1
    fi

    echo "Packaging $PUBLISH release..."
    vsce package "$PUBLISH"
    VSIX=$(ls -t *.vsix | head -1)

    echo "Publishing to VS Code Marketplace..."
    vsce publish --packagePath "$VSIX" -p "$VSCE_TOKEN"

    echo "Publishing to Open VSX..."
    npx ovsx publish "$VSIX" -p "$OVSX_TOKEN"

    rm "$VSIX"
    echo "Done!"
else
    echo "Done. Use --publish to release."
fi
