<script lang="ts">
	import * as Avatar from '$components/ui/avatar';
	import { COURSE_API } from '$lib/api';
	import { Play } from 'lucide-svelte';
	import { onMount } from 'svelte';

	// ----------------------
	// Exports
	// ----------------------

	export let courseId: string;
	export let hasCard: boolean;
	export let cardRefresh: boolean;

	// ----------------------
	// Variables
	// ----------------------

	// This will be set to the src of the course card if the course has a card
	let cardSrc = '';

	// ----------------------
	// Functions
	// ----------------------

	// Sets the src if the course has a card
	function setCardSrc() {
		cardRefresh = false;

		if (hasCard) {
			cardSrc = `${COURSE_API}/${courseId}/card?b=${new Date().getTime()}`;
		} else {
			cardSrc = '';
		}
	}

	// ----------------------
	// Reactive
	// ----------------------

	// Update course chapters when `assetRefresh` is set to true
	$: if (cardRefresh) {
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
		<Avatar.Image src={cardSrc} class="mx-auto min-h-0 max-w-full rounded-lg" />
		<Avatar.Fallback
			class="bg-background mx-auto flex h-48 max-w-72 place-content-center rounded-lg lg:w-full"
		>
			<Play class="fill-primary text-primary h-12 w-12 opacity-60" />
		</Avatar.Fallback>
	</Avatar.Root>
</div>
