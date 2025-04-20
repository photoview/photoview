/* global __dirname */

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
    'plugin:react-hooks/recommended',
    'plugin:@typescript-eslint/eslint-recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:@typescript-eslint/recommended-requiring-type-checking',
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
    tsconfigRootDir: __dirname,
    project: ['./tsconfig.json'],
    ecmaFeatures: {
      jsx: true,
    },
    ecmaVersion: 2018,
    sourceType: 'module',
  },
  plugins: [
    'react',
    'react-hooks',
    '@typescript-eslint'
  ],
  rules: {
    'no-unused-vars': 'off',
    'react/display-name': 'off',
    'react/react-in-jsx-scope': 'off',
    'react/prop-types': 'off', // TypeScript handles type checking
    'react/jsx-uses-react': 'off', // Not needed with new JSX transform
    'react-hooks/rules-of-hooks': 'error',
    'react-hooks/exhaustive-deps': 'warn',
    '@typescript-eslint/no-unused-vars': 'warn',
    '@typescript-eslint/no-var-requires': 'off',
    '@typescript-eslint/explicit-module-boundary-types': 'off',
    '@typescript-eslint/no-non-null-assertion': 'off',
    '@typescript-eslint/no-explicit-any': 'warn',
    '@typescript-eslint/ban-ts-comment': 'warn',
  },
  settings: {
    react: {
      version: 'detect',
      runtime: 'automatic'
    },
  },
  overrides: [
    {
      files: ['**/*.js', '**/*.jsx', '**/*.ts', '**/*.tsx'],
      rules: {
        '@typescript-eslint/explicit-module-boundary-types': 'off',
        '@typescript-eslint/no-floating-promises': 'off',
        '@typescript-eslint/no-misused-promises': 'off',
      },
    },
  ],
}
