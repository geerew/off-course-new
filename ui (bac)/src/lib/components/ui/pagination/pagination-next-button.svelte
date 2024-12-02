<script lang="ts">
	import { Icons } from '$components/icons';
	import { Button } from '$lib/components/ui/button';
	import { cn } from '$lib/utils';
	import { Pagination as PaginationPrimitive } from 'bits-ui';

	type $$Props = PaginationPrimitive.NextButtonProps;
	type $$Events = PaginationPrimitive.NextButtonEvents;

	let className: $$Props['class'] = undefined;
	export { className as class };
</script>

<PaginationPrimitive.NextButton asChild let:builder>
	<Button
		variant="ghost"
		class={cn('gap-1 pr-2.5', className)}
		builders={[builder]}
		on:click
		{...$$restProps}
	>
		<slot>
			<span>Next</span>
			<Icons.CaretRight class="h-4 w-4" />
		</slot>
	</Button>
</PaginationPrimitive.NextButton>
