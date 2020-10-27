module.exports = {
  env: {
    browser: true,
    es6: true,
  },
  extends: ['eslint:recommended', 'plugin:react/recommended'],
  globals: {
    Atomics: 'readonly',
    SharedArrayBuffer: 'readonly',
    process: 'readonly',
    module: 'readonly',
    require: 'readonly',
  },
  parserOptions: {
    ecmaFeatures: {
      jsx: true,
    },
    ecmaVersion: 2018,
    sourceType: 'module',
  },
  plugins: ['react', 'react-hooks'],
  rules: {
    'no-unused-vars': 'warn',
    'react/display-name': 'off',
  },
  settings: {
    react: {
      version: 'detect',
    },
  },
  parser: 'babel-eslint',
  overrides: [
    Object.assign(require('eslint-plugin-jest').configs.recommended, {
      files: ['**/*.test.js'],
      env: { jest: true },
      plugins: ['jest', 'jest-dom'],
      rules: Object.assign(
        require('eslint-plugin-jest').configs.recommended.rules,
        {
          'no-import-assign': 'off',
          'react/prop-types': 'off',
        }
      ),
    }),
  ],
}
