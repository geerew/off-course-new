<script lang="ts">
	import LatestCourses from '$components/LatestCourses.svelte';
	import OngoingCourses from '$components/OngoingCourses.svelte';
</script>

<div class="container px-4 py-4 md:px-8 md:py-8">
	<div class="flex flex-col gap-4 md:gap-10">
		<OngoingCourses />

		<LatestCourses />
	</div>
</div>
