#!/bin/bash
# Update version in main.go and commit to git.
# Usage:
#   ./update_version.sh           - increment last section (0.5.1 -> 0.5.2)
#   ./update_version.sh -v 0.6.0  - set version to 0.6.0

set -e

MAIN_GO="main.go"
if [ ! -f "$MAIN_GO" ]; then
	echo "main.go not found in current directory"
	exit 1
fi

get_current_version() {
	grep -oE 'Version = "[0-9]+\.[0-9]+\.[0-9]+"' "$MAIN_GO" | head -1 | sed 's/Version = "\(.*\)"/\1/'
}

increment_last_section() {
	local ver="$1"
	local last="${ver##*.}"
	local rest="${ver%.*}"
	echo "${rest}.$((last + 1))"
}

NEW_VERSION=""
while getopts "v:" opt; do
	case $opt in
		v) NEW_VERSION="$OPTARG" ;;
		*) exit 1 ;;
	esac
done

if [ -z "$NEW_VERSION" ]; then
	CURRENT=$(get_current_version)
	if [ -z "$CURRENT" ]; then
		echo "Could not find Version const in $MAIN_GO"
		exit 1
	fi
	NEW_VERSION=$(increment_last_section "$CURRENT")
	echo "Incrementing $CURRENT -> $NEW_VERSION"
else
	if [[ ! "$NEW_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
		echo "Invalid version format. Use X.Y.Z (e.g. 0.6.0)"
		exit 1
	fi
	echo "Setting version to $NEW_VERSION"
fi

if [[ "$OSTYPE" == "darwin"* ]]; then
	sed -i '' "s/Version = \"[0-9]*\.[0-9]*\.[0-9]*\"/Version = \"$NEW_VERSION\"/" "$MAIN_GO"
else
	sed -i "s/Version = \"[0-9]*\.[0-9]*\.[0-9]*\"/Version = \"$NEW_VERSION\"/" "$MAIN_GO"
fi

if ! git rev-parse --git-dir >/dev/null 2>&1; then
	echo "Not a git repo, skip commit"
	exit 0
fi

git add "$MAIN_GO"
git add update_version.sh 2>/dev/null || true
git commit -m "bump version to $NEW_VERSION"
echo "Committed: bump version to $NEW_VERSION"
