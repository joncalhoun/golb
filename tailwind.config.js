/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./posts/**/*.md", "./**/*.gohtml"],
  theme: {
    extend: {},
  },
  plugins: [
    require('@tailwindcss/typography'),
  ],
}

