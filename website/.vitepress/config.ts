import { defineConfig, type DefaultTheme } from 'vitepress'

const repo = 'https://github.com/drossan/go_logs'

// Sidebars mirror docs/wiki/_Sidebar.md, one per locale. Links are site routes (cleanUrls: no .md);
// the Spanish ones live under /es/.
const sidebarEn: DefaultTheme.SidebarItem[] = [
  {
    text: 'Getting Started',
    items: [
      { text: 'Home', link: '/' },
      { text: 'Getting Started', link: '/getting-started' },
      { text: 'Installation', link: '/installation' },
    ],
  },
  {
    text: 'Core Concepts',
    items: [
      { text: 'API Reference', link: '/api-reference' },
      { text: 'Structured Logging', link: '/structured-logging' },
      { text: 'Formatters', link: '/formatters' },
    ],
  },
  {
    text: 'Advanced',
    items: [
      { text: 'Context & Tracing', link: '/context-and-tracing' },
      { text: 'Hooks', link: '/hooks' },
      { text: 'File Rotation', link: '/file-rotation' },
      { text: 'Optional Modules', link: '/optional-modules' },
    ],
  },
  { text: 'Configuration', items: [{ text: 'Configuration', link: '/configuration' }] },
  { text: 'Migration', items: [{ text: 'v2 to v3', link: '/migration-v2-to-v3' }] },
  {
    text: 'Resources',
    items: [
      { text: 'Examples', link: '/examples' },
      { text: 'Performance', link: '/performance' },
      { text: 'Changelog', link: '/changelog' },
    ],
  },
]

const sidebarEs: DefaultTheme.SidebarItem[] = [
  {
    text: 'Inicio',
    items: [
      { text: 'Inicio', link: '/es/' },
      { text: 'Inicio rápido', link: '/es/getting-started' },
      { text: 'Instalación', link: '/es/installation' },
    ],
  },
  {
    text: 'Conceptos core',
    items: [
      { text: 'Referencia API', link: '/es/api-reference' },
      { text: 'Logging estructurado', link: '/es/structured-logging' },
      { text: 'Formateadores', link: '/es/formatters' },
    ],
  },
  {
    text: 'Avanzado',
    items: [
      { text: 'Contexto y tracing', link: '/es/context-and-tracing' },
      { text: 'Hooks', link: '/es/hooks' },
      { text: 'Rotación de archivos', link: '/es/file-rotation' },
      { text: 'Módulos opcionales', link: '/es/optional-modules' },
    ],
  },
  { text: 'Configuración', items: [{ text: 'Configuración', link: '/es/configuration' }] },
  { text: 'Migración', items: [{ text: 'v2 a v3', link: '/es/migration-v2-to-v3' }] },
  {
    text: 'Recursos',
    items: [
      { text: 'Ejemplos', link: '/es/examples' },
      { text: 'Rendimiento', link: '/es/performance' },
      { text: 'Changelog', link: '/es/changelog' },
    ],
  },
]

// Version dropdown in the nav, as in sdk.griddo.io: current major plus where to go for older ones.
// v1.x (the legacy global API) has no site of its own, so it points to its last tag on GitHub.
const versionNav = (current: string, legacy: string, changelog: string, migration: string, migrationText: string): DefaultTheme.NavItem => ({
  text: current,
  items: [
    { text: 'Changelog', link: changelog },
    { text: migrationText, link: migration },
    { text: legacy, link: `${repo}/tree/v1.2.5` },
  ],
})

export default defineConfig({
  // Served from GitHub Pages as a project site: https://drossan.github.io/go_logs/.
  // Every asset and internal link is prefixed with this base.
  base: '/go_logs/',
  title: 'go_logs',
  description: 'Structured logging for Go',
  lastUpdated: true,
  cleanUrls: true,
  // website/README.md documents the site for contributors; it is not a page.
  srcExclude: ['README.md'],
  // ignoreDeadLinks is deliberately left unset (default false): a broken internal link must fail
  // the build, so copying pages from the wiki cannot silently ship dead links.

  // English is the root locale (/), Spanish lives under /es/. Each locale carries its own
  // nav and sidebar; the theme renders the language switcher from this map.
  locales: {
    root: {
      label: 'English',
      lang: 'en',
      themeConfig: {
        nav: [
          { text: 'Guide', link: '/getting-started' },
          { text: 'API', link: '/api-reference' },
          { text: 'Examples', link: '/examples' },
          { text: 'Changelog', link: '/changelog' },
          versionNav('v3.x (current)', 'v1.x (legacy)', '/changelog', '/migration-v2-to-v3', 'Migrating from v2'),
        ],
        sidebar: sidebarEn,
      },
    },
    es: {
      label: 'Español',
      lang: 'es',
      link: '/es/',
      description: 'Logging estructurado para Go',
      themeConfig: {
        nav: [
          { text: 'Guía', link: '/es/getting-started' },
          { text: 'API', link: '/es/api-reference' },
          { text: 'Ejemplos', link: '/es/examples' },
          { text: 'Changelog', link: '/es/changelog' },
          versionNav('v3.x (actual)', 'v1.x (legacy)', '/es/changelog', '/es/migration-v2-to-v3', 'Migrar desde v2'),
        ],
        sidebar: sidebarEs,
        editLink: { pattern: `${repo}/edit/main/website/:path`, text: 'Editar esta página en GitHub' },
        // Default theme UI strings, translated for the Spanish locale.
        outline: { level: [2, 3], label: 'En esta página' },
        docFooter: { prev: 'Anterior', next: 'Siguiente' },
        lastUpdated: { text: 'Última actualización' },
        footer: {
          message: 'Publicado bajo licencia MIT.',
          copyright: 'Copyright © 2023-2026 Daniel Rosselló',
        },
        returnToTopLabel: 'Volver arriba',
        sidebarMenuLabel: 'Menú',
        darkModeSwitchLabel: 'Apariencia',
        langMenuLabel: 'Cambiar idioma',
      },
    },
  },

  head: [['link', { rel: 'icon', type: 'image/svg+xml', href: '/go_logs/logo.svg' }]],

  themeConfig: {
    logo: '/logo.svg',
    outline: { level: [2, 3] },
    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2023-2026 Daniel Rosselló',
    },
    search: {
      provider: 'local',
      options: {
        locales: {
          es: {
            translations: {
              button: { buttonText: 'Buscar', buttonAriaLabel: 'Buscar' },
              modal: {
                noResultsText: 'Sin resultados para',
                resetButtonTitle: 'Borrar búsqueda',
                footer: { selectText: 'seleccionar', navigateText: 'navegar', closeText: 'cerrar' },
              },
            },
          },
        },
      },
    },
    socialLinks: [{ icon: 'github', link: repo }],
    editLink: { pattern: `${repo}/edit/main/website/:path`, text: 'Edit this page on GitHub' },
  },
})
