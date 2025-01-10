<script lang="ts">
	import { cn } from '$lib/utils';

	let { class: componentClass }: { class?: string } = $props();
</script>

<div class="flex flex-row items-center gap-2">
	<div
		class={cn(
			'animate-flash size-[2.5em] rounded-full bg-red-400 delay-[-0.32s] duration-1500 ease-in-out',
			componentClass
		)}
	></div>
	<div
		class={cn(
			'animate-flash size-[2.5em] rounded-full bg-red-400 delay-[-0.16s] duration-1500 ease-in-out',
			componentClass
		)}
	></div>
	<div
		class={cn(
			'animate-flash size-[2.5em] rounded-full bg-red-400 duration-1500 ease-in-out',
			componentClass
		)}
	></div>
</div>
