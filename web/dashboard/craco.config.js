module.exports = {
    style: {
        postcss: {
            plugins: {
                'postcss-import': {},
                'tailwindcss/nesting': {},
                tailwindcss: {},
                autoprefixer: {},
            },   
        },
    },
    devServer: {
        allowedHosts: ['localhost','.gitpod.io','.ws.workstation.laszlo.cloud'],
    },
}
