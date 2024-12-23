<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { Error, Spinner } from '$lib/components/';

	let { children } = $props();

	$effect(() => {
		auth.me();
	});
</script>

{#if auth.error !== null}
	<Error message={auth.error} />
{:else if auth.user === null}
	<div class="flex w-full justify-center pt-14 sm:pt-20">
		<Spinner />
	</div>
{:else}
	<div class="px-4 lg:px-8">
		{@render children()}
	</div>
{/if}
