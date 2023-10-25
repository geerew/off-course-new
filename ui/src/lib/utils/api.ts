import { PUBLIC_BACKEND } from '$env/static/public';
import { FileSystemSchema, type FileSystem } from '$lib/types/fileSystem';
import {
	CourseSchema,
	ScanSchema,
	type Course,
	type CourseGetParams,
	type CoursePostParams,
	type CoursesGetParams,
	type Scan,
	type ScanPostParams
} from '$lib/types/models';
import { PaginationSchema, type Pagination } from '$lib/types/pagination';
import axios, { AxiosError, type AxiosResponse } from 'axios';
import { safeParse } from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const FS_API =
	process.env.NODE_ENV === 'production' ? '/api/filesystem' : `${PUBLIC_BACKEND}/api/filesystem`;

export const COURSE_API =
	process.env.NODE_ENV === 'production' ? '/api/courses' : `${PUBLIC_BACKEND}/api/courses`;

export const ATTACHMENT_API =
	process.env.NODE_ENV === 'production' ? '/api/courses' : `${PUBLIC_BACKEND}/api/attachments`;

export const SCAN_API =
	process.env.NODE_ENV === 'production' ? '/api/scans' : `${PUBLIC_BACKEND}/api/scans`;
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const ErrorMessage = (error: Error) => {
	return axios.isAxiosError(error) && error.response?.data ? error.response.data : error.message;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// FileSystem
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Gets filesystem information. When the path is empty, the available drives are returned. When
// the path is populated, the directories and files for this path are returned
export const GetFileSystem = async (path?: string): Promise<FileSystem | undefined> => {
	let fsInfo: FileSystem | undefined = undefined;
	let query = FS_API;

	if (path) {
		query += `/${window.btoa(encodeURIComponent(path))}`;
	}

	await axios
		.get(query)
		.then((response: AxiosResponse) => {
			const result = safeParse(FileSystemSchema, response.data);
			if (!result.success) throw new Error('Invalid response from server');
			fsInfo = result.output;
		})
		.catch((error: Error) => {
			throw error;
		});

	return fsInfo;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Courses
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET a paginated list of courses. Use `GetAllCourses()` to get all courses
export const GetCourses = async (params?: CoursesGetParams): Promise<Pagination> => {
	let resp: Pagination | undefined = undefined;

	await axios
		.get(COURSE_API, { params })
		.then((response: AxiosResponse) => {
			const result = safeParse(PaginationSchema, response.data);
			if (!result.success) throw new Error('Invalid response from server');
			resp = result.output;
		})
		.catch((error: Error) => {
			throw error;
		});

	if (!resp) throw new Error('Courses were not found');

	return resp;
};

// GET all courses. This calls GetCourses(...) until all courses are returned
export const GetAllCourses = async (params?: CoursesGetParams): Promise<Course[]> => {
	let resp: Course[] = [];

	let page = 1;
	let getMore = true;

	while (getMore) {
		await GetCourses({ orderBy: params?.orderBy, page: page, perPage: 100 })
			.then((data) => {
				if (data && data.totalItems > 0) {
					const result = safeParse(PaginationSchema, data);
					if (!result.success) throw new Error('Invalid response');

					resp ? (resp = [...resp, ...result.output.items]) : (resp = result.output.items);

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
				throw error;
			});
	}

	return resp;
};

// GET a course by ID
export const GetCourse = async (id: string, params?: CourseGetParams): Promise<Course> => {
	let course: Course | undefined = undefined;

	await axios
		.get(`${COURSE_API}/${id}`, { params })
		.then((response: AxiosResponse) => {
			const result = safeParse(CourseSchema, response.data);
			if (!result.success) throw new Error('Invalid response from server');
			course = result.output;
		})
		.catch((error: Error) => {
			throw error;
		});

	if (!course) throw new Error('Course was not found');

	return course;
};

// POST a course. The data object needs to contain a `title` and `path`
export const AddCourse = async (data: CoursePostParams): Promise<Course> => {
	let course: Course | undefined = undefined;

	await axios
		.post(COURSE_API, data, {
			headers: {
				'content-type': 'application/json'
			}
		})
		.then((response: AxiosResponse) => {
			const result = safeParse(CourseSchema, response.data);
			if (!result.success) throw new Error('Invalid response from server');
			course = result.output;
		})
		.catch((error: Error) => {
			throw error;
		});

	if (!course) throw new Error('Course was not created');

	return course;
};

// DELETE a course
export const DeleteCourse = async (id: string): Promise<boolean> => {
	await axios.delete(`${COURSE_API}/${id}`).catch((error: Error) => {
		throw error;
	});

	return true;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Attachments
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DOWNLOAD an attachment
export const DownloadAttachment = async (id: string): Promise<boolean> => {
	await axios
		.get(`${ATTACHMENT_API}/${id}/download`)
		.then((response: AxiosResponse) => {
			console.log(response);
			return true;
		})
		.catch((error: Error) => {
			throw error;
		});

	return false;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Scans
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET a scan by course ID
export const GetScanByCourseId = async (id: string): Promise<Scan> => {
	let scan: Scan | undefined = undefined;

	await axios
		.get(`${SCAN_API}/course/${id}`)
		.then((response: AxiosResponse) => {
			const result = safeParse(ScanSchema, response.data);
			if (!result.success) throw new Error('Invalid response from server');
			scan = result.output;
		})
		.catch((error: AxiosError) => {
			throw error;
		});

	if (!scan) throw new Error('Scan was not found');

	return scan;
};

// POST a scan. The data object needs to contain an `courseId`
export const AddScan = async (data: ScanPostParams): Promise<Scan> => {
	let scan: Scan | undefined = undefined;

	await axios
		.post(SCAN_API, data, {
			headers: {
				'content-type': 'application/json'
			}
		})
		.then((response: AxiosResponse) => {
			const result = safeParse(ScanSchema, response.data);
			if (!result.success) throw Error('Invalid response from server');
			scan = result.output;
		})
		.catch((error: Error) => {
			throw error;
		});

	if (!scan) throw new Error('Scan was not started');

	return scan;
};
