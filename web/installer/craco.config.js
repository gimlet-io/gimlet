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
        allowedHosts: ['localhost', '.gitpod.io', '.ws.workstation.laszlo.cloud'],
    },
    webpack: {
        configure: {
            output: {
                filename: '[name].js',
                chunkFilename: "1.chunk.js"
            },
            optimization: {
                runtimeChunk: false,
                splitChunks: {
                    // chunks: 'all',
                    cacheGroups: {
                        default: false,
                        // vendor chunk
                    },
                },
            }
        },
    },
    plugins: [
        {
            plugin: {
                overrideWebpackConfig: ({ webpackConfig }) => {
                    // find the plugin
                    let mcep;
                    webpackConfig.plugins.some(p => {
                        if (p.constructor.name === 'MiniCssExtractPlugin') {
                            mcep = p;
                            return true;
                        }
                    });
                    if (mcep) {
                        // change settings
                        mcep.options.filename = '[name].css';
                    }
                    return webpackConfig;
                },
            },
            options: {}
        },
    ],
}
