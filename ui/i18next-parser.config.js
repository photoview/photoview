module.exports = {
  skipDefaultValues: locale => locale != 'en',
  sort: true,
  locales: ['da', 'de', 'en', 'es', 'fr', 'it', 'pl', 'ru', 'sv'],
  input: 'src/**/*.{js,ts,jsx,tsx}',
  output: 'src/extractedTranslations/$LOCALE/$NAMESPACE.json',
}
