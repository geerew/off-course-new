<script lang="ts">
	import { Spinner } from '$lib/components';
	import { toast } from 'svelte-sonner';
	import { Input, InputPassword, SubmitButton } from '.';

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
	<SubmitButton disabled={!username || !password || posting}>
		{#if !posting}
			Login
		{:else}
			<Spinner class="bg-foreground-alt-3 size-4" />
		{/if}
	</SubmitButton>
</form>
