<script lang="ts">
	import { Icons } from '$components/icons';
	import { COURSE_API, GetBackendUrl } from '$lib/api';
	import type { ClassName } from '$lib/types/general';
	import { cn } from '$lib/utils';
	import { createAvatar } from '@melt-ui/svelte';
	import { onMount } from 'svelte';

	let className: ClassName = undefined;

	// ----------------------
	// Exports
	// ----------------------

	// The course ID
	export let courseId: string;

	// Whether the course has a card
	export let hasCard: boolean;

	// When true, the card will refresh. Bind this to trigger a refresh from the parent component
	export let refresh = false;

	// Class overrides
	export { className as class };
	export let imgClass: ClassName = undefined;
	export let fallbackClass: ClassName = undefined;

	// ----------------------
	// Variables
	// ----------------------

	// Renders the course card
	const {
		elements: { image, fallback },
		options: { src }
	} = createAvatar();

	// True when the image is loading
	let isLoading = true;

	// True after the component is mounted
	let mounted = false;

	// ----------------------
	// Functions
	// ----------------------

	// Sets the src if the course has a card
	async function setSrc() {
		await new Promise((resolve) => setTimeout(resolve, isLoading ? 500 : 0));

		if (hasCard) {
			src.set(`${GetBackendUrl(COURSE_API)}/${courseId}/card?b=${new Date().getTime()}`);
		} else {
			src.set('');
		}

		isLoading = false;
	}

	// ----------------------
	// Reactive
	// ----------------------

	// When mounted and the course scan status is empty, set the src
	$: if (mounted && refresh) {
		isLoading = true;
		refresh = false;
		setSrc();
	}

	// ----------------------
	// Lifecycle
	// ----------------------
	onMount(async () => {
		await setSrc();
		mounted = true;
	});
</script>

<div class={className}>
	<img {...$image} alt="Course Card" class={imgClass} />

	<!-- Fallback -->
	<div class={fallbackClass} {...$fallback}>
		<Icons.Hexagon
			weight="fill"
			class={cn(
				'size-12 fill-muted/40 stroke-none md:size-16',
				isLoading ? 'animate-spin duration-2.5s' : 'rotate-90'
			)}
		/>
	</div>
</div>
