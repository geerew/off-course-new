import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

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

export function capitalizeFirstLetter(str: string) {
	return String(str).charAt(0).toUpperCase() + String(str).slice(1);
}
