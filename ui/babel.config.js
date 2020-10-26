module.exports = {
  presets: ['@babel/preset-env', '@babel/preset-react'],
  plugins: [
    'styled-components',
    '@babel/plugin-transform-runtime',
    '@babel/plugin-transform-modules-commonjs',
    'graphql-tag',
    // [
    //   'transform-semantic-ui-react-imports',
    //   {
    //     convertMemberImports: true,
    //     addCssImports: true,
    //   },
    // ],
  ],
}
