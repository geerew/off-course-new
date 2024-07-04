<script lang="ts">
	import { Icons } from '$components/icons';

	// ----------------------
	// Exports
	// ----------------------
	export let available: boolean;
</script>

<div class="flex place-content-center">
	{#if available}
		<Icons.SquareCheck
			class="fill-success stroke-success [&>:nth-child(2)]:stroke-success-foreground"
		/>
	{:else}
		<Icons.SquareX class="fill-destructive stroke-destructive [&>:nth-child(n+2)]:stroke-white" />
	{/if}
</div>
