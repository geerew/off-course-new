/* eslint @typescript-eslint/no-var-requires: "off" */

const path = require('path');

module.exports = {
	plugins: {
		'tailwindcss/nesting': {},
		tailwindcss: {
			config: path.join(__dirname, './tailwind.config.js')
		},
		autoprefixer: {}
	}
};
