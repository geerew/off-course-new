<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { Error, Header, Spinner } from '$lib/components/';

	let { children } = $props();

	$effect(() => {
		auth.me();
	});
</script>

{#if auth.error !== null}
	<Error message={'Failed to fetch user: ' + auth.error} />
{:else if auth.user === null}
	<div class="flex w-full justify-center pt-14 sm:pt-20">
		<Spinner class="bg-foreground-alt-2 size-6" />
	</div>
{:else}
	<div class="px-4 pb-8 lg:px-8">
		<Header />
		{@render children()}
	</div>
{/if}
