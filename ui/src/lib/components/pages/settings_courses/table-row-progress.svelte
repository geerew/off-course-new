<script lang="ts">
	import { Progress } from '$components/ui/progress';

	// ----------------------
	// Exports
	// ----------------------
	export let percent: number;
</script>

<div class="flex min-w-24 flex-row items-center gap-2.5">
	<Progress value={percent} class="h-2 bg-foreground/40" />
	<span class="min-w-5 text-xs text-muted-foreground">{percent}%</span>
</div>
