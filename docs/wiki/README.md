# Wiki Documentation

This directory contains the wiki documentation for go_logs v3.

## Automatic Synchronization

The documentation in this directory is automatically synchronized to the [GitHub Wiki](https://github.com/drossan/go_logs/wiki) via a GitHub Actions workflow.

### How it works

1. When you push changes to `docs/wiki/**` on `main` or `develop` branches, the workflow triggers automatically
2. The workflow copies all markdown files from `docs/wiki/` to the repository's Wiki
3. Changes are committed and pushed to the Wiki repository

### Manual Sync

You can also trigger a manual sync from the Actions tab:

1. Go to **Actions** > **Sync Wiki Documentation**
2. Click **Run workflow**
3. Optionally provide a custom commit message
4. Click **Run workflow**

## File Structure

```
docs/wiki/
├── Home.md                    # Wiki home page (required)
├── Getting-Started.md         # Quick start guide
├── Installation.md            # Installation instructions
├── API-Reference.md           # Complete API documentation
├── Structured-Logging.md      # Structured logging guide
├── Formatters.md              # Formatter configuration
├── Context-and-Tracing.md     # Distributed tracing
├── Hooks.md                   # Hook system documentation
├── File-Rotation.md           # File rotation guide
├── Configuration.md           # Configuration options
├── Migration-v2-to-v3.md      # Migration guide
├── Examples.md                # Practical examples
├── Performance.md             # Benchmarks and optimization
└── README.md                  # This file (not synced)
```

## Contributing

To update the documentation:

1. Edit or create a markdown file in this directory
2. Ensure proper formatting and links
3. Commit and push to `main` or `develop`
4. The wiki will be updated automatically

## Notes

- Only `.md` files are synced to the Wiki
- The `README.md` file is excluded from sync
- Wiki pages use the filename (without .md) as the page title
- Use standard markdown formatting
- Internal wiki links should use the format: `[Link Text](Page-Name)`

## Local Preview

To preview the wiki locally, you can use any Markdown viewer or:

```bash
# Using grip (GitHub-flavored markdown)
grip docs/wiki/Home.md

# Or open in your preferred markdown viewer
```
