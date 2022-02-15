module.exports = {
    style: {
        postcss: {
            plugins: [
                require('tailwindcss'),
                require('autoprefixer'),
            ],
        },
    },
    devServer: {
        allowedHosts: ['localhost','.gitpod.io','.ws.workstation.laszlo.cloud'],
    },
}
