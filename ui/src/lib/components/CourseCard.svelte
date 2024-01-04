<script lang="ts">
	import type { ClassName } from '$lib/types/general';
	import type { Course } from '$lib/types/models';
	import { COURSE_API } from '$lib/utils/api';
	import { cn } from '$lib/utils/general';
	import { createAvatar } from '@melt-ui/svelte';
	import { onMount } from 'svelte';
	import { Icons } from './icons';

	let className: ClassName = undefined;

	// ----------------------
	// Exports
	// ----------------------

	export let course: Course;
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

	// True after the component is mounted
	let mounted = false;

	// ----------------------
	// Functions
	// ----------------------

	// Sets the src if the course has a card
	function setSrc() {
		if (course && course.hasCard) {
			src.set(`${COURSE_API}/${course.id}/card?b=${new Date().getTime()}`);
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

<div class={className}>
	<img {...$image} alt="Course Card" class={cn('rounded-md', imgClass)} />

	<slot>
		<div
			class={cn(
				'bg-accent-1 flex h-48 w-[20rem] !cursor-default place-content-center items-center rounded-md lg:w-full',
				fallbackClass
			)}
			{...$fallback}
		>
			<Icons.play class="h-12 w-12 fill-gray-400 text-gray-400" />
		</div>
	</slot>
</div>
