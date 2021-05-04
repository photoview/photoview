module.exports = {
  root: true,
  parser: '@typescript-eslint/parser',
  env: {
    browser: true,
    es6: true,
  },
  ignorePatterns: ['node_modules', 'dist'],
  extends: [
    'eslint:recommended',
    'plugin:react/recommended',
    'plugin:@typescript-eslint/eslint-recommended',
    'plugin:@typescript-eslint/recommended',
    'prettier',
  ],
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
  plugins: ['react', 'react-hooks', '@typescript-eslint'],
  rules: {
    'no-unused-vars': 'off',
    'react/display-name': 'off',
    '@typescript-eslint/no-unused-vars': 'warn',
    '@typescript-eslint/no-var-requires': 'off',
    '@typescript-eslint/explicit-module-boundary-types': 'off',
    '@typescript-eslint/no-non-null-assertion': 'off',
  },
  settings: {
    react: {
      version: 'detect',
    },
  },
  // parser: 'babel-eslint',
  overrides: [
    Object.assign(require('eslint-plugin-jest').configs.recommended, {
      files: ['**/*.test.js', '**/*.test.ts', '**/*.test.tsx'],
      env: { jest: true },
      plugins: ['jest', 'jest-dom'],
      rules: Object.assign(
        require('eslint-plugin-jest').configs.recommended.rules,
        {
          'no-import-assign': 'off',
          'react/prop-types': 'off',
          'jest/valid-title': 'off',
        }
      ),
    }),
    {
      files: ['**/*.js'],
      rules: {
        '@typescript-eslint/explicit-module-boundary-types': 'off',
      },
    },
  ],
}
