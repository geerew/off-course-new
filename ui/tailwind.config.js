import { join } from 'path';
import { fontFamily } from 'tailwindcss/defaultTheme';

/** @type {import('tailwindcss').Config} */
export default {
	darkMode: 'class',
	content: [join(__dirname, './src/**/*.{html,js,svelte,ts}')],
	theme: {
		extend: {
			container: {
				center: true,
				screens: {
					'2xl': '1400px'
				}
			},
			fontFamily: {
				inter: ['Inter Variable', ...fontFamily.sans],
				'ubuntu-mono': ['Ubuntu Mono', 'monospace']
			},
			colors: {
				background: {
					DEFAULT: 'hsl(var(--background))',
					'alt-1': 'hsl(var(--background-alt-1))'
				},
				foreground: {
					DEFAULT: 'hsl(var(--foreground))',
					muted: 'hsl(var(--foreground-muted))'
				},
				muted: {
					DEFAULT: 'hsl(var(--muted))'
				},
				primary: {
					DEFAULT: 'hsl(var(--primary))',
					foreground: 'hsl(var(--primary-foreground))'
				},
				secondary: {
					DEFAULT: 'hsl(var(--secondary))',
					foreground: 'hsl(var(--secondary-foreground))'
				},
				success: {
					DEFAULT: 'hsl(var(--success))'
				},
				error: {
					DEFAULT: 'hsl(var(--error))'
				},
				border: {
					DEFAULT: 'hsl(var(--border))'
				},
				accent: {
					1: 'hsl(var(--accent-1))'
				}
			},
			boxShadow: {
				focus: '0 0 0 2px rgb(0 0 0 / 0.1), 0 0 0 1px rgb(0 0 0 / 0.1)'
			}
		}
	},
	plugins: [require('@tailwindcss/forms')]
};
