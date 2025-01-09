<script lang="ts">
	import { Logo } from '$lib/components';
	import { RightArrow } from '$lib/components/icons';
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
		const response = await fetch('/api/auth/register', {
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
			<div class="text-foreground-alt-1 text-center text-lg">Create an account</div>
		</div>

		<form onsubmit={handleSubmit} class="flex flex-col gap-5">
			<input
				name="username"
				type="text"
				bind:value={username}
				class={cn(
					'bg-background-alt-2 placeholder:text-foreground-alt-2 w-full rounded-md border border-transparent p-2.5 ring-0 duration-250 ease-in-out placeholder:text-sm placeholder:tracking-wide focus:brightness-110 focus:outline-none',
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
						'bg-background-alt-2 placeholder:text-foreground-alt-2 w-full rounded-md border border-transparent p-2.5 pe-10 ring-0 duration-250 ease-in-out placeholder:text-sm placeholder:tracking-wide focus:brightness-110 focus:outline-none',
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
					<svg
						id="eye-closed"
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="text-foreground-alt-2 visible size-5"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M3.98 8.223A10.477 10.477 0 0 0 1.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.451 10.451 0 0 1 12 4.5c4.756 0 8.773 3.162 10.065 7.498a10.522 10.522 0 0 1-4.293 5.774M6.228 6.228 3 3m3.228 3.228 3.65 3.65m7.894 7.894L21 21m-3.228-3.228-3.65-3.65m0 0a3 3 0 1 0-4.243-4.243m4.242 4.242L9.88 9.88"
						/>
					</svg>

					<svg
						id="eye-open"
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="text-foreground-alt-2 hidden size-5"
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
				class="bg-background-primary text-foreground-alt-3 disabled:bg-background-primary-alt-1 group mt-2 flex w-full cursor-pointer flex-row place-content-center items-center gap-2 rounded-md p-2.5 disabled:cursor-not-allowed"
				disabled={!usernameValid || !password || posting}
			>
				<span>Create account</span>
				<RightArrow
					class="relative left-0 size-5 transition-all duration-200 ease-in-out group-enabled:group-hover:left-1.5"
				/>
			</button>
		</form>

		<div class="text-foreground-alt-2 text-center">
			Already have an account?
			<a href="/auth/login/" class="hover:text-background-primary font-semibold duration-200">
				Login
			</a>
		</div>
	</div>
</div>
