const colors = require('tailwindcss/colors');

module.exports = {
  content: [
    './src/**/*.{js,jsx,ts,tsx}',
    './public/index.html',
    './src/**/*.css',
    './node_modules/helm-react-ui/src/**/*.js'
  ],
  theme: {
    extend: {
      colors: {
        green : colors.emerald,
        yellow: colors.amber,
        grey: colors.slate,
      },
    },
  },
  plugins: [
    require('@tailwindcss/typography'),
    require('@tailwindcss/forms')
  ],
}
