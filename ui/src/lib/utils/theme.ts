/**
 * Retrieves the current theme from local storage or falls back to the user's preference.
 *
 * The function first checks if the theme is set in the local storage. If not found, it checks
 * the user's system preference for dark or light theme.
 *
 * @returns {string} The current theme, either 'dark' or 'light'.
 */
export function currentTheme() {
	// Get the current theme
	const currentTheme = localStorage.getItem('theme') || 'light';

	// If there is no current theme set it based up the user's preference
	if (currentTheme) return currentTheme;
	return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

/**
 * Reflects the provided theme preference on the document.
 *
 * If the provided theme is 'light', it removes the 'dark' class from the document's root. Otherwise,
 * it adds the 'dark' class. Additionally, it sets the 'data-theme' attribute on the document's first
 * child element to the provided theme. Lastly, if there's an element with the ID 'theme-btn', it
 * sets its 'aria-label' to indicate the option to switch to the other theme.
 *
 * @param {string} theme - The theme preference, either 'dark' or 'light'.
 */
export function reflectPreference(theme: string) {
	if (theme === 'light') {
		document.documentElement.classList.remove('dark');
	} else {
		document.documentElement.classList.add('dark');
	}

	document.firstElementChild?.setAttribute('data-theme', theme);

	document
		.querySelector('#theme-btn')
		?.setAttribute('aria-label', `Switch to ${theme === 'light' ? 'dark' : 'light'} mode`);
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

/**
 * Saves the provided theme preference to local storage and then reflects it on the document.
 *
 * This function first saves the theme to local storage under the 'theme' key and then calls
 * `reflectPreference` to apply the theme to the document.
 *
 * @param {string} theme - The theme preference, either 'dark' or 'light'.
 */
export function setPreference(theme: string) {
	localStorage.setItem('theme', theme);
	reflectPreference(theme);
}
