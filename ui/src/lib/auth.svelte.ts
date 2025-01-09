import { toast } from 'svelte-sonner';
import { safeParse } from 'valibot';
import { UserSchema, type User } from './models';

class Auth {
	#user = $state<User | null>(null);
	#userLetter = $state<string | null>(null);
	#isAdmin = $state<boolean>(false);
	#error = $state<string | null>(null);

	constructor() {}

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

	async logout(): Promise<void> {
		const response = await fetch('/api/auth/logout', { method: 'POST' });

		if (response.ok) {
			this.#error = null;
			this.#user = null;
			this.#userLetter = null;
			this.#isAdmin = false;
			window.location.href = '/auth/login';
		} else {
			const data = await response.json();
			this.#error = data.message;
		}
	}

	async delete(): Promise<void> {
		const response = await fetch('/api/auth/me', { method: 'DELETE' });

		if (response.ok) {
			this.#error = null;
			this.#user = null;
			this.#userLetter = null;
			this.#isAdmin = false;
			window.location.href = '/auth/login';
		} else {
			const data = await response.json();
			toast.error(`${data.message}`);
		}
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
