import type { Asset, CourseChapters } from '$lib/types/models';
import { clsx, type ClassValue } from 'clsx';
import type { SortKey } from 'svelte-headless-table/lib/plugins/addSortBy';
import { twMerge } from 'tailwind-merge';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

/**
 * Combines class names using `clsx` and then merges them using `twMerge`.
 *
 * @param {...ClassValue[]} inputs - The list of class values to be joined and merged.
 * @returns {string} A string of merged class values.
 *
 */
export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(...inputs));
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

/**
 * Checks if the current environment is a browser environment.
 * @returns {boolean} True if the current environment is a browser environment, false otherwise.
 */
export const isBrowser = typeof document !== 'undefined';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

/**
 * Converts a camelCase or PascalCase string to snake_case.
 *
 * @param {string} str - The input string in camelCase or PascalCase.
 * @returns {string} - The converted string in snake_case.
 */
export function toSnakeCase(str: string): string {
	// Replace lowercase followed by uppercase, e.g., 'thisIs' -> 'this_Is'
	let result = str.replace(/([a-z])([A-Z])/g, '$1_$2');

	// Replace uppercase followed by uppercase then lowercase, e.g., 'AWord' -> 'A_Word'
	result = result.replace(/([A-Z])([A-Z][a-z])/g, '$1_$2');

	return result.toLowerCase();
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

/**
 * Converts an array of SortKey objects into a flattened order-by string.
 *
 * @param {SortKey[]} sortKeys - An array of SortKey objects to be converted.
 * @returns {string | undefined} - A flattened order-by string or undefined if the input array is
 * empty.
 */
export function flattenOrderBy(sortKeys: SortKey[]): string | undefined {
	// Return undefined if the array is empty
	if (sortKeys.length === 0) {
		return undefined;
	}

	// Convert the array of sort keys to an array of strings
	const orderStrings = sortKeys.map((sortKey) => `${toSnakeCase(sortKey.id)} ${sortKey.order}`);

	// Join the array of strings with a comma and return
	return orderStrings.join(', ');
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

/**
 * Constructs a chapter-based structure for the given course assets.
 *
 * @function
 * @param {Asset[]} courseAssets - An array of assets associated with a course.
 * @returns {CourseChapters} - An object with chapter names as keys and an array of associated
 * assets as values. Assets without a chapter are grouped under the '(no chapter)' key.
 */
export function buildChapterStructure(courseAssets: Asset[]): CourseChapters {
	const chapters: CourseChapters = {};

	// Loop through each asset and build the chapter structure
	for (const courseAsset of courseAssets) {
		const chapter = courseAsset.chapter || '(no chapter)';
		!chapters[chapter] ? (chapters[chapter] = [courseAsset]) : chapters[chapter]?.push(courseAsset);
	}

	return chapters;
}
