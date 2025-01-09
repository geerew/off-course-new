<script lang="ts">
	import { cn } from '$lib/utils';

	let { class: className = '' } = $props();
</script>

<span
	class={cn(
		'loader text-foreground-alt-2 relative size-[2.5em] rounded-full text-base text-[10px] before:absolute before:top-0 before:-left-[3.5em] before:size-[2.5em] before:rounded-full after:absolute after:top-0 after:left-[3.5em] after:size-[2.5em] after:rounded-full',
		className
	)}
></span>

<style>
	.loader,
	.loader:before,
	.loader:after {
		animation: bblFadInOut 1.5s infinite ease-in-out;
	}
	.loader {
		animation-delay: -0.16s;
	}

	.loader:before {
		animation-delay: -0.32s;
	}

	@keyframes bblFadInOut {
		0%,
		80%,
		100% {
			box-shadow: 0 2.5em 0 -1.3em;
		}
		40% {
			box-shadow: 0 2.5em 0 0;
		}
	}
</style>
