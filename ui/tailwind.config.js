module.exports = {
  mode: 'jit',
  purge: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      screens: {
        xs: '480px',
      },
      boxShadow: {
        separator: '0 0 4px 0 rgba(0, 0, 0, 0.1)',
      },
      colors: {
        green: {
          50: '#f0ffee',
          100: '#dcffd8',
          200: '#b6fdb3',
          300: '#7cf587',
          400: '#56e263',
          500: '#4abf3c',
          600: '#30a23e',
          700: '#168332',
          800: '#006624',
          900: '#00541d',
        },
        dark: {
          bg: '#24292e',
          bg2: '#30363e',
          text: '#fafafa',
          border: '#3b3b3b',
          border2: '#1E1E1E',
          input: {
            bg: '#383e46',
            border: '#4a515a',
            text: '#ccdbe4',
          },
        },
      },
    },
  },
  variants: {
    extend: {},
  },
  plugins: [],
}
