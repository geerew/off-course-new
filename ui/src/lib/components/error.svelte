<script lang="ts">
	let { message } = $props();
</script>

<div class="flex w-full flex-col gap-2.5 pt-14 sm:pt-20">
	<div class="flex flex-row items-center justify-center gap-2.5 text-lg">
		<svg
			xmlns="http://www.w3.org/2000/svg"
			fill="none"
			viewBox="0 0 24 24"
			stroke-width="1.5"
			stroke="currentColor"
			class="size-6"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z"
			/>
		</svg>
		<p class="text-foreground-alt-1 text-center">{message}</p>
	</div>

	<div class="text-foreground-alt-2 flex justify-center gap-2.5">Please try again</div>
</div>
