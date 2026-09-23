/** @type {import('tailwindcss').Config} */
const frontendConfig = require('./frontend/tailwind.config.js');

module.exports = {
  ...frontendConfig,
  content: [
    './frontend/components/**/*.{js,vue,ts}',
    './frontend/layouts/**/*.vue',
    './frontend/pages/**/*.vue',
    './frontend/plugins/**/*.{js,ts}',
    './frontend/app.vue',
  ],
};
