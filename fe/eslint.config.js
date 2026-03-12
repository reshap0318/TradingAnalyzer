import js from '@eslint/js'
import pluginVue from 'eslint-plugin-vue'
import globals from 'globals'
import configPrettier from 'eslint-config-prettier'
import pluginPrettier from 'eslint-plugin-prettier'
import tsParser from '@typescript-eslint/parser'
import pluginTs from '@typescript-eslint/eslint-plugin'

export default [
  {
    name: 'app/files-to-lint',
    files: ['**/*.{js,mjs,jsx,vue,ts,tsx}'],
  },

  {
    name: 'app/files-to-ignore',
    ignores: ['**/dist/**', '**/dist-ssr/**', '**/coverage/**', '**/node_modules/**'],
  },

  js.configs.recommended,
  ...pluginVue.configs['flat/recommended'],

  {
    files: ['**/*.{ts,tsx,vue}'],
    languageOptions: {
      parser: tsParser,
      parserOptions: {
        ecmaVersion: 'latest',
        sourceType: 'module',
      },
    },
    plugins: {
      '@typescript-eslint': pluginTs,
    },
    rules: {
      ...pluginTs.configs.recommended?.rules,
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      '@typescript-eslint/no-explicit-any': 'warn',
    },
  },

  {
    languageOptions: {
      globals: {
        ...globals.browser,
      },
    },

    rules: {
      'no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      'vue/multi-word-component-names': 'off',
      'vue/no-unused-vars': 'error',
      'prettier/prettier': 'error',
    },

    plugins: {
      prettier: pluginPrettier,
    },
  },

  configPrettier,
]
