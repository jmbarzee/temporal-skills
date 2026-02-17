#!/bin/bash

# Version management
VERSION_REGEX="^v?[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$"

# Get latest version
get_latest_version() {
    git tag -l "v*" | grep -E "$VERSION_REGEX" | sort -V | tail -1 | sed 's/v//'
}

# Validate version format
validate_version() {
    echo "$1" | grep -E "$VERSION_REGEX" > /dev/null && echo "valid" || echo "invalid"
}

# Bump version based on type
bump_version() {
    local version=$1
    local type=$2
    
    echo "$version" | awk -F. -v type="$type" '{
        if (type == "major") {
            print $1+1".0.0"
        } else if (type == "minor") {
            print $1"."$2+1".0"
        } else if (type == "patch") {
            print $1"."$2"."$3+1
        } else {
            print "invalid"
        }
    }'
}

# Main function to determine version
determine_version() {
    local version=$1
    local type=$2
    
    if [ -z "$version" ]; then
        current=$(get_latest_version)
        if [ -z "$current" ]; then
            echo "1.0.0"
        else
            if [ -z "$type" ]; then
                echo "Error: TYPE must be specified (major, minor, or patch) when VERSION is not provided" >&2
                exit 1
            fi
            bumped=$(bump_version "$current" "$type")
            if [ "$bumped" = "invalid" ]; then
                echo "Error: Invalid version bump type '$type'. Must be major, minor, or patch" >&2
                exit 1
            fi
            echo "$bumped"
        fi
    else
        version=$(echo "$version" | sed 's/^v//')
        if [ "$(validate_version "$version")" != "valid" ]; then
            echo "Error: Invalid version format '$version'. Must match semver format (e.g., 1.0.0)" >&2
            exit 1
        fi
        echo "$version"
    fi
}

# If script is called directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    determine_version "$@"
fi 