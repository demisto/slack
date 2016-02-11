var path = require('path');
var webpack = require('webpack');
var ExtractTextPlugin = require('extract-text-webpack-plugin');
var HtmlWebpackPlugin = require('html-webpack-plugin');
var LiveReloadPlugin = require('webpack-livereload-plugin');

var assets = 'assets';

module.exports = {
  devtool: 'source-map',
  entry: {
    app: './src/main'
  },
  output: {
    path: path.join(__dirname, 'dist'),
    filename: '[name].js',
    chunkFilename: '[id].js',
    publicPath: '/'
  },
  plugins: [
    new webpack.HotModuleReplacementPlugin(),
    new LiveReloadPlugin({port: 35729, appendScriptTag: true}),
    new webpack.NoErrorsPlugin(),
    new ExtractTextPlugin(assets + '/[name].css', { allChunks: true }),
    new webpack.DefinePlugin({
      'process.env': {
        'NODE_ENV': JSON.stringify('development')
      },
      __DEV_TOOLS__: true
    }),
    new HtmlWebpackPlugin({
      title: 'Slack Frame',
      filename: 'index.html',
      template: 'index.template.html',
      favicon: path.join(__dirname, assets + '/images/favicon.ico')
    }),
    new webpack.ProvidePlugin({
      'fetch': 'imports?this=>global!exports?global.fetch!whatwg-fetch'
    })
  ],
  module: {
    loaders: [
      {
        test: /\.css$/,
        loader: ExtractTextPlugin.extract('style', 'css!cssnext')
      },
      {
        test: /\.less$/,
        loader: ExtractTextPlugin.extract('style', 'css!cssnext!less')
      },
      {
        test: /\.js$/,
        loaders: ['babel'],
        include: path.join(__dirname, 'src')
      },
      {
        test: /\.(jpe?g|png|gif|svg)$/i,
        loaders: [
          'file?digest=hex&name=' + assets + '/[hash].[ext]',
          'image-webpack?bypassOnDebug&optimizationLevel=7&interlaced=false'
        ]
      },
      {
        test: /\.(ttf|eot|svg|woff|woff2)(\?v=[0-9]\.[0-9]\.[0-9])?$/,
        loader: 'file?name=' + assets + '/[hash].[ext]'
      },
      {
        test: /\.(wav|mp3)$/i,
        loader: 'file?name=' + assets + '/[hash].[ext]'
      },
      {
        test: /\.json$/,
        loader: 'json-loader'
      }
    ]
  },
  cssnext: {
    browsers: 'last 2 versions'
  }
};
