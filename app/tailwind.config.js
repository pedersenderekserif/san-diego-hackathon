/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{vue,js,ts,jsx,tsx}'
  ],
  theme: {
    extend: {
      colors: {
        base: '#041739',
        cream: '#F6F4F2',
        brand: {
          50: '#EEF7FF',
          100: '#D8EDFF',
          200: '#BADFFF',
          300: '#8BCEFF',
          400: '#55B1FF',
          500: '#2E8FFF',
          600: '#0F6AF9',
          700: '#1058E5',
          800: '#1446B9',
          900: '#0B1E76',
          950: '#071246',
        },
        purple: {
          50: '#FAF5FE',
          100: '#F5EAFD',
          200: '#EBD3FB',
          300: '#DEB1F6',
          400: '#C780EA',
          500: '#B353E2',
          600: '#A129DB',
          700: '#8C20C0',
          800: '#6C1995',
          900: '#4C1269',
          950: '#2C0A3D',
        },
        green: {
          50: '#F9FFE6',
          100: '#F1FDCA',
          200: '#E2FC9A',
          300: '#C5F54A',
          400: '#B4EB30',
          500: '#95D111',
          600: '#74A709',
          700: '#577F0C',
          800: '#466410',
          900: '#3C5512',
          950: '#1E2F04',
        },
        cyan: {
          50: '#E5FAFF',
          100: '#CCF5FF',
          200: '#99EBFF',
          300: '#80E5FF',
          400: '#4DDBFF',
          500: '#1FD2FF',
          600: '#00BCEB',
          700: '#00A3CC',
          800: '#0087A8',
          900: '#005E75',
          950: '#003542',
        },
      }
    }
  },
  plugins: []
}
