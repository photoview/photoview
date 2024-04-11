module.exports = {
  defaultValue: function (locale, namespace, key, value) {
    if (locale != 'en'){
      return '';
    }
    return value || key;
  },
  sort: true,
  locales: [
    'da',
    'de',
    'en',
    'es',
    'fr',
    'it',
    'pl',
    'pt',
    'ru',
    'sv',
    'zh-CN',
    'zh-HK',
  ],
  input: 'src/**/*.{js,ts,jsx,tsx}',
  output: 'src/extractedTranslations/$LOCALE/$NAMESPACE.json',
}
