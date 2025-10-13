#!/bin/bash
# Automated GitHub documentation workflow for lowkey
# Usage: ./gh-doc-workflow.sh <issue-number> "<title>" "<body>" "<file-changes>"

set -e

REPO="onedusk/lowkey"
ISSUE_NUM=$1
TITLE=$2
BODY=$3
FILE_CHANGES=$4

if [ -z "$ISSUE_NUM" ] || [ -z "$TITLE" ]; then
    echo "Usage: $0 <issue-number> \"<title>\" \"<body>\" \"<file-changes>\""
    exit 1
fi

# Set profile
git s-p onedusk

# Create issue
echo "Creating issue #${ISSUE_NUM}..."
ISSUE_URL=$(gh issue create --repo $REPO --title "$TITLE" --body "$BODY")
echo "Created: $ISSUE_URL"

# Extract issue number from URL
ISSUE_ID=$(echo $ISSUE_URL | grep -oE '[0-9]+$')

# Create branch
BRANCH="docs/${ISSUE_ID}-$(echo $TITLE | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | cut -c1-40)"
echo "Creating branch: $BRANCH"
git checkout -b "$BRANCH"

# Apply file changes (passed as argument)
echo "$FILE_CHANGES"

# Commit
echo "Committing changes..."
git add .
git commit -m "docs: ${TITLE}

${BODY}

Closes #${ISSUE_ID}"

# Push
echo "Pushing branch..."
git push -u origin "$BRANCH"

# Create PR
echo "Creating PR..."
PR_URL=$(gh pr create --title "docs: ${TITLE}" --body "Closes #${ISSUE_ID}" --repo $REPO)
echo "Created: $PR_URL"

# Extract PR number
PR_ID=$(echo $PR_URL | grep -oE '[0-9]+$')

# Merge PR
echo "Merging PR #${PR_ID}..."
gh pr merge $PR_ID --squash --delete-branch --repo $REPO

# Return to main
echo "Returning to main..."
git checkout main
git pull origin main

echo "✓ Complete! Issue #${ISSUE_ID} closed, PR #${PR_ID} merged"
