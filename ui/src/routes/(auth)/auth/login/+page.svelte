<script lang="ts">
	import { Logo } from '$lib/components';
	import FormLogin from '$lib/components/form/form-login.svelte';
</script>

<div class="flex h-full min-h-110 flex-col items-center justify-center gap-8">
	<Logo />

	<div class="flex min-w-sm flex-col gap-5">
		<div class="mb-2.5 flex flex-col gap-2">
			<div class="text-foreground-alt-1 text-center text-lg">Login to your account</div>
		</div>

		<FormLogin endpoint="/api/auth/login" />

		<div class="text-foreground-alt-2 text-center">
			Don't have an account?
			<a href="/auth/register/" class="hover:text-background-primary font-semibold duration-200">
				Sign up
			</a>
		</div>
	</div>
</div>
