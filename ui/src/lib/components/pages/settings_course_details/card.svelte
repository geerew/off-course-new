<script lang="ts">
	import Loading from '$components/generic/loading.svelte';
	import * as Avatar from '$components/ui/avatar';
	import { COURSE_API, GetBackendUrl } from '$lib/api';
	import { Play } from 'lucide-svelte';
	import { onMount } from 'svelte';

	// ----------------------
	// Exports
	// ----------------------

	export let courseId: string;
	export let hasCard: boolean;
	export let refresh: boolean;

	// ----------------------
	// Variables
	// ----------------------

	// This will be set to the src of the course card if the course has a card
	let cardSrc = '';

	// ----------------------
	// Functions
	// ----------------------

	// Sets the src if the course has a card. During a refresh, there will be a small delay to
	// prevent flickering
	async function setCardSrc() {
		await new Promise((resolve) => setTimeout(resolve, refresh ? 500 : 0));

		refresh = false;

		if (hasCard) {
			cardSrc = `${GetBackendUrl(COURSE_API)}/${courseId}/card?b=${new Date().getTime()}`;
		} else {
			cardSrc = '';
		}
	}

	// ----------------------
	// Reactive
	// ----------------------

	// Update course chapters when `assetRefresh` is set to true
	$: if (refresh) {
		setCardSrc();
	}

	// ----------------------
	// Lifecycle
	// ----------------------

	onMount(() => {
		setCardSrc();
	});
</script>

<!-- Card -->
<div class="order-1 text-xl font-bold md:order-2 md:text-2xl">
	<Avatar.Root class="flex h-48 max-h-48 w-auto flex-col rounded-none">
		{#if refresh}
			<Loading class="border-alt-1 mx-auto h-48 max-w-72 rounded-lg border" />
		{:else}
			<Avatar.Image
				src={cardSrc}
				class="border-alt-1/60 mx-auto min-h-0 max-w-full rounded-lg border"
			/>

			<Avatar.Fallback
				class="bg-background mx-auto flex h-48 max-w-72 place-content-center rounded-lg lg:w-full"
			>
				<Play class="fill-primary text-primary h-12 w-12 opacity-60" />
			</Avatar.Fallback>
		{/if}
	</Avatar.Root>
</div>
