# Wiki Documentation

This directory contains the wiki documentation for go_logs v3.

## Multi-Language Support

The wiki is available in two languages:

- **English** (default): `Home.md`, `Getting-Started.md`, etc.
- **Espanol**: `Home-es.md`, `Getting-Started-es.md`, etc.

## Custom Sidebar

A `_Sidebar.md` file provides navigation for both languages with links to all documentation pages.

## Automatic Synchronization

The documentation in this directory is automatically synchronized to the [GitHub Wiki](https://github.com/drossan/go_logs/wiki) via a GitHub Actions workflow.

### How it works

1. When you push changes to `docs/wiki/**` on `main` or `develop` branches, the workflow triggers automatically
2. The workflow copies all markdown files (except README.md) from `docs/wiki/` to the repository's Wiki
3. Changes are committed and pushed to the Wiki repository

## File Structure

```
docs/wiki/
├── Home.md                    # Wiki home page (English)
├── Home-es.md                 # Wiki home page (Spanish)
├── Getting-Started.md         # Quick start guide (English)
├── Getting-Started-es.md      # Quick start guide (Spanish)
├── Installation.md            # Installation (English)
├── Installation-es.md         # Installation (Spanish)
├── API-Reference.md           # API docs (English)
├── API-Reference-es.md        # API docs (Spanish)
├── Structured-Logging.md      # Structured logging (English)
├── Structured-Logging-es.md   # Structured logging (Spanish)
├── Formatters.md              # Formatters (English)
├── Formatters-es.md           # Formatters (Spanish)
├── Context-and-Tracing.md     # Tracing (English)
├── Context-and-Tracing-es.md  # Tracing (Spanish)
├── Hooks.md                   # Hooks (English)
├── Hooks-es.md                # Hooks (Spanish)
├── File-Rotation.md           # File rotation (English)
├── File-Rotation-es.md        # File rotation (Spanish)
├── Configuration.md           # Configuration (English)
├── Configuration-es.md        # Configuration (Spanish)
├── Migration-v2-to-v3.md      # Migration (English)
├── Migration-v2-to-v3-es.md   # Migration (Spanish)
├── Examples.md                # Examples (English)
├── Examples-es.md             # Examples (Spanish)
├── Performance.md             # Performance (English)
├── Performance-es.md          # Performance (Spanish)
├── _Sidebar.md                # Custom sidebar navigation
└── README.md                  # This file (not synced)
```

## Contributing

To update the documentation:

1. Edit or create a markdown file in this directory
2. For multi-language support, create both English and Spanish versions
3. Ensure proper formatting and links
4. Commit and push to `main` or `develop`
5. The wiki will be updated automatically

## Notes

- Only `.md` files are synced to the Wiki
- `README.md` and `_Sidebar.md` are handled specially
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
