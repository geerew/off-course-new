import { safeParse } from 'valibot';
import { UserSchema, type User } from './models';

class Auth {
	#user = $state<User | null>(null);
	#userLetter = $state<string | null>(null);
	#isAdmin = $state<boolean>(false);
	#error = $state<string | null>(null);

	constructor() {}

	// Get information about the current user
	async me(): Promise<void> {
		const response = await fetch('/api/auth/me');
		if (response.ok) {
			const data = await response.json();
			const result = safeParse(UserSchema, data);

			if (!result.success) {
				this.#error = 'Invalid response from the server';
				return;
			}

			this.#user = result.output;
			this.#userLetter = this.#user.username.charAt(0).toUpperCase();
			this.#isAdmin = result.output.role === 'admin';
			this.#error = null;
		} else {
			// When the user is not authenticated, the server will return a 403 status code, so
			// we redirect to the login page
			if (response.status === 403) {
				window.location.href = '/auth/login';
			}

			const data = await response.json();
			this.#error = data.message;
		}
	}

	// Logout the current user. This will remove the session cookie and redirect to the login page
	async logout(): Promise<void> {
		const response = await fetch('/api/auth/logout', { method: 'POST' });

		if (response.ok) {
			this.empty();
			window.location.href = '/auth/login';
		} else {
			const data = await response.json();
			this.#error = data.message;
		}
	}

	// Clear the user information and redirect to the login page
	empty(): void {
		this.#error = null;
		this.#user = null;
		this.#userLetter = null;
		this.#isAdmin = false;
	}

	get user() {
		return this.#user;
	}

	get userLetter() {
		return this.#userLetter;
	}

	get isAdmin() {
		return this.#isAdmin;
	}

	get error() {
		return this.#error;
	}
}

export const auth = new Auth();
