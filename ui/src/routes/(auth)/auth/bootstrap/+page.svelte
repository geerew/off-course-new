<script lang="ts">
	import { Logo } from '$lib/components';
	import { cn } from '$lib/utils.js';
	import { toast } from 'svelte-sonner';

	// Username
	let username = $state('');
	let usernameValid = $state(true);

	// Password
	let password = $state('');
	let passwordValid = $state(true);

	let posting = $state(false);

	// Runs any time the username or password changes to clear any potential errors
	$effect(() => {
		if (username.length > 0) {
			usernameValid = true;
		}

		if (password.length > 0) {
			passwordValid = true;
		}
	});

	// Handles the form submission
	async function handleSubmit(event: Event) {
		posting = true;
		event.preventDefault();

		// Validate the username
		if (username.length === 0) {
			usernameValid = false;
		}

		// Validate the password
		if (password.length === 0) {
			passwordValid = false;
		}

		if (!usernameValid || !passwordValid) {
			return;
		}

		// Create the user
		const response = await fetch('/api/auth/bootstrap', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				username,
				password
			})
		});

		if (response.ok) {
			window.location.href = '/';
		} else {
			// TODO handle error with toast
			const data = await response.json();
			toast.error(data.message);
			posting = false;
		}
	}

	// Toggles the visibility of the password field
	function togglePasswordVisibility(node: HTMLElement) {
		node.addEventListener('click', () => {
			const passwordEl = node.previousElementSibling as HTMLInputElement;
			if (passwordEl === null) return;
			passwordEl.type = passwordEl.type === 'password' ? 'text' : 'password';

			const eyeOpen = node.querySelector('#eye-open');
			const eyeClosed = node.querySelector('#eye-closed');

			if (eyeOpen !== null && eyeClosed !== null) {
				eyeOpen.classList.toggle('hidden');
				eyeClosed.classList.toggle('hidden');
			}
		});
	}
</script>

<div class="flex h-full min-h-110 flex-col items-center justify-center gap-8">
	<Logo />

	<div class="flex min-w-sm flex-col gap-5">
		<div class="mb-2.5 flex flex-col gap-2">
			<div class="text-muted-foreground text-center text-lg">Create administrator account</div>
		</div>

		<form onsubmit={handleSubmit} class="flex flex-col gap-5">
			<input
				name="username"
				type="text"
				bind:value={username}
				class={cn(
					'bg-muted-background placeholder:text-muted-foreground w-full rounded-md border border-transparent p-2.5 ring-0 duration-250 ease-in-out placeholder:text-sm placeholder:tracking-wide focus:brightness-110 focus:outline-none',
					!usernameValid && 'border-red-500'
				)}
				placeholder="Username"
			/>

			<div class="0 relative w-full overflow-hidden rounded-md">
				<input
					name="password"
					type="password"
					bind:value={password}
					class={cn(
						'bg-muted-background placeholder:text-muted-foreground w-full rounded-md border border-transparent p-2.5 pe-10 ring-0 duration-250 ease-in-out placeholder:text-sm placeholder:tracking-wide focus:brightness-110 focus:outline-none',
						!passwordValid && 'border-red-500'
					)}
					placeholder="Password"
				/>

				<button
					type="button"
					class="absolute top-1/2 right-0 h-full w-8 -translate-y-1/2 hover:cursor-pointer"
					aria-label="Toggle password visibility"
					use:togglePasswordVisibility
				>
					<!--  Eye open -->
					<svg
						id="eye-closed"
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="text-muted-foreground visible size-5"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M3.98 8.223A10.477 10.477 0 0 0 1.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.451 10.451 0 0 1 12 4.5c4.756 0 8.773 3.162 10.065 7.498a10.522 10.522 0 0 1-4.293 5.774M6.228 6.228 3 3m3.228 3.228 3.65 3.65m7.894 7.894L21 21m-3.228-3.228-3.65-3.65m0 0a3 3 0 1 0-4.243-4.243m4.242 4.242L9.88 9.88"
						/>
					</svg>

					<!-- Eye closed -->
					<svg
						id="eye-open"
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="text-muted-foreground hidden size-5"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178Z"
						/>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z"
						/>
					</svg>
				</button>
			</div>

			<button
				type="submit"
				class="bg-primary-background disabled:text-muted-foreground mt-2 flex w-full cursor-pointer flex-row place-content-center items-center gap-2 rounded-md p-2.5 hover:brightness-110 disabled:cursor-not-allowed disabled:brightness-100"
				disabled={!usernameValid || !password || posting}
			>
				<span>Create account</span>

				<svg
					xmlns="http://www.w3.org/2000/svg"
					viewBox="0 0 20 20"
					fill="currentColor"
					class="size-5"
				>
					<path
						fill-rule="evenodd"
						d="M3 10a.75.75 0 0 1 .75-.75h10.638L10.23 5.29a.75.75 0 1 1 1.04-1.08l5.5 5.25a.75.75 0 0 1 0 1.08l-5.5 5.25a.75.75 0 1 1-1.04-1.08l4.158-3.96H3.75A.75.75 0 0 1 3 10Z"
						clip-rule="evenodd"
					/>
				</svg>
			</button>
		</form>
	</div>
</div>
