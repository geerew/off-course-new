<script lang="ts">
	import { RightArrow } from '$lib/components/icons';
	import { Button } from 'bits-ui';
	import { toast } from 'svelte-sonner';
	import { Input, InputPassword } from '.';

	let { endpoint }: { endpoint: string } = $props();

	let username = $state('');
	let password = $state('');
	let posting = $state(false);

	async function submitForm(event: Event) {
		event.preventDefault();
		posting = true;

		const response = await fetch(endpoint, {
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
</script>

<form onsubmit={submitForm} class="flex flex-col gap-5">
	<Input bind:value={username} name="username" type="text" placeholder="Username" />
	<InputPassword bind:value={password} placeholder="password" />
	<Button.Root
		type="submit"
		class="bg-background-primary text-foreground-alt-3 disabled:bg-background-primary-alt-1 group mt-2 flex w-full cursor-pointer flex-row place-content-center items-center gap-2 rounded-md p-2.5 disabled:cursor-not-allowed"
		disabled={!username || !password || posting}
	>
		<span>Login</span>
		<RightArrow
			class="relative left-0 size-5 transition-all duration-200 ease-in-out group-enabled:group-hover:left-1.5"
		/>
	</Button.Root>
</form>
