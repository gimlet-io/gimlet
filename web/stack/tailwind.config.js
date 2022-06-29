const defaultTheme = require('tailwindcss/defaultTheme')

module.exports = {
  future: {
    removeDeprecatedGapUtilities: true,
    purgeLayersByDefault: true,
  },
    // enabled: false,
    // That will make sure all "reset" rules are kept no matter what
    layers: ['components', 'utilities'], // https://github.com/tailwindlabs/tailwindcss-forms/issues/43#issuecomment-791465128
    content: [
      './public/**/*.html',
      './src/**/*.js',
      './src/**/*.css',
      './node_modules/helm-react-ui/src/**/*.js',
      './node_modules/helm-react-ui/*.js',
    ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter var', ...defaultTheme.fontFamily.sans],
      },
    },
  },
  plugins: [
    require('@tailwindcss/typography'),
    require('@tailwindcss/forms')
  ]
}
