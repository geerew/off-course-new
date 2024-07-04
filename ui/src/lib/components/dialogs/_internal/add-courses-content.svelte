<script lang="ts">
	import { Err, Loading } from '$components/generic';
	import { Icons } from '$components/icons';
	import { Badge } from '$components/ui/badge';
	import { Button } from '$components/ui/button';
	import { Separator } from '$components/ui/separator';
	import { AddCourse, GetFileSystem } from '$lib/api';
	import { PathClassification, type DirInfo, type FileSystem } from '$lib/types/fileSystem';
	import { cn } from '$lib/utils';
	import { createEventDispatcher } from 'svelte';
	import { toast } from 'svelte-sonner';

	// ----------------------
	// Exports
	// ----------------------

	export let open: boolean;

	// ----------------------
	// Variables
	// ----------------------

	// Dispatcher. This is triggered when the user clicks the add button
	const dispatch = createEventDispatcher();

	// True when the initial loading of the drives/volumes is happening. We are considered to be
	// loading drives/volumes when the path is empty
	let loadingDrives = false;

	// True when a path is loading. It is used to disable clicking on other paths. When this is
	// true, a loading indicator will be shown on the path that is loading and the user will not
	// be able to click on other paths
	let loadingPath = false;

	// Set when an error of any kind occurs when loading the drives/volumes or a path
	let errorMsg = '';

	// True when the refresh button is click. This is used to show a loading indicator and will
	// prevent the user from clicking the refresh button again
	let refreshing = false;

	// Holds the information for the current level. When first opened, this will hold the drives
	// and when a path is clicked, this will hold the information for that path
	let pathInfo: FileSystem = {
		count: 0,
		directories: [],
		files: []
	};

	// A sequential list of paths. As the used navigates through the filesystem, the path is added
	// to this array. When the back button is clicked, the last path is popped from the array and
	// the user is taken to that path
	let paths: string[] = [];

	let currentLoadingPath = '';

	// This is bound to the content element and used to reset the scroll position to the top
	// following navigation
	let body: HTMLElement;

	// An array of the selected courses. When first opened this will be empty. As the user selects
	// and unselects courses, they will be added and removed from this array
	let selectedCourses: Record<string, string> = {};

	// ----------------------
	// Reactive
	// ----------------------

	// True when loading a drive/path or doing a refresh
	$: isLoadingOrRefreshing = loadingDrives || loadingPath || refreshing;

	// True when loading a drive/path, doing a refresh, got an error or the number of selected
	// courses is 0
	$: disableAddButton =
		loadingDrives ||
		loadingPath ||
		refreshing ||
		errorMsg !== '' ||
		Object.keys(selectedCourses).length === 0;

	$: if (open) {
		paths = [];
		selectedCourses = {};
		loadingDrives = false;
		loadingPath = false;
		errorMsg = '';
		refreshing = false;

		(async () => {
			await loadDrives();
		})();
	}

	// ----------------------
	// Function
	// ----------------------

	// Generic load function
	async function load(path: string) {
		try {
			const response = await GetFileSystem(path);

			if (body) body.scrollTop = 0;

			// Set the selected state of the directories. This will ensure previously selected courses are
			// still selected even as the user navigated
			response.directories?.forEach((d) => {
				if (d.path in selectedCourses) {
					// If the course is in the selectedCourses list, mark it as selected. Skip
					// the classification check as we want to allows the user to unselect selected
					// courses
					d.isSelected = true;
					return;
				}

				if (d.classification === PathClassification.None) {
					// Check if this path is an ancestor of any selected course
					d.classification = Object.keys(selectedCourses).some((key) => {
						return key.startsWith(d.path);
					})
						? PathClassification.Ancestor
						: PathClassification.None;
				}
			});

			return response;
		} catch (error) {
			errorMsg = error instanceof Error ? error.message : (error as string);
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Load the drives
	async function loadDrives() {
		loadingDrives = true;

		try {
			const flickerPromise = new Promise((resolve) => setTimeout(resolve, 200));
			const [loadResponse] = await Promise.all([load(''), flickerPromise]);
			if (loadResponse) pathInfo = loadResponse;
		} catch (error) {
			errorMsg = error instanceof Error ? error.message : (error as string);
		} finally {
			loadingDrives = false;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Load a path
	async function loadPath(path: string) {
		loadingPath = true;

		try {
			const flickerPromise = new Promise((resolve) => setTimeout(resolve, 200));
			const [loadResponse] = await Promise.all([load(path), flickerPromise]);
			if (loadResponse) pathInfo = loadResponse;
		} catch (error) {
			errorMsg = error instanceof Error ? error.message : (error as string);
		} finally {
			loadingPath = false;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Move into a path. This will add the path to the `paths` array then load that path
	async function moveInto(path: string) {
		currentLoadingPath = path;
		await loadPath(path);
		paths = [...paths, path];
	}
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Move back to the previous path. This will pop the last path. If the paths array is empty, the drives will
	// be loaded, else the previous path will be loaded
	async function moveBack() {
		paths.pop();
		paths.length === 0 ? await loadDrives() : await loadPath(paths[paths.length - 1]);
		paths = [...paths];
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Refresh the current path. This function is designed to take at least 1 second to prevent
	// flickering
	async function refresh() {
		if (refreshing) return;

		refreshing = true;

		try {
			const flickerPromise = new Promise((resolve) => setTimeout(resolve, 1000));
			const [loadResponse] = await Promise.all([load(currentLoadingPath), flickerPromise]);
			if (loadResponse) pathInfo = loadResponse;
		} catch (error) {
			errorMsg = error instanceof Error ? error.message : (error as string);
		}

		refreshing = false;
	}

	// Add selected courses
	async function add() {
		const keys = Object.keys(selectedCourses);

		for (let i = 0; i < keys.length; i++) {
			try {
				await AddCourse(selectedCourses[keys[i]], keys[i]);
				toast.success(`Course added: ${selectedCourses[keys[i]]}`);
			} catch (error) {
				toast.error(`Error adding course: ${selectedCourses[keys[i]]}`);
			}
		}

		dispatch('added');

		open = false;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Select all in the current path
	function selectAll() {
		let doRefresh = false;

		// Loop over all directories and mark them as selected if they have the classification none
		pathInfo.directories.forEach((d) => {
			if (d.classification === PathClassification.None && !d.isSelected) {
				d.isSelected = true;
				selectedCourses[d.path] = d.title;
				doRefresh = true;
			}
		});

		if (doRefresh) {
			pathInfo = { ...pathInfo };
			selectedCoursesToast();
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Unselect all in the current path
	function unselectAll() {
		let doRefresh = false;

		// Loop over all directories and mark them as unselected if they have the classification none
		pathInfo.directories.forEach((d) => {
			if (d.classification === PathClassification.None && d.isSelected) {
				d.isSelected = false;
				delete selectedCourses[d.path];
				doRefresh = true;
			}
		});

		if (doRefresh) {
			pathInfo = { ...pathInfo };
			selectedCoursesToast();
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Checks if the course is in the selectedCourses array. If it is, it will be removed, making
	// it unselected. If it is not, it will be added, making it selected.
	function FlipSelected(dirInfo: DirInfo) {
		if (dirInfo.classification !== PathClassification.None) return;

		// Update selected courses (either add and remove)
		dirInfo.path in selectedCourses
			? delete selectedCourses[dirInfo.path]
			: (selectedCourses[dirInfo.path] = dirInfo.title);

		// Flip the selected state
		dirInfo.isSelected = !dirInfo.isSelected;

		// Update to trigger a re-render
		pathInfo = { ...pathInfo };

		selectedCoursesToast();
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Display a toast when a course is selected/deselected
	const selectedCoursesToast = () => {
		const count = Object.keys(selectedCourses).length;
		let message = 'Selected ' + count + ' course' + (count > 1 ? 's' : '');

		if (count === 0) message = 'Deselected all courses';

		toast.success(message, {
			duration: 2000
		});
	};
</script>

{#if open}
	<!-- Header -->
	<div
		class="border-alt-1/60 flex h-14 shrink-0 items-center justify-between border-b px-3 text-base font-medium"
	>
		<div class="flex items-center gap-2">
			<Icons.StackPlus class="size-4" />
			<span>Course Selection</span>
		</div>

		<!-- Refresh -->
		<Button
			variant="ghost"
			disabled={isLoadingOrRefreshing}
			class="hover:bg-alt-1 group px-2.5 disabled:opacity-100"
			on:click={refresh}
		>
			<Icons.Refresh
				class={cn(
					'group-hover:text-foreground text-muted-foreground size-5',
					refreshing && 'animate-spin'
				)}
			/>
		</Button>
	</div>

	<!-- Body -->
	<div
		bind:this={body}
		class="flex min-h-[10rem] grow flex-col overflow-x-hidden overflow-y-scroll"
		tabindex="-1"
	>
		{#if loadingDrives}
			<Loading />
		{:else if errorMsg}
			<Err class="min-h-max" errorMessage={errorMsg} />
		{:else}
			<div class="flex flex-col">
				<!-- Back button -->
				{#if paths.length > 0}
					{#key paths[paths.length - 1]}
						<div class="border-alt-1/40 flex h-14 flex-row items-center border-b">
							<Button
								variant="ghost"
								disabled={loadingPath || refreshing}
								class="hover:bg-alt-1/40 flex h-14 flex-grow flex-row items-center justify-start rounded-none border-b pr-0"
								on:click={async (el) => {
									// find this buttons child element and show the loader by removing the hidden and adding the flex
									if (!el.target || !(el.target instanceof Element)) return;

									const buttonElement = el.target.closest('button');
									if (!buttonElement) return;

									const loader = buttonElement.nextElementSibling;
									if (loader) {
										loader.classList.remove('hidden');
										loader.classList.add('flex');
									}

									// Determine the new back path. If we are only 1 level
									// deep, load the drives, or else load the path before
									await moveBack();
								}}
							>
								<div class="flex grow gap-2 text-sm">
									<Icons.CornerUpLeft
										class="text-muted-foreground group-hover:text-foreground size-4"
									/>
									<span>Back</span>
								</div>
							</Button>

							<div
								class="hidden h-full min-w-20 shrink-0 place-content-center items-center opacity-100"
								id="back-loader"
							>
								<Loading class="px-0 py-0" loaderClass="size-5" />
							</div>
						</div>
					{/key}
				{/if}

				<!-- Directories -->
				{#each pathInfo.directories as dirInfo (dirInfo.path)}
					<div class="border-alt-1/40 flex h-14 flex-row items-center border-b last:border-none">
						<!-- Path (left) -->
						<Button
							variant="ghost"
							disabled={loadingPath ||
								refreshing ||
								dirInfo.classification === PathClassification.Course ||
								dirInfo.isSelected}
							class="hover:bg-alt-1/20 h-full flex-grow justify-start whitespace-normal rounded-none text-start"
							on:click={async () => {
								await moveInto(dirInfo.path);
							}}
						>
							<span class="flex grow text-sm">{dirInfo.title}</span>
						</Button>

						<div class="flex h-full w-20 shrink-0 place-content-center items-center">
							<Separator orientation="vertical" class="bg-alt-1/40 h-full" />

							{#if dirInfo.classification !== PathClassification.Course}
								<!-- Checkbox (right) -->
								{#if loadingPath && dirInfo.path === currentLoadingPath}
									<Loading class="px-0 py-0" loaderClass="size-5" />
								{:else}
									<Button
										variant="ghost"
										tabindex={dirInfo.classification === PathClassification.Ancestor ? -1 : 0}
										disabled={loadingPath ||
											refreshing ||
											dirInfo.classification === PathClassification.Ancestor}
										class="hover:bg-alt-1/20 group h-full w-full shrink-0 place-content-center items-center rounded-none duration-200 disabled:opacity-100 sm:w-20"
										on:click={() => {
											FlipSelected(dirInfo);
										}}
									>
										<input
											class="bg-muted group-hover:border-muted-foreground checked:bg-primary checked:hover:bg-primary indeterminate:bg-secondary pointer-events-none cursor-pointer rounded border-2 p-2 duration-200 checked:border-transparent indeterminate:opacity-60 group-hover:checked:border-transparent group-hover:checked:brightness-90"
											tabindex="-1"
											checked={dirInfo.isSelected ?? false}
											type="checkbox"
											indeterminate={dirInfo.classification === PathClassification.Ancestor}
										/>
									</Button>
								{/if}
							{:else}
								<div class="flex w-full place-content-center">
									<Badge
										variant="outline"
										class={cn(
											'text-muted-foreground/60 border-muted-foreground/40 rounded px-1.5 text-center text-xs'
										)}
									>
										Added
									</Badge>
								</div>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Footer -->
	<div
		class="border-alt-1/60 flex shrink-0 flex-row items-center justify-between gap-3 border-t p-3"
	>
		<!-- Select/unselect -->
		<div class="hidden gap-3 sm:flex">
			<Button
				variant="outline"
				disabled={isLoadingOrRefreshing}
				class="bg-muted border-alt-1/60 hover:bg-alt-1/60 group h-8 w-24"
				on:click={selectAll}
			>
				Select All
			</Button>

			<Button
				variant="outline"
				disabled={isLoadingOrRefreshing}
				class="bg-muted border-alt-1/60 hover:bg-alt-1/60 w-26 group h-8"
				on:click={unselectAll}
			>
				Unselect All
			</Button>
		</div>

		<!-- Close/add -->
		<div class="flex w-full justify-end gap-3">
			<Button
				variant="outline"
				class="bg-muted border-alt-1/60 hover:bg-alt-1/60 h-8 w-20"
				on:click={() => (open = false)}
			>
				Cancel
			</Button>

			<Button class="h-8 px-6" disabled={disableAddButton} on:click={add}>Add</Button>
		</div>
	</div>
{/if}
