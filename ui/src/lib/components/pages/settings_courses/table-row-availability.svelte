<script lang="ts">
	import { CheckSquare2Icon, XSquare } from 'lucide-svelte';

	// ----------------------
	// Exports
	// ----------------------
	export let available: boolean;
</script>

<div class="flex place-content-center">
	{#if available}
		<span title="available">
			<CheckSquare2Icon
				class="fill-success stroke-success [&>:nth-child(2)]:stroke-success-foreground"
			>
				Available
			</CheckSquare2Icon>
		</span>
	{:else}
		<span title="unavailable">
			<XSquare class="fill-destructive stroke-destructive [&>:nth-child(n+2)]:stroke-white">
				Unavailable
			</XSquare>
		</span>
	{/if}
</div>
