<script lang="ts">
	import { AspectRatio } from '$components/ui/aspect-ratio';
	import { COURSE_API, GetBackendUrl } from '$lib/api';
	import type { Course } from '$lib/types/models';
	import { createAvatar } from '@melt-ui/svelte';
	import { Play } from 'lucide-svelte';
	import { onMount } from 'svelte';

	// ----------------------
	// Exports
	// ----------------------
	export let course: Course;

	// ----------------------
	// Variables
	// ----------------------

	// Renders the course card
	const {
		elements: { image, fallback },
		options: { src }
	} = createAvatar();

	// True after the component is mounted
	let mounted = false;

	// ----------------------
	// Functions
	// ----------------------

	// Sets the src if the course has a card
	function setSrc() {
		if (course && course.hasCard) {
			src.set(`${GetBackendUrl(COURSE_API)}/${course.id}/card?b=${new Date().getTime()}`);
		} else {
			src.set('');
		}
	}

	// ----------------------
	// Reactive
	// ----------------------

	// When mounted and the course scan status is empty, set the src
	$: if (mounted && course.scanStatus === '') {
		setSrc();
	}

	// ----------------------
	// Lifecycle
	// ----------------------
	onMount(() => {
		setSrc();
		mounted = true;
	});
</script>

<AspectRatio ratio={16 / 9} class="bg-muted mx-h-48 overflow-hidden">
	<!-- Image -->
	<img {...$image} alt="Course Card" class="w-full" />

	<!-- Fallback -->
	<div
		class="bg-background mx-auto flex h-full w-full place-content-center items-center lg:w-full"
		{...$fallback}
	>
		<Play class="fill-primary text-primary h-12 w-12 opacity-60" />
	</div>
</AspectRatio>
