export default defineNuxtConfig({
  devtools: { enabled: true },
  css: ['~/assets/main.css'],
  app: {
    head: {
      title: 'Pulse',
      meta: [
        { name: 'description', content: 'Pulse network management console.' },
      ],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
      ],
    },
  },
  runtimeConfig: {
    public: {
      apiBase: '',
    },
  },
  typescript: {
    strict: true,
  },
});
