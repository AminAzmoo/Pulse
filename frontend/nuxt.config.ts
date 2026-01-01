export default defineNuxtConfig({
  css: ['~/assets/css/main.css'],
  app: {
    head: {
      titleTemplate: '%s | Pulse',
      title: 'Pulse',
      meta: [
        { name: 'description', content: 'Pulse network console' }
      ]
    }
  }
})
