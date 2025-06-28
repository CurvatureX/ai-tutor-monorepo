import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

const config: Config = {
  title: 'CurvTech',
  tagline: 'Documentation for AI Tutor',
  favicon: 'img/favicon.ico',

  // Future flags, see https://docusaurus.io/docs/api/docusaurus-config#future
  future: {
    v4: true, // Improve compatibility with the upcoming Docusaurus v4
  },

  // Set the production url of your site here
  url: 'https://your-docusaurus-site.example.com',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/ai-tutor-monorepo/',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'CurvatureX', // GitHub 组织名
  projectName: 'ai-tutor-monorepo', // 仓库名

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: false, // Disable the default docs plugin
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      },
    ],
  ],

  plugins: [
    [
      '@docusaurus/plugin-content-docs',
      {
        id: 'tech_design',
        path: '../../docs/tech_design',
        routeBasePath: 'tech_design',
        sidebarPath: './sidebars.ts',
      },
    ],
    [
      '@docusaurus/plugin-content-docs',
      {
        id: 'api_documentation',
        path: '../../docs/api_documentation',
        routeBasePath: 'api_documentation',
        sidebarPath: './sidebars.ts',
      },
    ],
  ],

  themeConfig: {
    // Replace with your project's social card
    image: 'img/docusaurus-social-card.jpg',
    navbar: {
      title: 'CurvTech',
      logo: {
        alt: 'My Site Logo',
        src: 'img/logo.svg',
      },
      items: [
        {
          type: 'doc',
          docId: '架构设计',
          docsPluginId: 'tech_design',
          position: 'left',
          label: 'Tech Design',
        },
        {
          type: 'doc',
          docId: 'intro',
          docsPluginId: 'api_documentation',
          position: 'left',
          label: 'API Documentation',
        },
        {
          href: 'https://github.com/CurvatureX/ai-tutor-monorepo',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Docs',
          items: [
            {
              label: 'Tech Design',
              to: '/tech_design/架构设计',
            },
            {
              label: 'API Documentation',
              to: '/api_documentation/intro',
            },
          ],
        },
        {
          title: 'More',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/CurvatureX/ai-tutor-monorepo',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} , CurvTech Inc.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
    },
  },
  trailingSlash: true,
};

export default config;
