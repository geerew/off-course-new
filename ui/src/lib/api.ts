import { PUBLIC_BACKEND } from '$env/static/public';
import { FileSystemSchema, type FileSystem } from '$lib/types/fileSystem';
import {
	AssetSchema,
	CourseSchema,
	ScanSchema,
	type Asset,
	type AssetsGetParams,
	type Course,
	type CoursesGetParams,
	type Scan
} from '$lib/types/models';
import { PaginationSchema, type Pagination } from '$lib/types/pagination';
import axios, { AxiosError, type AxiosResponse } from 'axios';
import { safeParse } from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const FS_API =
	process.env.NODE_ENV === 'production' ? '/api/filesystem' : `${PUBLIC_BACKEND}/api/filesystem`;

export const COURSE_API =
	process.env.NODE_ENV === 'production' ? '/api/courses' : `${PUBLIC_BACKEND}/api/courses`;

export const ASSET_API =
	process.env.NODE_ENV === 'production' ? '/api/assets' : `${PUBLIC_BACKEND}/api/assets`;

export const ATTACHMENT_API =
	process.env.NODE_ENV === 'production' ? '/api/attachments' : `${PUBLIC_BACKEND}/api/attachments`;

export const SCAN_API =
	process.env.NODE_ENV === 'production' ? '/api/scans' : `${PUBLIC_BACKEND}/api/scans`;
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const ErrorMessage = (error: Error) => {
	return axios.isAxiosError(error) && error.response?.data ? error.response.data : error.message;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// FileSystem
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get information about a directory
//
// When the path is empty, the available drives are returned. When the path is populated, the
// directories and files for this path are returned
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

// GET - Get a paginated list of courses. Use `GetAllCourses()` to get an unpaginated list of
// courses
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get all courses (not paginated). This calls GetCourses(...) until all courses are
// are returned
export const GetAllCourses = async (params?: CoursesGetParams): Promise<Course[]> => {
	let resp: Course[] = [];

	let page = 1;
	let getMore = true;

	while (getMore) {
		await GetCourses({ ...params, page: page, perPage: 100 })
			.then((data) => {
				if (data && data.totalItems > 0) {
					const result = safeParse(PaginationSchema, data);
					if (!result.success) throw new Error('Invalid response');

					resp = resp.concat(result.output.items as Course[]);

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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get a course by ID
export const GetCourse = async (id: string): Promise<Course> => {
	let course: Course | undefined = undefined;

	await axios
		.get(`${COURSE_API}/${id}`)
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// POST - Create a course
export const AddCourse = async (title: string, path: string): Promise<Course> => {
	let course: Course | undefined = undefined;

	await axios
		.post(
			COURSE_API,
			{ title, path },
			{
				headers: {
					'content-type': 'application/json'
				}
			}
		)
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// PUT - Update a course
export const UpdateCourse = async (course: Course): Promise<boolean> => {
	const res = await axios
		.put(`${COURSE_API}/${course.id}`, course, {
			headers: {
				'content-type': 'application/json'
			}
		})
		.then((response: AxiosResponse) => {
			if (!safeParse(CourseSchema, response.data).success)
				throw new Error('Invalid response from server');
			return true;
		})
		.catch((error: Error) => {
			throw error;
		});

	return res;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DELETE - Delete a course
export const DeleteCourse = async (id: string): Promise<boolean> => {
	await axios.delete(`${COURSE_API}/${id}`).catch((error: Error) => {
		throw error;
	});

	return true;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Assets
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get a paginated list of assets for a course. Use `GetAllCourseAssets()` to get an
// unpaginated list of assets for a course
//
// Requires a course ID
export const GetCourseAssets = async (
	courseId: string,
	params?: AssetsGetParams
): Promise<Pagination> => {
	let resp: Pagination | undefined = undefined;

	await axios
		.get(`${COURSE_API}/${courseId}/assets`, { params })
		.then((response: AxiosResponse) => {
			const result = safeParse(PaginationSchema, response.data);
			if (!result.success) throw new Error('Invalid response from server');
			resp = result.output;
		})
		.catch((error: Error) => {
			throw error;
		});

	if (!resp) throw new Error('Assets were not found');

	return resp;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get all assets (not paginated) for a course. This calls GetAssets(...) until all assets
// are returned
//
// Requires a course ID
export const GetAllCourseAssets = async (
	courseId: string,
	params?: AssetsGetParams
): Promise<Asset[]> => {
	let resp: Asset[] = [];

	let page = 1;
	let getMore = true;

	while (getMore) {
		await GetCourseAssets(courseId, { ...params, page: page, perPage: 100 })
			.then((data) => {
				if (data && data.totalItems > 0) {
					const result = safeParse(PaginationSchema, data);
					if (!result.success) throw new Error('Invalid response');

					resp = resp.concat(result.output.items as Asset[]);

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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// PUT - Update an asset
export const UpdateAsset = async (asset: Asset): Promise<boolean> => {
	const res = await axios
		.put(`${ASSET_API}/${asset.id}`, asset, {
			headers: {
				'content-type': 'application/json'
			}
		})
		.then((response: AxiosResponse) => {
			if (!safeParse(AssetSchema, response.data).success)
				throw new Error('Invalid response from server');
			return true;
		})
		.catch((error: Error) => {
			throw error;
		});

	return res;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Scans
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get a scan by course ID
//
// Requires a course ID
export const GetScan = async (courseId: string): Promise<Scan> => {
	let scan: Scan | undefined = undefined;

	await axios
		.get(`${SCAN_API}/${courseId}`)
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// POST - Create a scan for a course
//
// Requires a course ID
export const AddScan = async (courseId: string): Promise<Scan> => {
	let scan: Scan | undefined = undefined;

	await axios
		.post(
			SCAN_API,
			{ courseId },
			{
				headers: {
					'content-type': 'application/json'
				}
			}
		)
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
