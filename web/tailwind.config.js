const colors = require('tailwindcss/colors');
const defaultTheme = require('tailwindcss/defaultTheme')

module.exports = {
  darkMode: 'selector',
  content: [
    './src/**/*.{js,jsx,ts,tsx}',
    './public/index.html',
    './src/**/*.css',
    './node_modules/helm-react-ui/src/**/*.js'
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['geist', ...defaultTheme.fontFamily.sans],
      },
    },
  },
  plugins: [
    require('@tailwindcss/typography'),
    require('@tailwindcss/forms')
  ],
}
