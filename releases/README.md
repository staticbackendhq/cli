# Release Notes

This directory contains release notes for each version.

## Usage

Create a markdown file named after your release tag:

```
releases/v1.6.0.md
releases/v1.6.1.md
releases/v2.0.0-beta.1.md
```

The filename must **exactly match** the git tag name.

## Example

When you create tag `v1.6.1` and push it, the GitHub Action will:
1. Look for `releases/v1.6.1.md`
2. If found, use it as the release body
3. If not found, use default release notes

## Template

```markdown
## What's New in vX.Y.Z

### 🚀 Features
- New feature description

### 🔧 Improvements
- Improvement description

### 🐛 Bug Fixes
- Bug fix description

### 📦 Installation

\`\`\`bash
npm install -g @staticbackend/cli
\`\`\`
```

## Tips

- Write release notes **before** creating the tag
- Use emojis to make sections stand out
- Include code examples for new features
- Link to relevant issues/PRs
