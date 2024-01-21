<script lang="ts">
	import { Progress } from '$components/ui/progress';

	// ----------------------
	// Exports
	// ----------------------
	export let percent: number;
</script>

<div class="flex min-w-24 flex-row items-center gap-2.5">
	<Progress value={percent} class="bg-primary/40 h-2" />
	<span class="text-muted-foreground min-w-5 text-xs">{percent}%</span>
</div>
