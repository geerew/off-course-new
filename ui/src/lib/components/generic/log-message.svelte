<script lang="ts">
	import Badge from '$components/ui/badge/badge.svelte';

	export let message: string;
	export let data: Record<string, any>;
</script>

<div class="flex flex-col gap-2.5">
	<span>{message}</span>
	<div class="flex flex-row flex-wrap items-center gap-2.5">
		{#if data}
			{#each Object.keys(data) as item}
				{#if item === 'execTime'}
					<Badge variant="outline" class="bg-alt-1 px-1.5">{item}: {Math.floor(data[item])}ms</Badge
					>
				{:else if item === 'error'}
					<Badge variant="outline" class="bg-destructive px-1.5">{item}: {data[item]}</Badge>
				{:else}
					<Badge variant="outline" class="bg-alt-1 px-1.5">{item}: {data[item]}</Badge>
				{/if}
			{/each}
		{/if}
	</div>
</div>
