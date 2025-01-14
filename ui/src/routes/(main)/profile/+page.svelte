<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { Spinner } from '$lib/components';
	import { Input, SubmitButton } from '$lib/components/form';
	import InputPassword from '$lib/components/form/input-password.svelte';
	import { Edit } from '$lib/components/icons';
	import { AlertDialog, Dialog, Separator } from 'bits-ui';
	import { toast } from 'svelte-sonner';

	// Dialog controls
	let dialogDisplayName = $state(false);
	let dialogPassword = $state(false);
	let dialogDeleteAccount = $state(false);

	// Input elements for focus
	let displayNameInputEl = $state<HTMLInputElement>();
	let passwordCurrentEl = $state<HTMLInputElement>();

	let displayNameValue = $state<string>('');

	// Password fields
	let currentPasswordValue = $state('');
	let newPasswordValue = $state('');
	let confirmPasswordValue = $state('');

	// False when any of the password fields are empty
	let passwordSubmitDisabled = $derived.by(() => {
		return currentPasswordValue === '' || newPasswordValue === '' || confirmPasswordValue === '';
	});

	// True when a request is being made
	let isPosting = $state(false);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Send a DELETE request to delete the account
	async function submitDeleteAccount(event: Event) {
		event.preventDefault();
		isPosting = true;

		const response = await fetch('/api/auth/me', {
			method: 'DELETE',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				current_password: currentPasswordValue
			})
		});

		if (response.ok) {
			auth.empty();
			window.location.href = '/auth/login';
		} else {
			const data = await response.json();
			toast.error(`${data.message}`);
			isPosting = false;
		}
	}

	// Send a PUT request to update the display name
	async function submitDisplayNameForm(event: Event) {
		event.preventDefault();
		isPosting = true;

		const response = await fetch('/api/auth/me', {
			method: 'PUT',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				display_name: displayNameValue
			})
		});

		if (response.ok) {
			await auth.me();
			dialogDisplayName = false;
		} else {
			const data = await response.json();
			toast.error(data.message);
			isPosting = false;
		}
	}

	// Send a PUT request to update the password
	async function submitPasswordForm(event: Event) {
		event.preventDefault();
		isPosting = true;

		if (newPasswordValue !== confirmPasswordValue) {
			toast.error('Passwords do not match');
			isPosting = false;
			return;
		}

		const response = await fetch('/api/auth/me', {
			method: 'PUT',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				current_password: currentPasswordValue,
				password: newPasswordValue
			})
		});

		if (response.ok) {
			await auth.me();
			dialogPassword = false;
		} else {
			const data = await response.json();
			toast.error(data.message);
			isPosting = false;
		}
	}
</script>

{#if auth.user !== null}
	<div class="mx-auto flex max-w-2xl flex-col place-content-center items-start gap-5">
		<!-- Username -->
		<div class="flex flex-col gap-3">
			<div class="text-foreground-alt-2 text-[15px] uppercase">Username</div>
			<span class="text-background-primary text-2xl">{auth.user.username}</span>
		</div>

		<Separator.Root class="bg-background-alt-3 my-2 h-px w-full shrink-0" />

		<!-- Display name -->
		<div class="flex flex-col gap-3">
			<div class="flex flex-row items-center gap-3">
				<div class="text-foreground-alt-2 text-[15px] uppercase">Display Name</div>

				<Dialog.Root
					bind:open={dialogDisplayName}
					onOpenChange={() => {
						displayNameValue = '';
						isPosting = false;
					}}
				>
					<Dialog.Trigger
						class="text-foreground-alt-2 hover:text-foreground-alt-1 mb-0.5 cursor-pointer duration-200"
					>
						<Edit class="size-4.5 stroke-2" />
					</Dialog.Trigger>

					<Dialog.Portal>
						<Dialog.Overlay
							class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80"
						/>
						<Dialog.Content
							class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95 data-[state=open]:slide-in-from-top-5 bg-background-alt-1 data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:slide-out-to-top-5 fixed top-20 left-1/2 z-50 w-[20rem] -translate-x-1/2 overflow-hidden rounded-lg data-[state=closed]:duration-200 data-[state=open]:duration-200"
							onOpenAutoFocus={(e) => {
								e.preventDefault();
								displayNameInputEl?.focus();
							}}
							onCloseAutoFocus={(e) => {
								e.preventDefault();
							}}
						>
							<form onsubmit={submitDisplayNameForm}>
								<div class="flex flex-col gap-2.5 p-5">
									<div>Display Name:</div>
									<Input
										bind:ref={displayNameInputEl}
										bind:value={displayNameValue}
										name="display name"
										type="text"
										placeholder={auth.user.displayName}
									/>
								</div>

								<div
									class="bg-background-alt-2 border-background-alt-3 flex w-full items-center justify-end gap-2 border-t px-5 py-2.5"
								>
									<Dialog.Close
										type="button"
										class="border-background-alt-4 text-foreground-alt-1 hover:bg-background-alt-4 hover:text-foreground w-24 cursor-pointer rounded-md border py-2 duration-200 select-none"
									>
										Cancel
									</Dialog.Close>

									<SubmitButton
										type="submit"
										disabled={displayNameValue === '' || isPosting}
										class="h-10 w-24 py-2"
									>
										{#if !isPosting}
											Update
										{:else}
											<Spinner class="bg-foreground-alt-3 size-2" />
										{/if}
									</SubmitButton>
								</div>
							</form>
						</Dialog.Content>
					</Dialog.Portal>
				</Dialog.Root>
			</div>
			<span class="text-background-primary text-2xl">{auth.user.displayName}</span>
		</div>

		<Separator.Root class="bg-background-alt-3 my-2 h-px w-full shrink-0" />

		<!-- Password -->
		<div class="flex flex-col gap-3">
			<div class="text-foreground-alt-2 text-[15px] uppercase">Password</div>

			<Dialog.Root
				bind:open={dialogPassword}
				onOpenChange={() => {
					currentPasswordValue = '';
					newPasswordValue = '';
					confirmPasswordValue = '';
					isPosting = false;
				}}
			>
				<Dialog.Trigger
					class="bg-background-alt-4 hover:bg-background-alt-5 text-foreground-alt-1 hover:text-foreground w-38 cursor-pointer rounded-md py-2 duration-200 select-none"
				>
					Change Password
				</Dialog.Trigger>

				<Dialog.Portal>
					<Dialog.Overlay
						class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80"
					/>
					<Dialog.Content
						class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95 data-[state=open]:slide-in-from-top-5 bg-background-alt-1 data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:slide-out-to-top-5 fixed top-20 left-1/2 z-50 w-[20rem] -translate-x-1/2 overflow-hidden rounded-lg data-[state=closed]:duration-200 data-[state=open]:duration-200"
						onOpenAutoFocus={(e) => {
							e.preventDefault();
							passwordCurrentEl?.focus();
						}}
						onCloseAutoFocus={(e) => {
							e.preventDefault();
						}}
					>
						<form onsubmit={submitPasswordForm}>
							<div class="flex flex-col gap-4 p-5">
								<div class="flex flex-col gap-2.5">
									<div>Current Password:</div>
									<InputPassword
										bind:ref={passwordCurrentEl}
										bind:value={currentPasswordValue}
										name="current password"
									/>
								</div>

								<Separator.Root class="bg-background-alt-3 mt-2 h-px w-full shrink-0" />

								<div class="flex flex-col gap-2.5">
									<div>New Password:</div>
									<InputPassword bind:value={newPasswordValue} name="new password" />
								</div>

								<div class="flex flex-col gap-2.5">
									<div>Confirm Password:</div>
									<InputPassword bind:value={confirmPasswordValue} name="confirm password" />
								</div>
							</div>

							<div
								class="bg-background-alt-2 border-background-alt-3 flex w-full items-center justify-end gap-2 border-t px-5 py-2.5"
							>
								<Dialog.Close
									type="button"
									class="border-background-alt-4 text-foreground-alt-1 hover:bg-background-alt-4 hover:text-foreground w-24 cursor-pointer rounded-md border py-2 duration-200 select-none"
								>
									Cancel
								</Dialog.Close>

								<SubmitButton
									type="submit"
									disabled={passwordSubmitDisabled || isPosting}
									class="h-10 w-24 py-2"
								>
									{#if !isPosting}
										Update
									{:else}
										<Spinner class="bg-foreground-alt-3 size-2" />
									{/if}
								</SubmitButton>
							</div>
						</form>
					</Dialog.Content>
				</Dialog.Portal>
			</Dialog.Root>
		</div>

		<Separator.Root class="bg-background-alt-3 my-2 h-px w-full shrink-0" />

		<!-- Delete account -->
		<div class="flex flex-col gap-3">
			<div class="text-foreground-alt-2 text-[15px] uppercase">Delete Account</div>
			<AlertDialog.Root
				bind:open={dialogDeleteAccount}
				onOpenChange={() => {
					currentPasswordValue = '';
					isPosting = false;
				}}
			>
				<AlertDialog.Trigger
					class="bg-background-error hover:bg-background-error-alt-1 text-foreground-alt-1 hover:text-foreground w-36 cursor-pointer rounded-md py-2 duration-200 select-none"
				>
					Delete Account
				</AlertDialog.Trigger>

				<AlertDialog.Portal>
					<AlertDialog.Overlay
						class="bg-background/70 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50"
					/>

					<AlertDialog.Content
						interactOutsideBehavior="close"
						onOpenAutoFocus={(e) => {
							e.preventDefault();
							passwordCurrentEl?.focus();
						}}
						onCloseAutoFocus={(e) => {
							e.preventDefault();
						}}
						class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95 data-[state=open]:slide-in-from-top-5 data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:slide-out-to-top-5 fixed top-20 left-1/2 z-50 w-full max-w-lg min-w-[20rem] -translate-x-1/2 overflow-hidden px-10 data-[state=closed]:duration-200 data-[state=open]:duration-200"
					>
						<div class="bg-background-alt-1 overflow-hidden rounded-lg">
							<form onsubmit={submitDeleteAccount}>
								<div class="flex flex-col gap-2.5 p-5">
									<div class="flex items-center justify-center">
										<svg
											xmlns="http://www.w3.org/2000/svg"
											fill="none"
											viewBox="0 0 24 24"
											stroke-width="1.5"
											stroke="currentColor"
											class="text-foreground-error size-14"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z"
											/>
										</svg>
									</div>

									<AlertDialog.Description
										class="text-foreground-alt-1 flex flex-col gap-2 text-center"
									>
										<span class="text-lg">Are you sure you want to delete your account?</span>
										<span class="text-foreground-alt-2">All associated data will be deleted</span>
									</AlertDialog.Description>

									<Separator.Root class="bg-background-alt-3 mt-2 h-px w-full shrink-0" />

									<div class="flex max-w-[16rem] flex-col gap-2.5 px-2.5">
										<div>Confirm Password:</div>
										<InputPassword
											bind:ref={passwordCurrentEl}
											bind:value={currentPasswordValue}
											name="current password"
										/>
									</div>
								</div>

								<div
									class="bg-background-alt-2 border-background-alt-3 flex w-full items-center justify-end gap-2 border-t px-5 py-2.5"
								>
									<AlertDialog.Cancel
										type="button"
										class="border-background-alt-4 text-foreground-alt-1 hover:bg-background-alt-4 hover:text-foreground w-24 cursor-pointer rounded-md border py-2 duration-200 select-none"
									>
										Cancel
									</AlertDialog.Cancel>

									<SubmitButton
										type="submit"
										disabled={currentPasswordValue === '' || isPosting}
										class="bg-background-error disabled:bg-background-error/80 enabled:hover:bg-background-error-alt-1 text-foreground-alt-1 enabled:hover:text-foreground h-10 w-24 py-2"
									>
										{#if !isPosting}
											Delete
										{:else}
											<Spinner class="bg-foreground-alt-1 size-2" />
										{/if}
									</SubmitButton>
								</div>
							</form>
						</div>
					</AlertDialog.Content>
				</AlertDialog.Portal>
			</AlertDialog.Root>
		</div>
	</div>
{/if}
