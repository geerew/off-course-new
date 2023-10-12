import { PUBLIC_BACKEND } from '$env/static/public';
import { addToast } from '$lib/stores/addToast';
import type { FileSystemInfo } from '$lib/types/fileSystem';
import type { Course, CourseGet, CoursePost, CoursesGet, Scan, ScanPost } from '$lib/types/models';
import type { PaginationResponse } from '$lib/types/pagination';
import axios, { AxiosError, type AxiosResponse } from 'axios';
import { get } from 'svelte/store';
import { isBrowser } from './general';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const FS_API =
	process.env.NODE_ENV === 'production' ? '/api/filesystem' : `${PUBLIC_BACKEND}/api/filesystem`;

export const COURSE_API =
	process.env.NODE_ENV === 'production' ? '/api/courses' : `${PUBLIC_BACKEND}/api/courses`;

export const SCAN_API =
	process.env.NODE_ENV === 'production' ? '/api/scans' : `${PUBLIC_BACKEND}/api/scans`;
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const ErrorMessage = (error: Error) => {
	return axios.isAxiosError(error) && error.response?.data ? error.response.data : error.message;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// FileSystem
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Gets filesystem information. When the path is empty, the available drives are returned. When
// the path is populated, the directories and files for this path are returned
export const GetFileSystem = async (
	path?: string,
	muteToast = false
): Promise<FileSystemInfo | undefined> => {
	let fsInfo: FileSystemInfo | undefined = undefined;
	let query = FS_API;

	if (path) {
		query += `/${window.btoa(encodeURIComponent(path))}`;
	}

	await axios
		.get(query)
		.then((response: AxiosResponse) => {
			fsInfo = response.data as FileSystemInfo;
		})
		.catch((error: Error) => {
			!muteToast &&
				get(addToast)({
					data: { message: `Failed to lookup path`, status: 'error' }
				});

			throw ErrorMessage(error);
		});

	return fsInfo;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Courses
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET a paginated list of courses. Use `GetAllCourses()` to get all courses
export const GetCourses = async (
	params?: CoursesGet,
	muteToast = false
): Promise<PaginationResponse | undefined> => {
	let resp: PaginationResponse | undefined = undefined;

	await axios
		.get(COURSE_API, { params })
		.then((response: AxiosResponse) => {
			resp = response.data as PaginationResponse;
		})
		.catch((error: Error) => {
			!muteToast &&
				get(addToast)({
					data: {
						message: `Failed to get courses`,
						status: 'error'
					}
				});

			throw ErrorMessage(error);
		});

	return resp;
};

// GET all courses. This calls GetCourses(...) until all courses are returned
export const GetAllCourses = async (params?: CoursesGet, muteToast = false): Promise<Course[]> => {
	let resp: Course[] = [];

	let page = 1;
	let getMore = true;

	while (getMore) {
		await GetCourses({ orderBy: params?.orderBy, page: page, perPage: 100 }, muteToast)
			.then((data) => {
				if (data && data.totalItems > 0) {
					resp ? (resp = [...resp, ...(data.items as Course[])]) : (resp = data.items as Course[]);

					if (data.page !== data.totalPages) {
						page++;
					} else {
						getMore = false;
					}
				} else {
					getMore = false;
				}
			})
			.catch((error) => {
				!muteToast &&
					get(addToast)({
						data: { message: `Failed to get all courses`, status: 'error' }
					});

				throw ErrorMessage(error);
			});
	}

	return resp;
};

// GET a course by ID
export const GetCourse = async (
	id: string,
	params?: CourseGet,
	muteToast = false
): Promise<Course | undefined> => {
	let resp: PaginationResponse | undefined = undefined;

	await axios
		.get(`${COURSE_API}/${id}`, { params })
		.then((response: AxiosResponse) => {
			resp = response.data as PaginationResponse;
		})
		.catch((error: Error) => {
			if (!muteToast && isBrowser) {
				get(addToast)({
					data: {
						message: `Failed to get course`,
						status: 'error'
					}
				});
			}

			throw ErrorMessage(error);
		});

	return resp;
};

// POST a course. The data object needs to contain a `title` and `path`
export const AddCourse = async (data: CoursePost, muteToast = false): Promise<Course> => {
	let course: Course | undefined = undefined;

	await axios
		.post(COURSE_API, data, {
			headers: {
				'content-type': 'application/json'
			}
		})
		.then((response: AxiosResponse) => {
			course = response.data as Course;
		})
		.catch((error: Error) => {
			!muteToast &&
				get(addToast)({
					data: {
						message: `Failed to add course <span class="font-mono text-error text-sm">${data.title}</span>`,
						status: 'error'
					}
				});

			throw ErrorMessage(error);
		});

	if (!course) throw new Error('Course was not created');

	return course;
};

// DELETE a course
export const DeleteCourse = async (id: string, muteToast = false): Promise<boolean> => {
	await axios.delete(`${COURSE_API}/${id}`).catch((error: Error) => {
		!muteToast &&
			get(addToast)({
				data: {
					message: `Failed to delete course`,
					status: 'error'
				}
			});

		throw ErrorMessage(error);
	});

	return true;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Scans
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET a scan by course ID
export const GetScanByCourseId = async (
	id: string,
	muteToast = false
): Promise<Scan | undefined> => {
	let scan: Scan | undefined = undefined;

	await axios
		.get(`${SCAN_API}/course/${id}`)
		.then((response: AxiosResponse) => {
			scan = response.data as Scan;
		})
		.catch((error: AxiosError) => {
			!muteToast &&
				get(addToast)({
					data: {
						message: `Failed to get scan`,
						status: 'error'
					}
				});

			throw error;
		});

	return scan;
};

// POST a scan. The data object needs to contain an `courseId`
export const AddScan = async (data: ScanPost, muteToast = false): Promise<Scan> => {
	let scan: Scan | undefined = undefined;

	await axios
		.post(SCAN_API, data, {
			headers: {
				'content-type': 'application/json'
			}
		})
		.then((response: AxiosResponse) => {
			scan = response.data as Scan;
		})
		.catch((error: Error) => {
			!muteToast &&
				get(addToast)({
					data: {
						message: `Failed to start a scan`,
						status: 'error'
					}
				});

			throw ErrorMessage(error);
		});

	if (!scan) throw new Error('Scan was not started');

	return scan;
};
