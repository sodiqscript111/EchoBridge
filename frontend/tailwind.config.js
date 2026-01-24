/** @type {import('tailwindcss').Config} */
export default {
    content: ['./src/**/*.{html,js,svelte,ts}'],
    theme: {
        extend: {
            colors: {
                primary: {
                    DEFAULT: '#FFFFFF', // White
                    500: '#FFFFFF',
                },
                surface: 'rgba(255, 255, 255, 0.05)',
                zinc: {
                    950: '#09090b', // Deep Zinc
                }
            },
            fontFamily: {
                sans: ['Inter', 'system-ui', 'sans-serif'],
            }
        },
    },
    plugins: [],
}
