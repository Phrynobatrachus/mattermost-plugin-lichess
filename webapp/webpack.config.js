var path = require('path')

module.exports = {
    entry: [
        './src/index.jsx'
    ],
    resolve: {
        modules: [
            'src',
            'node_modules',
        ],
        extensions: ['*', '.js', '.jsx'],
    },
    module: {
        rules: [
            {
                exclude: /node_modules/,
                use: {
                    loader: 'babel-loader',
                    options: {
                        presets: ['@babel/preset-react',
                            [
                                '@babel/preset-env',
                                {
                                    modules: "commonjs",
                                    targets: {
                                        node: "current"
                                    }
                                }
                            ]
                        ]
                    }
                }
            }
        ]
    },
    output: {
        path: path.join(__dirname, '/dist'),
        publicPath: '/',
        filename: 'main.js'
    }
}