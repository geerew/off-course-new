<script lang="ts">
	import { Button } from 'bits-ui';
	import type { HTMLInputAttributes } from 'svelte/elements';
	import { Input } from '.';

	type Props = HTMLInputAttributes & { ref?: HTMLInputElement };

	let {
		value = $bindable(''),
		ref = $bindable(),
		class: containerClass = '',
		...restProps
	}: Props = $props();

	let passwordEl = $state<HTMLInputElement | undefined>(ref);
	let passwordEyeOpenEl = $state<SVGElement>();
	let passwordEyeClosedEl = $state<SVGElement>();

	function togglePasswordVisibility(
		passwordEl: HTMLInputElement | undefined,
		passwordEyeOpenEl: SVGElement | undefined,
		passwordEyeClosedEl: SVGElement | undefined
	) {
		if (!passwordEl) return;
		passwordEl.type = passwordEl.type === 'password' ? 'text' : 'password';

		if (!passwordEyeOpenEl || !passwordEyeClosedEl) return;
		passwordEyeOpenEl.classList.toggle('hidden');
		passwordEyeClosedEl.classList.toggle('hidden');
	}
</script>

<div class="relative w-full overflow-hidden rounded-md">
	<Input bind:ref bind:value name="password" type="password" class="pe-10" {...restProps} />

	<Button.Root
		type="button"
		class="absolute top-1/2 right-0 h-full w-8 -translate-y-1/2 hover:cursor-pointer"
		aria-label="Toggle password visibility"
		onclick={() => {
			togglePasswordVisibility(passwordEl, passwordEyeOpenEl, passwordEyeClosedEl);
		}}
	>
		<svg
			bind:this={passwordEyeClosedEl}
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
			bind:this={passwordEyeOpenEl}
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
	</Button.Root>
</div>
