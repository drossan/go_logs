# go_logs documentation site

Source of <https://drossan.github.io/go_logs/>, built with [VitePress](https://vitepress.dev/).

**This site is the canonical source of the user documentation from v3.1.0 on.** `docs/wiki/` is still synced to the GitHub wiki, but it is not edited first: change the page here and then carry the change over to the wiki.

## Development

Requires Node 22 and pnpm.

```bash
cd website
pnpm install
pnpm dev       # http://localhost:5173/go_logs/
pnpm build     # output in .vitepress/dist
pnpm preview   # serve the build (restart it after rebuilding)
```

From the repository root: `pnpm --dir website build`.

The build fails on dead internal links (VitePress default; `ignoreDeadLinks` is deliberately not set).

`pnpm-workspace.yaml` approves the `esbuild` build script (`allowBuilds`); without it pnpm 11 exits with an error on `pnpm install`.

## Structure

```
website/
├── .vitepress/
│   ├── config.ts        # base, locales (en at /, es at /es/), nav, sidebar, search, footer
│   └── theme/           # default theme + brand colors (custom.css)
├── public/logo.svg
├── index.md             # home (layout: home)
├── changelog.md         # release notes
├── <page>.md            # English pages
└── es/                  # Spanish pages, same file names
```

## Adding or changing a page

- Put the English page in `website/<page>.md` and the Spanish one in `website/es/<page>.md` (kebab-case, lowercase).
- Add it to the sidebar of both locales in `.vitepress/config.ts`.
- Link to pages with site routes (`/formatters`, `/es/formatters`), not file names.
- Import paths are always `github.com/drossan/go_logs/v3` (Slack: `github.com/drossan/go_logs/slack/v3`). Check every API you cite against the code.
- Add each release to `changelog.md` and `es/changelog.md`.
