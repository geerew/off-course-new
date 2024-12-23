import { safeParse } from 'valibot';
import { UserSchema, type User } from './models';

class Auth {
	#user = $state<User | null>(null);
	#isAdmin = $state<boolean>(false);
	#error = $state<string | null>(null);

	constructor() {}

	async me(): Promise<void> {
		const response = await fetch('/api/auth/me');
		if (response.ok) {
			console.log('response', response);
			const data = await response.json();
			const result = safeParse(UserSchema, data);

			if (!result.success) {
				this.#error = 'Invalid response from the server';
				return;
			}

			this.#user = result.output;
			this.#isAdmin = result.output.role === 'admin';
			this.#error = null;
		} else {
			const data = await response.json();
			this.#error = `Failed to fetch user: ${data.message}`;
		}
	}

	async logout(): Promise<void> {
		const response = await fetch('/api/auth/logout');

		if (response.ok) {
			this.#error = null;
			window.location.href = '/auth/login';
		} else {
			const data = await response.json();
			this.#error = `Failed to logout: ${data.message}`;
		}
	}

	set user(value: User | null) {
		this.#user = value;
	}

	get user() {
		return this.#user;
	}

	get isAdmin() {
		return this.#isAdmin;
	}

	get error() {
		return this.#error;
	}
}

export const auth = new Auth();
