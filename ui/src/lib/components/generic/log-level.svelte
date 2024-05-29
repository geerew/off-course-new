<script lang="ts">
	import Badge from '$components/ui/badge/badge.svelte';

	export let level: number;
</script>

{#if level === -4}
	<Badge variant="outline" class="bg-alt-1">DEBUG</Badge>
{:else if level === 0}
	<Badge variant="outline" class="bg-success">INFO</Badge>
{:else if level === 4}
	<Badge variant="outline" class="bg-primary">WARN</Badge>
{:else if level === 8}
	<Badge variant="outline" class="bg-destructive">ERROR</Badge>
{:else}
	{level}
{/if}
