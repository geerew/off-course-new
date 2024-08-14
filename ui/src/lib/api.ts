import { FileSystemSchema, type FileSystem } from '$lib/types/fileSystem';
import {
	AssetSchema,
	CourseSchema,
	CourseTagSchema,
	ScanSchema,
	TagSchema,
	type Asset,
	type AssetsGetParams,
	type Course,
	type CourseTag,
	type CoursesGetParams,
	type LogsGetParams,
	type Scan,
	type Tag,
	type TagGetParams,
	type TagsGetParams
} from '$lib/types/models';
import { PaginationSchema, type Pagination } from '$lib/types/pagination';
import axios from 'axios';
import { array, safeParse, string } from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const isProduction = process.env.NODE_ENV === 'production';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const FS_API = '/api/filesystem';
export const COURSE_API = '/api/courses';
export const ASSET_API = '/api/assets';
export const ATTACHMENT_API = '/api/attachments';
export const TAGS_API = '/api/tags';
export const SCAN_API = '/api/scans';
export const LOG_API = '/api/logs';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Determine custom backend port. Default to 9081
const backendPort = import.meta.env.BACKEND_PORT || '9081';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get the backend URL based on whether the app is in production or development
export function GetBackendUrl(api: string) {
	if (isProduction) {
		return api;
	} else {
		return `http://localhost:${backendPort}${api}`;
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// FileSystem
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get information about a directory
//
// When the path is empty, the available drives are returned. When the path is populated, the
// directories and files for this path are returned
export async function GetFileSystem(path?: string): Promise<FileSystem> {
	try {
		let query = GetBackendUrl(FS_API);
		if (path) query += `/${window.btoa(encodeURIComponent(path))}`;

		const response = await axios.get<FileSystem>(query);
		const result = safeParse(FileSystemSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to retrieve file system: ${error}`);
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Courses
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get a paginated list of courses. Use `GetAllCourses()` to get an unpaginated list of
// courses
export async function GetCourses(params?: CoursesGetParams): Promise<Pagination> {
	try {
		const response = await axios.get<Pagination>(GetBackendUrl(COURSE_API), { params });
		const result = safeParse(PaginationSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to retrieve courses: ${error}`);
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get all courses (not paginated). This calls GetCourses(...) until all courses are
// are returned
export async function GetAllCourses(params?: CoursesGetParams): Promise<Course[]> {
	let allCourses: Course[] = [];
	let page = 1;
	let totalPages = 1;

	do {
		try {
			const data = await GetCourses({ ...params, page, perPage: 100 });

			if (data.totalItems > 0) {
				allCourses = [...allCourses, ...(data.items as Course[])];
				totalPages = data.totalPages;
				page++;
			} else {
				break;
			}
		} catch (error) {
			if (axios.isAxiosError(error)) {
				throw error;
			} else {
				throw new Error(`Failed to fetch all courses: ${error}`);
			}
		}
	} while (page <= totalPages);

	return allCourses;
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get a course by ID
export async function GetCourse(id: string): Promise<Course> {
	try {
		const response = await axios.get<Course>(`${GetBackendUrl(COURSE_API)}/${id}`);
		const result = safeParse(CourseSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to retrieve course: ${error}`);
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Gets the course id from the search params and queries the api for the course
export async function GetCourseFromParams(params: URLSearchParams): Promise<Course> {
	const id = params && params.get('id');
	if (!id) throw new Error('Missing course ID');

	return GetCourse(id);
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// POST - Create a course
export async function AddCourse(title: string, path: string): Promise<Course> {
	try {
		const response = await axios.post<Course>(
			GetBackendUrl(COURSE_API),
			{ title, path },
			{
				headers: {
					'content-type': 'application/json'
				}
			}
		);
		const result = safeParse(CourseSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to create course: ${error}`);
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// PUT - Update a course
export async function UpdateCourse(course: Course): Promise<Course> {
	try {
		const response = await axios.put<Course>(`${GetBackendUrl(COURSE_API)}/${course.id}`, course, {
			headers: {
				'content-type': 'application/json'
			}
		});
		const result = safeParse(CourseSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to update course: ${error}`);
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DELETE - Delete a course
export async function DeleteCourse(id: string): Promise<boolean> {
	try {
		await axios.delete(`${GetBackendUrl(COURSE_API)}/${id}`);
		return true;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to delete course: ${error}`);
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Course Tags
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get a list of tags for a course
export async function GetCourseTags(courseId: string): Promise<CourseTag[]> {
	try {
		const response = await axios.get<CourseTag[]>(`${GetBackendUrl(COURSE_API)}/${courseId}/tags`);
		const result = safeParse(array(CourseTagSchema), response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to retrieve course tags: ${error}`);
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// POST - Add a tag to a course. The tag will be created if it does not exist
export async function AddCourseTag(courseId: string, tag: string): Promise<CourseTag> {
	try {
		const response = await axios.post<CourseTag>(
			`${GetBackendUrl(COURSE_API)}/${courseId}/tags/`,
			{ tag },
			{
				headers: {
					'content-type': 'application/json'
				}
			}
		);
		const result = safeParse(CourseTagSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to add course tag: ${error}`);
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DELETE - Delete a course tag
export async function DeleteCourseTag(courseId: string, tagId: string): Promise<boolean> {
	try {
		await axios.delete(`${GetBackendUrl(COURSE_API)}/${courseId}/tags/${tagId}`);
		return true;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to delete course tag: ${error}`);
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Assets
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get a paginated list of assets for a course. Use `GetAllCourseAssets()` to get an
// unpaginated list of assets for a course
//
// Requires a course ID
export async function GetCourseAssets(
	courseId: string,
	params?: AssetsGetParams
): Promise<Pagination> {
	try {
		const response = await axios.get<Pagination>(
			`${GetBackendUrl(COURSE_API)}/${courseId}/assets`,
			{ params }
		);
		const result = safeParse(PaginationSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to get course assets: ${error}`);
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get all assets (not paginated) for a course. This calls GetAssets(...) until all assets
// are returned
export async function GetAllCourseAssets(
	courseId: string,
	params?: AssetsGetParams
): Promise<Asset[]> {
	let allAssets: Asset[] = [];
	let page = 1;
	let totalPages = 1;

	do {
		try {
			const response = await GetCourseAssets(courseId, { ...params, page, perPage: 100 });

			if (response.totalItems > 0) {
				const result = safeParse(PaginationSchema, response);
				if (!result.success) throw new Error('Invalid response');

				allAssets = [...allAssets, ...(response.items as Asset[])];
				totalPages = response.totalPages;
				page++;
			} else {
				break;
			}
		} catch (error) {
			if (axios.isAxiosError(error)) {
				throw error;
			} else {
				throw new Error(`Failed to get all course assets: ${error}`);
			}
		}
	} while (page <= totalPages);

	return allAssets;
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// PUT - Update an asset
export async function UpdateAsset(asset: Asset): Promise<Asset> {
	try {
		const response = await axios.put<Asset>(`${GetBackendUrl(ASSET_API)}/${asset.id}`, asset, {
			headers: {
				'content-type': 'application/json'
			}
		});

		const parseResult = safeParse(AssetSchema, response.data);

		if (!parseResult.success) throw new Error('Invalid response from server');
		return parseResult.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to update asset: ${error}`);
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Scans
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get a scan by course ID
export async function GetScan(courseId: string): Promise<Scan> {
	try {
		const response = await axios.get<Scan>(`${GetBackendUrl(SCAN_API)}/${courseId}`);
		const result = safeParse(ScanSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to get scan: ${error}`);
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// POST - Create a scan for a course
export async function AddScan(courseId: string): Promise<Scan> {
	try {
		const response = await axios.post<Scan>(
			GetBackendUrl(SCAN_API),
			{ courseId },
			{
				headers: {
					'content-type': 'application/json'
				}
			}
		);

		const result = safeParse(ScanSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to add scan job: ${error}`);
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Tags
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get a tag by ID or name
export async function GetTag(idOrName: string, params?: TagGetParams): Promise<Tag> {
	try {
		const response = await axios.get<Tag>(`${GetBackendUrl(TAGS_API)}/${idOrName}`, { params });
		const result = safeParse(TagSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to get tag: ${error}`);
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get a paginated list of tags
export async function GetTags(params?: TagsGetParams): Promise<Pagination> {
	try {
		const response = await axios.get<Pagination>(GetBackendUrl(TAGS_API), { params });
		const result = safeParse(PaginationSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to retrieve tags: ${error}`);
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get all tags (not paginated). This calls GetTags(...) until all tags are
// are returned
export async function GetAllTags(params?: CoursesGetParams): Promise<Tag[]> {
	let allTags: Tag[] = [];
	let page = 1;
	let totalPages = 1;

	do {
		try {
			const data = await GetTags({ ...params, page, perPage: 100 });

			if (data.totalItems > 0) {
				allTags = [...allTags, ...(data.items as Tag[])];
				totalPages = data.totalPages;
				page++;
			} else {
				break;
			}
		} catch (error) {
			if (axios.isAxiosError(error)) {
				throw error;
			} else {
				throw new Error(`Failed to fetch all tags: ${error}`);
			}
		}
	} while (page <= totalPages);

	return allTags;
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// POST - Create a tag
export async function AddTag(tag: string): Promise<Tag> {
	try {
		const response = await axios.post<Tag>(
			GetBackendUrl(TAGS_API),
			{ tag },
			{
				headers: {
					'content-type': 'application/json'
				}
			}
		);

		const result = safeParse(TagSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to add tag: ${error}`);
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// PUT - Update a tag
export async function UpdateTag(tag: Tag): Promise<Tag> {
	try {
		const response = await axios.put<Tag>(`${GetBackendUrl(TAGS_API)}/${tag.id}`, tag, {
			headers: {
				'content-type': 'application/json'
			}
		});
		const result = safeParse(TagSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to update tag: ${error}`);
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DELETE - Delete a tag
export async function DeleteTag(tagId: string): Promise<boolean> {
	try {
		await axios.delete(`${GetBackendUrl(TAGS_API)}/${tagId}`);
		return true;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to delete course tag: ${error}`);
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Logs
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get a paginated list of logs
export async function GetLogs(params?: LogsGetParams): Promise<Pagination> {
	try {
		const response = await axios.get<Pagination>(GetBackendUrl(LOG_API), { params });
		const result = safeParse(PaginationSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to retrieve logs: ${error}`);
		}
	}
}

// GET - Get a list of log types
export async function GetLogTypes(): Promise<string[]> {
	try {
		const response = await axios.get<string[]>(`${GetBackendUrl(LOG_API)}/types`);
		const result = safeParse(array(string()), response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to retrieve log types: ${error}`);
		}
	}
}
