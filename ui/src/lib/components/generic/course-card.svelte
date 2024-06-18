<script lang="ts">
	import { COURSE_API, GetBackendUrl } from '$lib/api';
	import type { Course } from '$lib/types/models';
	import { createAvatar } from '@melt-ui/svelte';
	import { Hexagon } from 'lucide-svelte';
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

<div class="aspect-w-16 aspect-h-7 sm:aspect-w-16 sm:aspect-h-7">
	<img
		{...$image}
		alt="Course Card"
		class="rounded-lg object-cover object-center sm:rounded-b-none md:object-top"
	/>

	<!-- Fallback -->
	<div
		class="bg-alt-1 mx-auto flex place-content-center items-center rounded-lg sm:rounded-b-none lg:w-full"
		{...$fallback}
	>
		<Hexagon class="fill-muted/40 size-12 rotate-90 stroke-none md:size-16" />
	</div>
</div>
