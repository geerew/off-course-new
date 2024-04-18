import type { Asset, CourseChapters } from '$lib/types/models';
import { clsx, type ClassValue } from 'clsx';
import type { SortKey } from 'svelte-headless-table/plugins';
import { cubicOut } from 'svelte/easing';
import type { TransitionConfig } from 'svelte/transition';
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
	return twMerge(clsx(inputs));
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

export const NO_CHAPTER = '(no chapter)';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

/**
 * Constructs a chapter-based structure for the given course assets.
 *
 * @function
 * @param {Asset[]} courseAssets - An array of assets associated with a course
 * @returns {CourseChapters} - An object with chapter names as keys and an array of associated
 * assets as values. Assets without a chapter are grouped under the `NO_CHAPTER` key
 */
export function buildChapterStructure(courseAssets: Asset[]): CourseChapters {
	const chapters: CourseChapters = {};

	// Loop through each asset and build the chapter structure
	for (const courseAsset of courseAssets) {
		const chapter = courseAsset.chapter || NO_CHAPTER;
		!chapters[chapter] ? (chapters[chapter] = [courseAsset]) : chapters[chapter]?.push(courseAsset);
	}

	return chapters;
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

/**
 *
 *
 */
type FlyAndScaleParams = {
	y?: number;
	x?: number;
	start?: number;
	duration?: number;
};

export const flyAndScale = (
	node: Element,
	params: FlyAndScaleParams = { y: -8, x: 0, start: 0.95, duration: 150 }
): TransitionConfig => {
	const style = getComputedStyle(node);
	const transform = style.transform === 'none' ? '' : style.transform;

	const scaleConversion = (valueA: number, scaleA: [number, number], scaleB: [number, number]) => {
		const [minA, maxA] = scaleA;
		const [minB, maxB] = scaleB;

		const percentage = (valueA - minA) / (maxA - minA);
		const valueB = percentage * (maxB - minB) + minB;

		return valueB;
	};

	const styleToString = (style: Record<string, number | string | undefined>): string => {
		return Object.keys(style).reduce((str, key) => {
			if (style[key] === undefined) return str;
			return str + `${key}:${style[key]};`;
		}, '');
	};

	return {
		duration: params.duration ?? 200,
		delay: 0,
		css: (t) => {
			const y = scaleConversion(t, [0, 1], [params.y ?? 5, 0]);
			const x = scaleConversion(t, [0, 1], [params.x ?? 0, 0]);
			const scale = scaleConversion(t, [0, 1], [params.start ?? 0.95, 1]);

			return styleToString({
				transform: `${transform} translate3d(${x}px, ${y}px, 0) scale(${scale})`,
				opacity: t
			});
		},
		easing: cubicOut
	};
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const Throttle = <R, A extends unknown[]>(
	fn: (...args: A) => R,
	delay: number
): [(...args: A) => R | undefined, () => void] => {
	let wait = false;
	let timeout: undefined | number;

	return [
		(...args: A) => {
			if (wait) return undefined;

			const val = fn(...args);

			wait = true;

			timeout = window.setTimeout(() => {
				wait = false;
			}, delay);

			return val;
		},

		() => {
			wait = false;
			clearTimeout(timeout);
		}
	];
};
