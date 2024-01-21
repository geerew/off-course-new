<script lang="ts">
	import { Error, Loading } from '$components';
	import Badge from '$components/ui/badge/badge.svelte';
	import { Button } from '$components/ui/button';
	import { Separator } from '$components/ui/separator';
	import * as Sheet from '$components/ui/sheet';
	import { AddCourse, ErrorMessage, GetAllCourses, GetFileSystem } from '$lib/api';
	import { addToast } from '$lib/stores/addToast';
	import type { FileInfo, FileSystem } from '$lib/types/fileSystem';
	import { cn } from '$lib/utils';
	import { BookPlus, CornerUpLeft, RefreshCw } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';

	// ----------------------
	// Variables
	// ----------------------

	// Dispatcher. This is triggered when the user clicks the add button
	const dispatch = createEventDispatcher();

	// True when the sheet is open. This is used to reset stuff on open
	let isOpen = false;

	// True when the initial loading of the drives/volumes is happening. We are considered to be
	// loading drives/volumes when the path is empty
	let loadingDrives = false;

	// True when a path is loading. It is used to disable clicking on other paths. When this is
	// true, a loading indicator will be shown on the path that is loading and the user will not
	// be able to click on other paths
	let loadingPath = false;

	// True when an error of any kind occurs when loading the drives/volumes or a path
	let gotError = false;

	// True when the refresh button is click. This is used to show a loading indicator and will
	// prevent the user from clicking the refresh button again
	let refreshing = false;

	// Holds the information for the current level. When first opened, this will hold the drives
	// and when a path is clicked, this will hold the information for that path
	let pathInfo: FileSystem;

	// This is bound to the content element and used to reset the scroll position to the top
	// following navigation
	let body: HTMLElement;

	// A sequential list of paths. As the used navigates through the filesystem, the path is added
	// to this array. When the back button is clicked, the last path is popped from the array and
	// the user is taken to that path
	let paths: string[] = [];

	// The currently selected path. This is used to show a loading indicator on the correct row
	let selectedPath = '';

	// An array of the selected courses. When first opened this will be empty. As the user selects
	// and unselects courses, they will be added and removed from this array
	let selectedCourses: Record<string, string> = {};

	// This is used by the select all button. As courses are manually selected they will be removed
	// from this array and when they are manually unselected they will be added back to this array.
	// When the used clicks the select all button, the courses in this array will be added to the
	// selected courses array
	let selectableCourses: FileInfo[] = [];

	// True when we are required to pull the existing courses from the backend
	let getExistingCourses = true;

	// An array of the existing courses. This is pulled from the backend and used to disable
	// existing courses from being selected/unselected
	let existingCourses: string[] = [];

	// ----------------------
	// Reactive
	// ----------------------

	// Reset everything every time the sheet is opened
	$: if (isOpen) {
		paths = [];
		selectedCourses = {};
		selectableCourses = [];
		loadingDrives = false;
		loadingPath = false;
		gotError = false;
		refreshing = false;
		getExistingCourses = true;

		// Load the drives when first opened
		(async () => await load('', false, false))();
	}

	// True when loading a drive/path or doing a refresh
	$: isLoadingOrRefreshing = loadingDrives || loadingPath || refreshing;

	// True when loading a drive/path, doing a refresh, got an error or the number of selected
	// courses is 0
	$: disableAddButton =
		loadingDrives ||
		loadingPath ||
		refreshing ||
		gotError ||
		Object.keys(selectedCourses).length === 0;

	// ----------------------
	// Functions
	// ----------------------

	// Load will load the drives or information about a path, based upon when path is populated.
	//
	// When path is empty, the drives will be loaded and when path is populated, information about
	// the path will be loaded, such as files and directories.
	//
	// When movingBack is true, the last path will be popped from the paths array. When false, the
	// path will be added to the paths array.
	//
	// When refresh is true, the paths array will not be manipulated
	const load = async (path: string, movingBack: boolean, refresh: boolean) => {
		path ? (loadingPath = true) : (loadingDrives = true);

		if (path) selectedPath = path;

		// Pull the existing courses so we can identify already added courses. This is used to stop
		// the user for selecting/unselecting existing courses
		if (getExistingCourses) {
			const success = await GetAllCourses({})
				.then((courses) => {
					existingCourses.push(...courses.map((c) => c.path));
					getExistingCourses = false;
					return true;
				})
				.catch((err) => {
					const errMsg = ErrorMessage(err);

					gotError = true;
					loadingDrives = false;
					console.error(errMsg);

					$addToast({
						data: {
							message: errMsg,
							status: 'error'
						}
					});

					return false;
				});

			if (!success) return;
		}

		// Pull the filesystem information for this path. When the path is empty it will load
		// drive information
		await GetFileSystem(path)
			.then((resp) => {
				if (body) body.scrollTop = 0;

				if (resp) {
					const selectedKeys = Object.keys(selectedCourses);
					const allPaths = selectedKeys.concat(existingCourses);

					resp.directories?.forEach((d) => {
						// Mark any existing courses
						if (existingCourses.find((c) => c === d.path)) d.isExistingCourse = true;

						// Mark parents. These are courses that contain a selected course
						if (isParent(d.path, allPaths)) d.isParent = true;

						// Mark selected courses
						if (selectedKeys.find((c) => c === d.path)) d.isSelected = true;
					});

					// Set the selectable courses. This will be a list of directories that are not
					// existing courses, selected courses, or parents.
					selectableCourses = resp?.directories.filter(
						({ isSelected, isExistingCourse, isParent }) =>
							!isExistingCourse && !isSelected && !isParent
					);

					pathInfo = resp;

					// Set the loading state to false
					path ? (loadingPath = false) : (loadingDrives = false);

					// When not part of a refresh, manipulate the paths array
					if (!refresh) {
						if (movingBack) {
							// Pop the last path as we are moving up a directory
							paths.pop();
							paths = [...paths];
						} else if (path) {
							// Add the path to the array as we are moving into a directory
							paths = [...paths, path];
						}
					}
				}
			})
			.catch((err) => {
				const errMsg = ErrorMessage(err);
				if (path) {
					console.error(errMsg);
					loadingPath = false;
					// TODO: Go back to previous path
				} else {
					gotError = true;
					loadingDrives = false;
				}

				$addToast({
					data: {
						message: errMsg,
						status: 'error'
					}
				});
			});
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Refreshes the file system. While running, `refreshing` will be true. The function will take
	// at least 1 second to run, to prevent flickering
	const refreshFileSystem = async () => {
		if (refreshing) return;
		refreshing = true;
		const currentPath = paths[paths.length - 1] ?? '';

		// Set a timeout to prevent flickering
		const oneSecondPromise = new Promise((resolve) => setTimeout(resolve, 1000));

		// Wait for load and the oneSecondPromise to resolve
		await Promise.all([load(currentPath, false, true), oneSecondPromise]);

		refreshing = false;
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Returns turn when a path is the parent of another path, as in, this parent path contains a
	// selected course somewhere in its directory structure
	const isParent = (path: string, paths: string[]) => {
		for (let i = 0; i < paths.length; i++) {
			if (paths[i] === path) return false;
			else if (paths[i].startsWith(path)) return true;
		}

		return false;
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Checks if the course is in the selectedCourses array. If it is, it will be removed, making
	// it unselected. If it is not, it will be added, making it selected.
	function FlipSelected(path: string) {
		// If the course is already selected, remove it
		if (path in selectedCourses) {
			delete selectedCourses[path];
		} else {
			selectedCourses[path] = path;
		}

		// Set the course in pathInfo to be selected/unselected
		pathInfo.directories.forEach((d) => {
			if (d.path === path) d.isSelected = !d.isSelected;
		});

		// Update the pathInfo object to trigger a re-render
		pathInfo = { ...pathInfo };
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Select all courses that are currently not selected
	function selectAll() {
		// Loop over all selectable courses and add them to the selected courses array if they
		// are not already selected
		let doReact = false;
		selectableCourses.forEach((course) => {
			if (!(course.path in selectedCourses)) {
				doReact = true;
				selectedCourses[course.path] = course.title;
			}
		});

		// When we added courses to the selected courses array, manually mark these directories as
		// selected
		if (doReact) {
			pathInfo.directories.forEach((d) => {
				if (selectableCourses.find((c) => c.path === d.path)) d.isSelected = true;
			});

			pathInfo = { ...pathInfo };
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Unselect all courses
	function unselectAll() {
		// Loop over all selectable courses and remove them from the selected courses array
		let doReact = false;
		selectableCourses.forEach((course) => {
			if (course.path in selectedCourses) {
				doReact = true;
				delete selectedCourses[course.path];
			}
		});

		// When we removed courses to the selected courses array, manually mark these directories as
		// unselected
		if (doReact) {
			pathInfo.directories.forEach((d) => {
				if (selectableCourses.find((c) => c.path === d.path)) d.isSelected = false;
			});

			pathInfo = { ...pathInfo };

			// Update the selected courses array to trigger a re-render
			selectedCourses = { ...selectedCourses };
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function add() {
		let sawError = false;
		const keys = Object.keys(selectedCourses);
		for (let i = 0; i < keys.length; i++) {
			await AddCourse(selectedCourses[keys[i]], keys[i]).catch((err) => {
				console.error(err);
				sawError = true;
			});
		}

		!sawError &&
			$addToast({
				data: {
					message: `Course${keys.length > 1 ? 's' : ''} added`,
					status: 'success'
				}
			});

		dispatch('added');
	}
</script>

<Sheet.Root bind:open={isOpen}>
	<Sheet.Trigger asChild let:builder>
		<Button builders={[builder]} class="gap-2 rounded px-3 py-1.5">
			<BookPlus class="h-4 w-4" />
			<span>Add Courses</span>
		</Button>
	</Sheet.Trigger>

	<Sheet.Content
		side="right"
		class="flex w-[40em] max-w-full flex-col gap-0 self-end p-0 shadow-none sm:max-w-none [&>button[data-dialog-close]]:hidden"
	>
		<!-- Header -->
		<div
			class="flex h-16 shrink-0 items-center justify-between border-b px-3 text-base font-medium"
		>
			<div class="flex items-center gap-2">
				<BookPlus class="h-4 w-4" />
				<span>Course Selection</span>
			</div>

			<!-- Refresh -->
			<Button
				variant="ghost"
				disabled={isLoadingOrRefreshing}
				class="group px-2.5"
				on:click={refreshFileSystem}
			>
				<RefreshCw
					class={cn(
						'group-hover:text-foreground text-muted-foreground h-5 w-5 duration-200',
						refreshing && 'animate-spin'
					)}
				/>
			</Button>
		</div>

		<!-- Body -->
		<div bind:this={body} class="flex min-h-[15rem] grow flex-col overflow-y-scroll" tabindex="-1">
			{#if loadingDrives || refreshing}
				<div class="flex w-full flex-grow flex-col place-content-center items-center p-10">
					<Loading />
				</div>
			{:else if gotError}
				<Error class="min-h-max" />
			{:else}
				<div class="flex flex-col">
					<!-- Back button -->
					{#if paths.length > 0}
						{#key paths[paths.length - 1]}
							<Button
								variant="ghost"
								disabled={loadingPath}
								class="flex h-14 flex-grow flex-row items-center justify-start rounded-none border-b pr-0"
								on:click={async (el) => {
									// find this buttons child element and show the loader by removing the hidden and adding the flex
									if (!el.target || !(el.target instanceof Element)) return;

									const buttonElement = el.target.closest('button');
									if (!buttonElement) return;

									const loader = buttonElement.querySelector('#back-loader');

									if (loader) {
										loader.classList.remove('hidden');
										loader.classList.add('flex');
									}

									// Determine the new back path. If we are only 1 level
									// deep, load the drives, or else load the path before
									await load(paths.length > 1 ? paths[paths.length - 2] : '', true, false);
								}}
							>
								<div class="flex grow gap-2 text-sm">
									<CornerUpLeft class="text-muted-foreground group-hover:text-foreground h-4 w-4" />
									<span>Back</span>
								</div>

								<div
									class="hidden h-full min-w-20 shrink-0 place-content-center items-center"
									id="back-loader"
								>
									<Loading class="h-5 w-5" />
								</div>
							</Button>
						{/key}
					{/if}

					<!-- Directories -->
					{#each pathInfo.directories as dirInfo (dirInfo.path)}
						<div class="flex h-14 flex-row items-center border-b">
							<!-- Path (left) -->
							<Button
								variant="ghost"
								disabled={loadingPath || dirInfo.isExistingCourse || dirInfo.isSelected}
								class="h-full flex-grow justify-start rounded-none"
								on:click={async () => {
									await load(dirInfo.path, false, false);
								}}
							>
								<span class="flex grow text-sm">{dirInfo.title}</span>

								<!-- Added badge (when this is an existing course) -->
								{#if dirInfo.isExistingCourse}
									<Badge
										variant="outline"
										class="text-muted-foreground/60 border-muted-foreground/40 rounded px-1.5 text-center text-xs"
									>
										Added
									</Badge>
								{/if}
							</Button>

							<!-- Select/unselect (right) -->
							{#if !dirInfo.isExistingCourse}
								<Separator orientation="vertical" class="h-full" />

								<div class="flex h-full min-w-20 shrink-0 place-content-center items-center">
									{#if loadingPath && selectedPath === dirInfo.path}
										<Loading class="h-5 w-5" />
									{:else}
										<Button
											variant="ghost"
											tabindex={dirInfo.isParent ?? false ? -1 : 0}
											disabled={loadingPath || (dirInfo.isParent ?? false)}
											class="hover:bg-background group h-full w-14 shrink-0 place-content-center items-center rounded-none duration-200 disabled:opacity-100 sm:w-20"
											on:click={() => {
												FlipSelected(dirInfo.path);
											}}
										>
											<input
												class="bg-background group-hover:border-muted-foreground checked:bg-primary checked:hover:bg-primary indeterminate:bg-secondary pointer-events-none cursor-pointer rounded border-2 p-2 duration-150 checked:border-transparent indeterminate:opacity-60 group-hover:checked:border-transparent group-hover:checked:brightness-90"
												tabindex="-1"
												checked={(dirInfo.isSelected || dirInfo.isExistingCourse) ?? false}
												type="checkbox"
												indeterminate={dirInfo.isParent ?? false}
											/>
										</Button>
									{/if}
								</div>
							{/if}
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Footer -->
		<Sheet.Footer
			class="flex-row justify-between border-t px-3 py-4 text-sm sm:flex-row sm:justify-between"
		>
			<!-- Select/unselect -->
			<div class="flex gap-3">
				<Button
					variant="outline"
					disabled={isLoadingOrRefreshing}
					class="group h-auto rounded px-2.5 py-1.5 text-sm"
					on:click={selectAll}
				>
					Select All
				</Button>

				<Button
					variant="outline"
					disabled={isLoadingOrRefreshing}
					class="group h-auto rounded px-2.5 py-1.5 text-sm"
					on:click={unselectAll}
				>
					Unselect All
				</Button>
			</div>

			<!-- Close/add -->
			<div class="flex gap-3">
				<Sheet.Close asChild let:builder>
					<Button
						builders={[builder]}
						variant="outline"
						class="group h-auto rounded px-2.5 py-1.5 text-sm"
					>
						Close
					</Button>
				</Sheet.Close>

				<Sheet.Close asChild let:builder>
					<Button
						builders={[builder]}
						disabled={disableAddButton}
						class="group h-auto rounded px-2.5 py-1.5 text-sm"
						on:click={add}
					>
						Add ({Object.keys(selectedCourses).length})
					</Button>
				</Sheet.Close>
			</div>
		</Sheet.Footer>
	</Sheet.Content>
</Sheet.Root>
